package demoinfo

import (
	"bytes"
	"fmt"
	"github.com/stegmannc/csgo-demoparser/protom"
	"sync"
	"time"
)

type Player struct {
	SteamId   string `json:"steamId"`
	Name      string `json:"name"`
	UserId    int32  `json:"userId"`
	Team      int    `json:"team"`
	Connected bool   `json:"-"`
}

type DemoGameEvent struct {
	EventId     int32                  `json:"eventId"`
	Name        string                 `json:"name"`
	Tick        int32                  `json:"-"`
	TimeInRound time.Duration          `json:"timeInRound"`
	Data        map[string]interface{} `json:"data"`
}

func (event *DemoGameEvent) addData(name string, value interface{}) {
	event.Data[name] = value
}

func NewDemoGameEvent(eventid int32, name string, tick int32) *DemoGameEvent {
	return &DemoGameEvent{EventId: eventid, Name: name, Tick: tick, Data: make(map[string]interface{})}
}

type DemoContext struct {
	GameEventList *protom.CSVCMsg_GameEventList
	GameEventChan chan *DemoGameEvent
	StopChan      chan bool
	tickDuration  time.Duration
}

func (dc *DemoContext) GetGameEventDescriptor(eventid int32) *protom.CSVCMsg_GameEventListDescriptorT {
	descriptors := dc.GameEventList.GetDescriptors()
	for _, descriptor := range descriptors {
		if *descriptor.Eventid == eventid {
			return descriptor
		}
	}
	return nil
}

func NewDemoContext(header *DemoHeader) *DemoContext {
	formattedTickDuration := fmt.Sprintf("%gs", (header.PlaybackTime / float32(header.PlaybackTicks)))
	finalTickDuration, err := time.ParseDuration(formattedTickDuration)

	if err != nil {
		panic(err)
	}

	return &DemoContext{
		GameEventChan: make(chan *DemoGameEvent),
		StopChan:      make(chan bool),
		tickDuration:  finalTickDuration,
	}
}

type MatchInfo struct {
	mu         sync.Mutex
	Map        string        `json:"map"`
	ServerName string        `json:"serverName"`
	Players    []*Player     `json:"players"`
	Duration   time.Duration `json:"duration"`
}

func (mi *MatchInfo) AddPlayer(player *Player) {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	_, found := mi.findPlayerBySteamId(player.SteamId)
	if !found {
		mi.Players = append(mi.Players, player)
	}
}

func (mi *MatchInfo) RemovePlayer(playerToDelete *Player) bool {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	for i, player := range mi.Players {
		if player == playerToDelete {
			copy(mi.Players[i:], mi.Players[i+1:])
			mi.Players[len(mi.Players)-1] = nil
			mi.Players = mi.Players[:len(mi.Players)-1]
			return true
		}
	}
	return false

}

func (mi *MatchInfo) findPlayerBySteamId(steamId string) (*Player, bool) {
	for _, player := range mi.Players {
		if player.SteamId == steamId {
			return player, true
		}
	}
	return nil, false
}

func (mi *MatchInfo) findPlayerByUserId(userId int32) (*Player, bool) {
	for _, player := range mi.Players {
		if player.UserId == userId {
			return player, true
		}
	}
	return nil, false
}

type Round struct {
	mu        sync.Mutex
	Nr        int              `json:"nr"`
	Duration  time.Duration    `json:"duration"`
	Winner    int              `json:"winner"`
	Reason    int              `json:"reason"`
	Events    []*DemoGameEvent `json:"events"`
	StartTick int32            `json:"-"`
	EndTick   int32            `json:"-"`
}

func NewRound(startTick int32) *Round {
	return &Round{
		StartTick: startTick,
		Events:    make([]*DemoGameEvent, 0, 10),
	}
}

func (r *Round) AddEvent(event *DemoGameEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Events = append(r.Events, event)
}

type DemoStatistic struct {
	mu             sync.Mutex
	MatchInfo      *MatchInfo       `json:"matchInfo"`
	MatchStartTick int32            `json:"-"`
	Rounds         []*Round         `json:"rounds"`
	UnmappedEvents []*DemoGameEvent `json:"unmappedEvents"`
}

func (ds *DemoStatistic) AddNewRound(startTick int32) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.Rounds = append(ds.Rounds, NewRound(startTick))
}

func NewDemoStatistic(demoHeader *DemoHeader) *DemoStatistic {

	matchInfo := &MatchInfo{
		Map:        convertToString(demoHeader.Mapname[:]),
		ServerName: convertToString(demoHeader.Servername[:]),
	}
	return &DemoStatistic{
		MatchInfo:      matchInfo,
		Rounds:         make([]*Round, 0, 30),
		UnmappedEvents: make([]*DemoGameEvent, 0),
	}
}

func convertToString(data []byte) string {
	n := bytes.IndexByte(data[:], 0)
	return string(data[:n])
}

func (ds *DemoStatistic) FindRoundByTick(tick int32) (round *Round, found bool) {
	round = nil
	found = false

	for _, r := range ds.Rounds {
		if tick >= r.StartTick && (r.EndTick == 0 || tick < r.EndTick) {
			round = r
			found = true
		}
	}

	return
}

func (ds *DemoStatistic) RemoveWarmupRounds() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	var roundIndex int
	for i, round := range ds.Rounds {
		if ds.MatchStartTick >= round.StartTick && (round.EndTick == 0 || ds.MatchStartTick < round.EndTick) {
			roundIndex = i
		}
	}

	copy(ds.Rounds[roundIndex:], ds.Rounds[roundIndex:])
	sizeToRemove := len(ds.Rounds) - (len(ds.Rounds) - roundIndex)
	for k, n := 0, sizeToRemove; k < n; k++ {
		ds.Rounds[k] = nil
	}
	ds.Rounds = ds.Rounds[sizeToRemove:len(ds.Rounds)]
}

func (ds *DemoStatistic) RenumberRounds() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	for i, round := range ds.Rounds {
		round.Nr = i + 1
	}
}
