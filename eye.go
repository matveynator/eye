package main

import (
	"log"
	"time"
	"eye/pkg/config"
	"eye/pkg/database"
	"eye/pkg/mylog"
	"eye/pkg/hetzner"
)

func main() {

	settings := Config.ParseFlags()

	//run error log daemon
	go MyLog.ErrorLogWorker()
	go Database.Run(settings)
	go Database.Run(settings)
	go Database.Run(settings)

	err := Hetzner.UpdateCredentials(settings)
	if err != nil {
		log.Println("HZ err:", err)
	}

	for {
		time.Sleep(10 * time.Second)	
	}

}
