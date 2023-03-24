package main

import (
	"bytes"
	"common"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var me common.Cowboy // todo check if this really needs to be upper
var logger *zap.SugaredLogger

var orchestratorURI string

func main() {
	logger = common.GetLogger()

	validateReadyness()

	spawn()

	r := gin.Default()
	r.POST("/shot", takeShot)

	go startBattle()

	r.Run()
}

func validateReadyness() {
	if 4 != len(os.Args) {
		logger.Fatal("Wrong amoung of arguments! Usage: go run main.go <name> <health> <damage>")
	}
	orchestratorURI = os.Getenv("ORCH_URI")
	if orchestratorURI == "" {
		logger.Fatal("env variable ORCH_URI needs to be set")
	}
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

func startBattle() {
	logger.Infof("%s here, I am ready to fight.", me.Name)
	logger.Infof("%s here, selecting target", me.Name)

	for me.Health > 0 {
		target := getTarget()
		logger.Infof("%s here, selected %s as target. Going to shoot now ...", me.Name, target.Name)

		query := make(map[string]string)
		query["damage"] = strconv.Itoa(me.Damage)
		_, err := common.MakeHttpRequest(http.MethodPost, target.URI+"/shot", nil, query, nil)

		//We don't consider the cowboy not being reachable anymore as an error, because then he is just dead
		if err != nil && !strings.Contains(err.Error(), "EOF") && !strings.Contains(err.Error(), "connection refused") {
			errMessage := fmt.Sprintf("%s here, errored when trying to shoot: ", me.Name)
			logger.Errorln(errMessage, err)
		}
		logger.Infof("%s here, need to reload ...", me.Name)
		time.Sleep(1 * time.Second)
	}
}

func getTarget() common.Cowboy {
	body, err := common.MakeHttpRequest(http.MethodGet, orchestratorURI+"/cowboy", nil, nil, nil)
	if err != nil {
		//TODO aussteigen?
		logger.Errorln(err)
	}

	var cowboys []common.Cowboy
	json.Unmarshal([]byte(body), &cowboys)

	cowboys = removeMySelf(cowboys, me)
	if len(cowboys) == 0 {
		logger.Infof("%s here, no one left, I am the winner", me.Name)
		os.Exit(0)
	}

	randomIndex := rand.Intn(len(cowboys))
	return cowboys[randomIndex]
}

func removeMySelf(cowboys []common.Cowboy, me common.Cowboy) []common.Cowboy {
	for i, c := range cowboys {
		if c.Name == me.Name {
			cowboys[i] = cowboys[len(cowboys)-1]
			return cowboys[:len(cowboys)-1]
		}
	}
	return nil
}

func takeShot(c *gin.Context) {
	logger.Debugln("entered /shot endoint")

	damageString := c.Query("damage")
	logger.Debugf("query param damage: %s", damageString)

	damage, err := strconv.Atoi(damageString)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Debugf("damage: %d", damage)

	me.Health = me.Health - damage
	if me.Health <= 0 {
		_, err := common.MakeHttpRequest(http.MethodDelete, orchestratorURI+"/cowboy/"+me.Name, nil, nil, nil)
		if err != nil {
			logger.Error(err)
			c.Status(http.StatusInternalServerError)
		}
		logger.Infof("%s here, I think I just died", me.Name)
		os.Exit(0)
	} else {
		logger.Infof("%s here, got shot with damage of %d. Only %d health left", me.Name, damage, me.Health)

		bodyMessage := fmt.Sprintf(`{"Health": %d}`, me.Health)
		jsonBody := []byte(bodyMessage)
		bodyReader := bytes.NewReader(jsonBody)

		headers := make(map[string]string)
		headers["Content-Type"] = "application/json"
		_, err := common.MakeHttpRequest(http.MethodPut, orchestratorURI+"/cowboy/"+me.Name, bodyReader, nil, headers)
		if err != nil {
			logger.Error(err)
			c.Status(http.StatusInternalServerError)
		}
	}

	c.Status(http.StatusOK)
}
