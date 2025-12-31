package internal

import (
	"log"
	"os"
)

var Logger *log.Logger

func InitLogger() {
	f, err := os.OpenFile(ConfigData.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		panic(err)
	}
	Logger = log.New(f, "", log.LstdFlags)
}
