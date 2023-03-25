package common

import (
	"os"

	"go.uber.org/zap"
)

func GetLogger() *zap.SugaredLogger {
	loglevel := os.Getenv("LOG_LEVEL")

	var l *zap.Logger

	if loglevel == "prod" {
		l, _ = zap.NewProduction()
	} else {
		l = zap.NewExample()
	}

	defer l.Sync()
	return l.Sugar()
}
