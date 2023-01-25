package main

import (
	"time"
	"eye/pkg/config"
	"eye/pkg/database"
	"eye/pkg/mylog"
)

func main() {

	settings := Config.ParseFlags()

	//run error log daemon
	go MyLog.ErrorLogWorker()
	go Database.Run(settings)

	for {
		time.Sleep(10 * time.Second)	
	}

}
