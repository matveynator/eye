package Database

import (
	"log"
	"fmt"
	"time"

	"database/sql"

	"eye/pkg/mylog"
	"eye/pkg/data"
	"eye/pkg/config"
)

var DataSaveTasks chan Data.ServerData
var SettingsTasks chan Config.Settings
var DatabaseCriticalErrors chan error
var DatabaseWatchdogTasks chan int
var databaseRespawnLock chan int
var watchdogRespawnLock chan int
var WorkerId, WatchdogId int

func init() {

	//initial ids:
	WorkerId = 1
	WatchdogId = 1

	//initialise channel with tasks:
	DataSaveTasks = make(chan Data.ServerData, 1000)

	//initialise channel with tasks:
	SettingsTasks = make(chan Config.Settings, 10)

	//initialize blocking channel with database critical errors:
	DatabaseCriticalErrors = make (chan error, 1)

	//initialize watchdog blocking channel
	DatabaseWatchdogTasks = make(chan int, 1)

	//initialize blocking channel to guard respawn tasks
	databaseRespawnLock = make(chan int, 1)
	watchdogRespawnLock = make(chan int, 1)

}

func Run(config Config.Settings) {

	//start database workers
	if len(databaseRespawnLock)==0 {
		go func() {
			for {
				//lock
				databaseRespawnLock <- 1
				//spawn database worker
				go databaseWorkerRun(WorkerId, config)
				//increment id
				WorkerId++
			}
		}()
	}

	//later start watchdog (to guard database workers)
	if len(watchdogRespawnLock)==0 {
		//later start watchdog (to guard database workers)
		watchdogRespawnLock <- 1
		go watchdogRun(WatchdogId) 
		WatchdogId++
	}
}

func watchdogRun(watchdogId int) {
	for {
		if len(DataSaveTasks) == 0 && len(SettingsTasks) == 0 {
			if len(DatabaseWatchdogTasks) == 0 {
				time.Sleep(5 * time.Second)
				log.Printf("Watchdog %d is alive.", watchdogId)
				DatabaseWatchdogTasks <- 1
			}
		}
	}
}

//close dbConnection on programm exit
func cleanup(db *sql.DB) {
	err := db.Close() 
	if err != nil {
		log.Println("Error closing database connection:", err)
	}
	if len(databaseRespawnLock) == 1 {
		log.Printf("Freeing database worker slot...")
		<- databaseRespawnLock //unlock
	}
}

func databaseWorkerRun(workerId int, config Config.Settings ) {
	log.Printf("Started database worker %d", workerId)
	dbConnection, err := connectToDb(config)
	defer cleanup(dbConnection)

	if err != nil  {
		MyLog.Printonce(fmt.Sprintf("Database %s is unreachable. Error: %s",  config.DB_TYPE, err))
		return
	} else {
		MyLog.Println(fmt.Sprintf("Database worker #%d connected to %s database", workerId, config.DB_TYPE))
	}

	for {
		select {
		case <- DatabaseWatchdogTasks :
			_, err = dbConnection.Exec("UPDATE DBWatchDog SET UnixTime = ? WHERE ID = 1", time.Now().UnixMilli())
			if err != nil {
				//log.Println("len(DatabaseCriticalErrors) : ", len(DatabaseCriticalErrors))
				if len(DatabaseCriticalErrors) == 0 {
					DatabaseCriticalErrors <- err
				}
			} else {
				log.Printf("Database worker %d is alive.", workerId)
			}
		case dataSaveTask := <- DataSaveTasks :
			//в случае если есть задание в канале DataSaveTasks
			_, err := InsertServerDataInDB(dbConnection, dataSaveTask)
			if err != nil {
				log.Printf("Database worker %d exited due to dataSaveTask processing error: %s\n", workerId, err)
				//return data back to channel
				DataSaveTasks <- dataSaveTask
				if len(DatabaseCriticalErrors) == 0 {
					DatabaseCriticalErrors <- err
				}
			}
		case settingsTask := <- SettingsTasks :
				//в случае если есть задание в канале SettingsTasks
				err := SaveSettingsInDB(dbConnection, settingsTask)
				if err != nil {
					log.Printf("Database worker %d exited due to settingsTask processing error: %s\n", workerId, err)
					//return data back to channel
					SettingsTasks <- settingsTask
					if len(DatabaseCriticalErrors) == 0 {
						DatabaseCriticalErrors <- err
					}
					return
				} else {
					log.Println("OK. Settings saved.")
				}
		case databaseCriticalError := <-DatabaseCriticalErrors :
			//обнаружена критическая ошибка бд - завершаем гоурутину
			log.Printf("Database worker %d exited due to critical error: %s\n", workerId, databaseCriticalError)
			return
		default:
			time.Sleep(1 * time.Second)
			log.Printf("Database worker %d sleeping..", workerId)
		}
	}
}

