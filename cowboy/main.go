package main

import (
	"common"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

var me common.Cowboy // todo check if this really needs to be upper
var logger *zap.SugaredLogger

var rabbitMQ *common.RabbitMQ

type shootingData struct {
	Source string `json:"source"`
	Damage int    `json:"damage"`
}

func main() {
	logger = common.GetLogger()

	if 4 != len(os.Args) {
		logger.Fatal("Wrong amoung of arguments! Usage: go run main.go <name> <health> <damage>")
	}

	spawn()

	var err error
	rabbitMQ, err = common.NewRabbitMQ("amqp://guest:guest@localhost:5672/")
	if err != nil {
		logger.Fatalf("[%s]: failed to connect to RabbitMQ: %v", me.Name, err)
	}

	go participateBattle()

	startFighting()

	//defer rabbitMQ.Close()
}

func participateBattle() {
	rabbitMQ.Consume(processMessage)
}

func processMessage(body []byte) bool {
	var data shootingData

	err := json.Unmarshal(body, &data)
	if err != nil {
		var errMessage = fmt.Sprintf("[%s]: Error while taking a shot. Error: %s", me.Name, err.Error())
		logger.Error(errMessage)
		return false
	}

	if me.Name == data.Source {
		return false
	}

	me.Health = me.Health - data.Damage
	if me.Health <= 0 {
		//send a message to rabbitmq
		logger.Infof("[%s]: I think, I just died", me.Name)
		os.Exit(0)
	} else {
		logger.Infof("[%s]: got shot with damage of %d from %s. Only %d health left", me.Name, data.Damage, data.Source, me.Health)
		//health update message
	}

	return true
}

func spawn() {
	name := os.Args[1]

	health, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}

	damage, err := strconv.Atoi(os.Args[3])
	if err != nil {
		panic(err)
	}

	me = common.Cowboy{
		Name:   name,
		Health: health,
		Damage: damage,
	}
	logger.Infof("spawning cowboy: %s, %d, %d", me.Name, me.Health, me.Damage)

}

func startFighting() {
	logger.Infof("[%s]: I am ready to fight.", me.Name)

	for me.Health > 0 {
		msg := map[string]interface{}{
			"source": me.Name,
			"damage": me.Damage,
		}

		body, err := json.Marshal(msg)
		if err != nil {
			log.Fatalf("failed to marshal message body: %v", err)
		}

		err = rabbitMQ.Ch.Publish(
			"",             // exchange
			"cowboy-queue", // routing key
			false,          // mandatory
			false,          // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			},
		)

		if err != nil {
			log.Fatalf("failed to publish a message: %v", err)
		}
		fmt.Println("Message sent:", msg)

		logger.Infof("[%s]: need to reload ...", me.Name)
		time.Sleep(1 * time.Second)

		// Check if there are no consumers
		qinfo, err := rabbitMQ.Ch.QueueInspect("cowboy-queue")
		if err != nil {
			log.Fatalf("failed to inspect the queue: %v", err)
		}

		logger.Infof("consumers: %d", qinfo.Consumers)
		if qinfo.Consumers == 1 {
			logger.Infof("[%s]: no one left, I am the winner", me.Name)
			os.Exit(0)
		}
	}
}
