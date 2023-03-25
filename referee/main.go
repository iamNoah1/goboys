package main

import (
	"common"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger
var db *gorm.DB

var rabbitMQ *common.RabbitMQ

func main() {
	logger = common.GetLogger()
	logger.Infoln("starting referee")

	r := gin.Default()

	db, err := connectDB()
	if err != nil {
		logger.Fatalf("[Referee]: failed to connect to db: %v", err)
	}
	defer db.Close()

	rabbitMQ, err = common.NewRabbitMQ("amqp://guest:guest@localhost:5672/")
	if err != nil {
		logger.Fatalf("[Referee]: failed to connect to RabbitMQ: %v", err)
	}

	err = rabbitMQ.DeclareQueue("cowboy-queue")
	if err != nil {
		logger.Fatalf("[Referee]: failed to declare the queue for cowboy communication: %v", err)
	}

	r.POST("/cowboy", saveCowboys(db))
	r.POST("/startShooting", startShooting(db))

	r.Run()
}

func saveCowboys(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		logger.Debugln("[Referee]: entered POST /cowboy endoint")

		var cowboys []Cowboy
		if err := c.ShouldBindJSON(&cowboys); err != nil {
			var errMessage = fmt.Sprintf("[Referee]: could not bind cowboys, try again. Error: %s", err.Error())
			logger.Error(errMessage)
			c.String(http.StatusInternalServerError, errMessage)
		}

		for _, cowboy := range cowboys {
			if err := cowboy.Create(db); err != nil {
				var errMessage = fmt.Sprintf("[Referee]: could not save cowboy, try again. Error: %s", err.Error())
				logger.Error(errMessage)
				c.String(http.StatusInternalServerError, errMessage)
			}
		}
		c.JSON(http.StatusCreated, cowboys)
	}
}

func startShooting(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		logger.Debugln("[Referee]: entered POST /startShooting endoint")

		cowboys, err := GetAllCowboys(db)
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
}

func spawnCowboy(cowboy Cowboy) {
	cmd := exec.Command("../cowboy/cowboy", cowboy.Name, strconv.Itoa(cowboy.Health), strconv.Itoa(cowboy.Damage))
	cmd.Env = os.Environ()
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

func connectDB() (*gorm.DB, error) {
	dbURI := "host=localhost port=5432 user=postgres password=mysecretpassword sslmode=disable"
	db, err := gorm.Open("postgres", dbURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// Automatically create the cowboy table if it doesn't exist
	db.AutoMigrate(&Cowboy{})
	return db, nil
}
