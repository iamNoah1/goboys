package main

import (
	"common"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger
var db *gorm.DB

var rabbitMQ *common.RabbitMQ

const refereeQueueName = "referee-queue"
const cowboyQueueName = "cowboy-queue"

func main() {
	logger = common.GetLogger()
	logger.Infoln("[Referee]: starting sequence ...")

	r := gin.Default()

	initDB()
	logger.Infoln("[Referee]: connected to db")

	//Little hack for when running in docker compose, because apparently rabbitmq needs some time until it is available to connect.
	time.Sleep(10 * time.Second)

	initRabbit()
	logger.Infoln("[Referee]: connected to and initialized rabbitmq")

	r.POST("/cowboy", saveCowboys(db))
	r.POST("/startShooting", startShooting(db))

	go rabbitMQ.Consume(refereeQueueName, processMessage(db))

	defer rabbitMQ.Close()
	defer db.Close()
	r.Run()
}

func initDB() {
	var err error
	db, err = connectDB()
	if err != nil {
		logger.Fatalf("[Referee]: failed to connect to db: %v", err)
	}
	err = ClearCowboys(db)
	if err != nil {
		logger.Fatalf("[Referee]: could not clear cowboys: %v", err)
	}
}

func initRabbit() {
	rabbitHost := os.Getenv("RABBIT_HOST")
	if os.Getenv("RABBIT_HOST") == "" {
		rabbitHost = "localhost"
	}

	rabbitConnectionString := fmt.Sprintf("amqp://guest:guest@%s:5672/", rabbitHost)
	logger.Debugln(rabbitConnectionString)

	var err error
	rabbitMQ, err = common.NewRabbitMQ(rabbitConnectionString)
	if err != nil {
		logger.Fatalf("[Referee]: failed to connect to RabbitMQ: %v", err)
	}

	err = rabbitMQ.DeclareQueue(cowboyQueueName)
	if err != nil {
		logger.Fatalf("[Referee]: failed to declare the queue for cowboy communication: %v", err)
	}
	_, err = rabbitMQ.PurgeQueue(cowboyQueueName)
	if err != nil {
		logger.Fatalf("[Referee]: failed to purge queue for cowboy communication: %v", err)
	}

	err = rabbitMQ.DeclareQueue(refereeQueueName)
	if err != nil {
		logger.Fatalf("[Referee]: failed to declare the queue for referee communication: %v", err)
	}
	_, err = rabbitMQ.PurgeQueue(refereeQueueName)
	if err != nil {
		logger.Fatalf("[Referee]: failed to declare the queue for referee communication: %v", err)
	}
}

func processMessage(db *gorm.DB) func(body []byte) bool {
	return func(body []byte) bool {
		var data common.HealthData

		err := json.Unmarshal(body, &data)
		if err != nil {
			var errMessage = fmt.Sprintf("[Referee]: Error while updating cowboy. Error: %s", err.Error())
			logger.Error(errMessage)
			return false
		}

		c := Cowboy{
			Name:   data.Name,
			Health: data.Health,
		}

		if c.Health <= 0 {
			if err := c.Delete(db); err != nil {
				var errMessage = fmt.Sprintf("[Referee]: could not delete cowboy. Error: %s", err.Error())
				logger.Error(errMessage)
				return false
			}
		} else {
			if err := c.Update(db); err != nil {
				var errMessage = fmt.Sprintf("[Referee]: could not update cowboy. Error: %s", err.Error())
				logger.Error(errMessage)
				return false
			}
		}

		return true
	}
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
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "mysecretpassword"
	}

	dbSSLMode := os.Getenv("DB_SSLMODE")
	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}

	dbURI := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s", dbHost, dbPort, dbUser, dbPassword, dbSSLMode)

	db, err := gorm.Open("postgres", dbURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// Automatically create the cowboy table if it doesn't exist
	db.AutoMigrate(&Cowboy{})
	return db, nil
}
