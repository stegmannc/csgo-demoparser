package main

import (
	"flag"
	"log"
	"github.com/stegmannc/csgo-demoparser"
	"time"
)

var filename = flag.String("f", "", "Path to *.dem file")

func main() {
	flag.Parse()

	start := time.Now()
	parser, err := demoinfo.NewDemoParser(*filename)
	if err != nil {
		log.Fatal(err)
	}

	go parser.ParseTicks()
	statistics := aggregateEvents(parser.Context, parser.Header)
	demoinfo.PrintAsJson(statistics)
	elapsed := time.Since(start)
	log.Printf("Demo parsing took %s", elapsed)
}

func aggregateEvents(context *demoinfo.DemoContext, demoHeader *demoinfo.DemoHeader) *demoinfo.DemoStatistic {
	statistic := demoinfo.NewDemoStatistic(demoHeader)
	dispatcher := demoinfo.NewDemoEventDispatcher(statistic, context)
	dispatcher.RegisterHandler("player_connect", demoinfo.PlayerConnectHandler)
	dispatcher.RegisterHandler("player_disconnect", demoinfo.PlayerDisconnectHandler)
	dispatcher.RegisterHandler("player_team", demoinfo.PlayerTeamHandler)
	dispatcher.RegisterHandler("round_announce_match_start", demoinfo.RoundAnnounceMatchStartHandler)
	dispatcher.RegisterHandler("round_start", demoinfo.RoundStartHandler)
	dispatcher.RegisterHandler("round_end", demoinfo.RoundEndHandler)
	dispatcher.RegisterHandler("round_officially_ended", demoinfo.RoundOfficiallyEndHandler)
	dispatcher.RegisterHandler("cs_win_panel_match", demoinfo.CalculateMatchDurationHandler)

	dispatcher.RegisterHandler("player_hurt", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("player_death", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("item_purchase", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("bomb_beginplant", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("bomb_abortplant", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("bomb_planted", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("bomb_defused", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("bomb_exploded", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("bomb_dropped", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("bomb_pickup", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("defuser_pickup", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("bomb_begindefuse", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("bomb_abortdefuse", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("weapon_fire", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("silencer_detach", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("inspect_weapon", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("enter_bombzone", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("exit_bombzone", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("grenade_bounce", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("flashbang_detonate", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("smokegrenade_detonate", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("smokegrenade_expired", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("molotov_detonate", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("decoy_detonate", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("decoy_started", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("tagrenade_detonate", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("inferno_startburn", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("inferno_expire", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("inferno_extinguish", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("bullet_impact", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("player_blind", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("player_falldamage", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("door_moving", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("player_avenged_teammate", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("match_end_conditions", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("round_mvp", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("player_given_c4", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("tr_player_flashbanged", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("bot_takeover", demoinfo.CollectHandler)
	dispatcher.RegisterHandler("door_moving", demoinfo.CollectHandler)

	for {
		select {
		case event := <-context.GameEventChan:
			dispatcher.Dispatch(event.Name, event)
		//printAsJson(event)
		case <-context.StopChan:
			statistic.RemoveWarmupRounds()
			statistic.RenumberRounds()
			return statistic
		}
	}
}
