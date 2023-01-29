package Hetzner

import (
	"log"
	"errors"
	"eye/pkg/config"
	"eye/pkg/database"

	client "github.com/nl2go/hrobot-go"
)

func UpdateCredentials(settings Config.Settings) (err error) {
	if settings.HETZNER_ROBOT_USER != "" && settings.HETZNER_ROBOT_PASS != "" {
		log.Printf("robot user: %s\n", settings.HETZNER_ROBOT_USER)
		log.Printf("robot pass: %s\n", settings.HETZNER_ROBOT_PASS)

		robotClient := client.NewBasicAuthClient(settings.HETZNER_ROBOT_USER, settings.HETZNER_ROBOT_PASS)
		_, err = robotClient.IPGetList()
		if err != nil {
			return
		} else {
			log.Println("OK: connected to hetzner.")
			Database.SettingsTasks <- settings 
		}
	} else {
		err = errors.New("Error: hetzner credentials undefined.")
	}
	return
}
