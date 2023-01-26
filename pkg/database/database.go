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

var DatabaseTask chan Data.ServerData
var SettingsTaskChannel chan Config.Settings

var dbRespawnLock chan int
//по умолчанию оставляем только один процесс который будет брать задачи и записывать их в базу данных
var databaseWorkersMaxCount int = 1

func init() {
	//initialise channel with tasks:
	DatabaseTask = make(chan Data.ServerData)

	//initialise channel with tasks:
	SettingsTaskChannel = make(chan Config.Settings)

	//initialize blocking channel to guard respawn tasks
	dbRespawnLock = make(chan int, databaseWorkersMaxCount)
}

func Run(config Config.Settings) {

	go func() {
		for {
			log.Printf("Locking... Respawn lock count: %d\n", len(dbRespawnLock))
			// will block if there is databaseWorkersMaxCount in dbRespawnLock
			dbRespawnLock <- 1
			//sleep 1 second
			log.Printf("Locked. Respawn lock count: %d\n", len(dbRespawnLock))
			time.Sleep(1 * time.Second)
			go databaseWorkerRun(len(dbRespawnLock), config)
		}
	}()
}

//close dbConnection on programm exit
func deferCleanup(db *sql.DB) {
	err := db.Close() 
	if err != nil {
		log.Println("Error closing database connection:", err)
	}
	log.Printf("Unlocking... Respawn lock count: %d\n", len(dbRespawnLock))
	<- dbRespawnLock //unlock
	log.Printf("Unlocked. Respawn lock count: %d\n", len(dbRespawnLock))
}

func databaseWorkerRun(workerId int, config Config.Settings ) {
	log.Printf("Started database worker %d", workerId)
	dbConnection, err := connectToDb(config)
	defer deferCleanup(dbConnection)

	if err != nil  {
		MyLog.Printonce(fmt.Sprintf("Database %s is unreachable. Error: %s",  config.DB_TYPE, err))
		return

	} else {
		MyLog.Println(fmt.Sprintf("Database worker #%d connected to %s database", workerId, config.DB_TYPE))
	}

	//initialise dbConnection error channel
	connectionErrorChannel := make(chan error)

	go func() {
		defer deferCleanup(dbConnection)
		for {
			time.Sleep(5 * time.Second)
			_, err = dbConnection.Exec("UPDATE DBWatchDog SET UnixTime = ? WHERE ID = 1", time.Now().UnixMilli())
			if err != nil {
				connectionErrorChannel <- err
				return
			} else {
				log.Println("Database is alive.")
			}
		}
	}()

	for {
		select {
			//в случае если есть задание в канале DatabaseTask
		case currentDatabaseTask := <- DatabaseTask :
			//log.Println("Received new database task with TagID:", currentDatabaseTask.TagID)
			_, err := InsertServerDataInDB(dbConnection, currentDatabaseTask)
			if err != nil {
				log.Printf("Database worker %d exited due to DatabaseTask processing error: %s\n", workerId, err)
				return
			}
			//в случае если есть задание в канале SettingsTaskChannel
		case currentSettingsTask := <- SettingsTaskChannel :
			go func() {
				//log.Println("Received new database task with TagID:", currentDatabaseTask.TagID)
				err := SaveSettingsInDB(dbConnection, currentSettingsTask)
				if err != nil {
					log.Printf("Database worker %d exited due to SettingsTaskChannel processing error: %s\n", workerId, err)
					//return task back
					time.Sleep(1 * time.Second)
					log.Printf("Returning task back... SettingsTaskChannel channel count: %d\n", len(SettingsTaskChannel))

					SettingsTaskChannel <- currentSettingsTask
					log.Printf("Returned. SettingsTaskChannel channel count: %d\n", len(SettingsTaskChannel))
					//return
				} else {
					log.Println("OK. Settings saved.")
					log.Printf("SettingsTaskChannel channel count: %d\n", len(SettingsTaskChannel))
				}
			}()
		case networkError := <-connectionErrorChannel :
			//обнаружена сетевая ошибка - завершаем гоурутину
			log.Printf("Database worker %d exited due to connection error: %s\n", workerId, networkError)
			return
		}
	}
}

