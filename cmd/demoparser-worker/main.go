package main

import (
	"log"
	//"time"
	"encoding/json"
	"github.com/stegmannc/csgo-demoparser"
	"github.com/streadway/amqp"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type DemoParseRequest struct {
	Id   string `json:"id"`
	File string `json:"file"`
}

type DemoParsingResult struct {
	Id        string              `json:"id"`
	MatchInfo *demoinfo.MatchInfo `json:"matchInfo"`
	Rounds    []*demoinfo.Round   `json:"rounds"`
}

func main() {

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	uploadQueue, err := ch.QueueDeclare(
		"demo-upload", // name
		true,   // durable
		false,   // delete when usused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	_, err = ch.QueueDeclare(
		"demo_info_upload", // name
		true,   // durable
		false,   // delete when usused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	requests, err := ch.Consume(
		uploadQueue.Name, // queue
		"demo-parser", // consumer
		true,          // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range requests {
			log.Printf("Received a message: %s", d.Body)
			request := &DemoParseRequest{}
			if err := json.Unmarshal(d.Body, request); err != nil {
				log.Println("could not parse demo request json...")
				continue
			}
			go parseDemo(ch, request)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}

func parseDemo(ch *amqp.Channel, request *DemoParseRequest) {
	start := time.Now()
	parser, err := demoinfo.NewDemoParser(request.File)
	if err != nil {
		log.Fatal(err)
	}

	go parser.ParseTicks()
	statistics := aggregateEvents(parser.Context, parser.Header)
	demoinfo.PrintAsJson(statistics)
	elapsed := time.Since(start)
	log.Printf("Parsed demo %s in %s", request.Id, elapsed)

	result := DemoParsingResult{Id: request.Id, MatchInfo: statistics.MatchInfo, Rounds: statistics.Rounds}
	data, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}

	err = ch.Publish(
		"",                 // exchange
		"demo_info_upload", // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		})
	if err != nil {
		panic(err)
	}
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
