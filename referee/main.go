package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"common"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var mutex = &sync.Mutex{}
var logger *zap.SugaredLogger

func main() {
	logger = common.GetLogger()
	logger.Infoln("starting referee")

	r := gin.Default()
	r.POST("/startShooting", startShooting)
	r.POST("/cowboy", saveCowboys)
	r.GET("/cowboy", getCowboys)
	r.PUT("/cowboy/:name", updateCowboy)
	r.DELETE("/cowboy/:name", deleteCowboy)
	r.Run()
}

func saveCowboys(c *gin.Context) {
	logger.Debugln("[Referee]: entered POST /cowboy endoint")
	var cowboys []common.Cowboy

	if c.ShouldBind(&cowboys) == nil {
		logger.Debugln("[Referee]: registered cowboys: ")
		logger.Debugln(&cowboys)
	}

	port := 3000

	for i, cowboy := range cowboys {
		uri := fmt.Sprintf("http://localhost:%s", strconv.Itoa(port))
		cowboy.URI = uri
		cowboys[i] = cowboy
		port++
	}

	err := WriteCowboys(cowboys)

	if nil != err {
		var errMessage = fmt.Sprintf("[Referee]: could not save cowboys, try again. Error: %s", err.Error())
		logger.Error(errMessage)
		c.String(http.StatusInternalServerError, errMessage)
	} else {
		logger.Info("[Referee]: saved cowboys, ready to start the show off")
		c.Status(http.StatusCreated)
	}
}

func deleteCowboy(c *gin.Context) {
	logger.Debugln("[Referee]: entered DELETE /cowboy endoint")

	name := c.Param("name")

	cowboys, err := ReadCowboys()
	if nil != err {
		var errMessage = fmt.Sprintf("[Referee]: could not read cowboys for deletion. Error: %s", err.Error())
		logger.Error(errMessage)
		c.String(http.StatusInternalServerError, errMessage)
	}

	for i, cowboy := range cowboys {
		if cowboy.Name == name {
			cowboys = remove(cowboys, i)
		}
	}

	err = WriteCowboys(cowboys)

	if nil != err {
		var errMessage = fmt.Sprintf("[Referee]: could not save cowboys. Error: %s", err.Error())
		logger.Error(errMessage)
		c.String(http.StatusInternalServerError, errMessage)
	} else {
		logger.Info("[Referee]: updated cowboys")
		c.Status(http.StatusOK)
	}
}

func remove(s []common.Cowboy, i int) []common.Cowboy {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func updateCowboy(c *gin.Context) {
	logger.Debugln("[Referee]: entered PUT /cowboy endoint")

	name := c.Param("name")

	cowboys, err := ReadCowboys()
	if nil != err {
		var errMessage = fmt.Sprintf("[Referee]: could not read cowboys for update. Error: %s", err.Error())
		logger.Error(errMessage)
		c.String(http.StatusInternalServerError, errMessage)
	}

	for i, cowboy := range cowboys {
		if cowboy.Name == name {
			var updatedCowboy common.Cowboy

			if c.ShouldBind(&updatedCowboy) == nil {
				cowboys[i].Health = updatedCowboy.Health
			}
		}
	}

	err = WriteCowboys(cowboys)

	if nil != err {
		var errMessage = fmt.Sprintf("[Referee]: could not save cowboys. Error: %s", err.Error())
		logger.Error(errMessage)
		c.String(http.StatusInternalServerError, errMessage)
	} else {
		logger.Info("[Referee]: updated cowboy")
		c.Status(http.StatusOK)
	}
}

func getCowboys(c *gin.Context) {
	logger.Debugln("[Referee]: entered GET /cowboy endoint")

	cowboys, err := ReadCowboys()
	if nil != err {
		var errMessage = fmt.Sprintf("[Referee]: could not read cowboys. Error: %s", err.Error())
		logger.Error(errMessage)
		c.String(http.StatusInternalServerError, errMessage)
	}

	c.JSON(http.StatusOK, cowboys)
}

func startShooting(c *gin.Context) {
	logger.Debugln("[Referee]: entered POST /startShooting endoint")

	cowboys, err := ReadCowboys()
	if nil != err {
		var errMessage = fmt.Sprintf("[Referee]: could not read cowboys. Error: %s", err.Error())
		logger.Error(errMessage)
		c.String(http.StatusInternalServerError, errMessage)
	}

	for _, cowboy := range cowboys {
		go spawnCowboy(cowboy)
	}
	c.String(http.StatusCreated, "shooting started")
}

func spawnCowboy(cowboy common.Cowboy) {
	split := strings.Split(cowboy.URI, ":")
	portEnv := fmt.Sprintf("PORT=%s", split[2])

	cmd := exec.Command("../cowboy/cowboy", cowboy.Name, strconv.Itoa(cowboy.Health), strconv.Itoa(cowboy.Damage))
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, portEnv)
	cmd.Env = append(cmd.Env, "ORCH_URI=http://localhost:8080")
	cmd.Env = append(cmd.Env, "GIN_MODE="+os.Getenv("GIN_MODE"))
	cmd.Env = append(cmd.Env, "LOG_LEVEL="+os.Getenv("LOG_LEVEL"))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatal("[Referee]: could not spawn cowboy: ", err)
	}
	defer cmd.Wait()
}

func WriteCowboys(cowboys []common.Cowboy) error {
	mutex.Lock()
	defer mutex.Unlock()

	_, err := os.Create("./cowboy-db") // do we need that?
	if nil != err {
		return err
	}

	content, _ := json.Marshal(cowboys)
	return ioutil.WriteFile("./cowboy-db", content, 0644)
}

func ReadCowboys() ([]common.Cowboy, error) {
	mutex.Lock()
	defer mutex.Unlock()

	raw, err := ioutil.ReadFile("./cowboy-db")

	if nil != err {
		return nil, err
	}

	var cowboys []common.Cowboy
	err = json.Unmarshal(raw, &cowboys)

	if nil != err {
		return nil, err
	}

	return cowboys, err
}
