package main

import (
	"flag"
	"log"
	"time"
)

var filename = flag.String("f", "", "Path to *.dem file")

func main() {
	flag.Parse()

	start := time.Now()
	parser, err := NewDemoParser(*filename)
	if err != nil {
		log.Fatal(err)
	}

	go parser.ParseTicks()
	statistics := aggregateEvents(parser.Context, parser.Header)
	printAsJson(statistics)
	elapsed := time.Since(start)
	log.Printf("Demo parsing took %s", elapsed)
}

func aggregateEvents(context *DemoContext, demoHeader *DemoHeader) *DemoStatistic {
	statistic := NewDemoStatistic(demoHeader)
	dispatcher := NewDemoEventDispatcher(statistic, context)
	dispatcher.RegisterHandler("player_connect", PlayerConnectHandler)
	dispatcher.RegisterHandler("player_disconnect", PlayerDisconnectHandler)
	dispatcher.RegisterHandler("player_team", PlayerTeamHandler)
	dispatcher.RegisterHandler("round_announce_match_start", RoundAnnounceMatchStartHandler)
	dispatcher.RegisterHandler("round_start", RoundStartHandler)
	dispatcher.RegisterHandler("round_end", RoundEndHandler)
	dispatcher.RegisterHandler("round_officially_ended", RoundOfficiallyEndHandler)
	dispatcher.RegisterHandler("cs_win_panel_match", CalculateMatchDurationHandler)

	dispatcher.RegisterHandler("player_hurt", CollectHandler)
	dispatcher.RegisterHandler("player_death", CollectHandler)
	dispatcher.RegisterHandler("item_purchase", CollectHandler)
	dispatcher.RegisterHandler("bomb_beginplant", CollectHandler)
	dispatcher.RegisterHandler("bomb_abortplant", CollectHandler)
	dispatcher.RegisterHandler("bomb_planted", CollectHandler)
	dispatcher.RegisterHandler("bomb_defused", CollectHandler)
	dispatcher.RegisterHandler("bomb_exploded", CollectHandler)
	dispatcher.RegisterHandler("bomb_dropped", CollectHandler)
	dispatcher.RegisterHandler("bomb_pickup", CollectHandler)
	dispatcher.RegisterHandler("defuser_pickup", CollectHandler)
	dispatcher.RegisterHandler("bomb_begindefuse", CollectHandler)
	dispatcher.RegisterHandler("bomb_abortdefuse", CollectHandler)
	dispatcher.RegisterHandler("weapon_fire", CollectHandler)
	dispatcher.RegisterHandler("silencer_detach", CollectHandler)
	dispatcher.RegisterHandler("inspect_weapon", CollectHandler)
	dispatcher.RegisterHandler("enter_bombzone", CollectHandler)
	dispatcher.RegisterHandler("exit_bombzone", CollectHandler)
	dispatcher.RegisterHandler("grenade_bounce", CollectHandler)
	dispatcher.RegisterHandler("flashbang_detonate", CollectHandler)
	dispatcher.RegisterHandler("smokegrenade_detonate", CollectHandler)
	dispatcher.RegisterHandler("smokegrenade_expired", CollectHandler)
	dispatcher.RegisterHandler("molotov_detonate", CollectHandler)
	dispatcher.RegisterHandler("decoy_detonate", CollectHandler)
	dispatcher.RegisterHandler("decoy_started", CollectHandler)
	dispatcher.RegisterHandler("tagrenade_detonate", CollectHandler)
	dispatcher.RegisterHandler("inferno_startburn", CollectHandler)
	dispatcher.RegisterHandler("inferno_expire", CollectHandler)
	dispatcher.RegisterHandler("inferno_extinguish", CollectHandler)
	dispatcher.RegisterHandler("bullet_impact", CollectHandler)
	dispatcher.RegisterHandler("player_blind", CollectHandler)
	dispatcher.RegisterHandler("player_falldamage", CollectHandler)
	dispatcher.RegisterHandler("door_moving", CollectHandler)
	dispatcher.RegisterHandler("player_avenged_teammate", CollectHandler)
	dispatcher.RegisterHandler("match_end_conditions", CollectHandler)
	dispatcher.RegisterHandler("round_mvp", CollectHandler)
	dispatcher.RegisterHandler("player_given_c4", CollectHandler)
	dispatcher.RegisterHandler("tr_player_flashbanged", CollectHandler)
	dispatcher.RegisterHandler("bot_takeover", CollectHandler)
	dispatcher.RegisterHandler("door_moving", CollectHandler)

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
