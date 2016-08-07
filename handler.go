package main

import (
	"log"
	"time"
)

func PlayerConnectHandler(context *DemoContext, demoStatistic *DemoStatistic, event *DemoGameEvent) {
	steamId := event.Data["networkid"].(string)
	player, found := demoStatistic.MatchInfo.findPlayerBySteamId(steamId)

	if found {
		player.Connected = true
	} else {
		player := &Player{Name: event.Data["name"].(string),
			UserId:    event.Data["userid"].(int32),
			SteamId:   steamId,
			Connected: true,
		}
		demoStatistic.MatchInfo.AddPlayer(player)
	}
}

func PlayerDisconnectHandler(context *DemoContext, demoStatistic *DemoStatistic, event *DemoGameEvent) {
	player, found := demoStatistic.MatchInfo.findPlayerBySteamId(event.Data["networkid"].(string))

	if !found {
		log.Println("Received player_disconnect event, but could not find the player!?")
		return
	}

	if player.SteamId == "BOT" {
		demoStatistic.MatchInfo.RemovePlayer(player)
	} else {
		player.Connected = false
	}
}

func PlayerTeamHandler(context *DemoContext, demoStatistic *DemoStatistic, event *DemoGameEvent) {
	player, found := demoStatistic.MatchInfo.findPlayerByUserId(event.Data["userid"].(int32))

	if found && player.Team == 0 {
		player.Team = int(event.Data["team"].(int32))
	}

}

func RoundAnnounceMatchStartHandler(context *DemoContext, demoStatistic *DemoStatistic, event *DemoGameEvent) {
	demoStatistic.MatchStartTick = event.Tick
}

func RoundStartHandler(context *DemoContext, demoStatistic *DemoStatistic, event *DemoGameEvent) {
	demoStatistic.AddNewRound(event.Tick)
}

func RoundEndHandler(context *DemoContext, demoStatistic *DemoStatistic, event *DemoGameEvent) {
	round, found := demoStatistic.FindRoundByTick(event.Tick)

	if !found {
		log.Println("Could not handle round end")
		return
	}
	round.Winner = int(event.Data["winner"].(int32))
	round.Reason = int(event.Data["reason"].(int32))
}

func RoundOfficiallyEndHandler(context *DemoContext, demoStatistic *DemoStatistic, event *DemoGameEvent) {
	round, found := demoStatistic.FindRoundByTick(event.Tick)

	if !found {
		log.Println("Could not handle round end")
		return
	}
	round.EndTick = event.Tick
	round.Duration = time.Duration(event.Tick-round.StartTick) * context.tickDuration
}

func CollectHandler(context *DemoContext, demoStatistic *DemoStatistic, event *DemoGameEvent) {
	round, found := demoStatistic.FindRoundByTick(event.Tick)
	if found {
		event.TimeInRound = time.Duration(event.Tick-round.StartTick) * context.tickDuration
		round.AddEvent(event)
	} else {
		demoStatistic.UnmappedEvents = append(demoStatistic.UnmappedEvents, event)
	}
}

func CalculateMatchDurationHandler(context *DemoContext, demoStatistic *DemoStatistic, event *DemoGameEvent) {
	demoStatistic.MatchInfo.Duration = time.Duration(event.Tick) * context.tickDuration
}
