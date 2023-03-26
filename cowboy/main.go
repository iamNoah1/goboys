package main

import (
	"common"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

var me common.Cowboy // todo check if this really needs to be upper
var logger *zap.SugaredLogger

var rabbitMQ *common.RabbitMQ

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

	//Does this have to run as routine neccesarily?
	go participateBattle()

	startFighting()

	defer rabbitMQ.Close()
}

func participateBattle() {
	rabbitMQ.Consume("cowboy-queue", processMessage)
}

func processMessage(body []byte) bool {
	var data common.ShootingData

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

	//Maybe it is not so clever to make this here, because it has some effect on the success of processing or rescheduling the message.
	err = sendHealthUpdate()
	if err != nil {
		var errMessage = fmt.Sprintf("[%s]: failed to publish health update message: %v", me.Name, err)
		logger.Error(errMessage)
		return false
	}

	if me.Health <= 0 {
		logger.Infof("[%s]: I think, I just died", me.Name)
		os.Exit(0)
	} else {
		logger.Infof("[%s]: got shot with damage of %d from %s. Only %d health left", me.Name, data.Damage, data.Source, me.Health)
	}

	return true
}

func sendHealthUpdate() error {
	msg := common.HealthData{
		Name:   me.Name,
		Health: me.Health,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message body: %v", err)
	}

	return rabbitMQ.PublishJSON("", "referee-queue", body)
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
			//TODO kein fatal
			log.Fatalf("failed to marshal message body: %v", err)
		}

		err = rabbitMQ.PublishJSON("", "cowboy-queue", body)

		if err != nil {
			log.Fatalf("failed to publish a message: %v", err)
		}
		logger.Infof("[%s]: shooting!!!", me.Name)

		logger.Infof("[%s]: need to reload ...", me.Name)
		time.Sleep(1 * time.Second)

		qinfo, err := rabbitMQ.GetQueueInfo("cowboy-queue")
		if err != nil {
			log.Fatalf("failed to inspect the queue: %v", err)
		}

		if qinfo.Consumers == 1 {
			logger.Infof("[%s]: no one left, I am the winner", me.Name)
			os.Exit(0)
		}
	}
}
