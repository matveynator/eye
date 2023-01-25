package Database

import (
	"fmt"
	"log"
	"errors"
	"time"

	"database/sql"

	"eye/pkg/config"
	"eye/pkg/data"
)


func connectToDb(config Config.Settings)(db *sql.DB, err error) {
	if config.DB_TYPE == "genji" {
		db, err = sql.Open(config.DB_TYPE, config.DB_FULL_FILE_PATH)
		if err != nil {
			config.DB_TYPE = "sqlite"
			log.Println("Database error:", err)
			log.Println("Genji is unsupported on this architecture, switching to sqlite db type.")
			db, err = sql.Open(config.DB_TYPE, config.DB_FULL_FILE_PATH)
			if err != nil {
				err = errors.New(fmt.Sprintf("Database file error: %s", err.Error()))
				log.Println(err)
				log.Println("SQLite is unsupported on this architecture, please use: -dbtype postgres.")
				return
			} else {
				err = createTables(db, config)
				if err != nil {
					err = errors.New(fmt.Sprintf("Database create tables error: %s", err.Error()))
					log.Println(err)
					return
				}
			}
		} else {
			err = createTables(db, config)
			if err != nil {
				err = errors.New(fmt.Sprintf("Database create tables error: %s", err.Error()))
				log.Println(err)
				return
			}
		}
	} else if config.DB_TYPE == "sqlite" {
		db, err = sql.Open(config.DB_TYPE, config.DB_FULL_FILE_PATH)
		if err != nil {
			config.DB_TYPE = "genji"
			log.Println("Database file error:", err)
			log.Println("SQLite is unsupported on this architecture, switching to genji db type.")
			db, err = sql.Open(config.DB_TYPE, config.DB_FULL_FILE_PATH)
			if err != nil {
				err = errors.New(fmt.Sprintf("Database file error: %s", err.Error()))
				log.Println(err)
				log.Println("Genji is unsupported on this architecture, please use: -dbtype postgres.")
				return
			} else {
				err = createTables(db, config)
				if err != nil {
					err = errors.New(fmt.Sprintf("Database create tables error: %s", err.Error()))
					log.Println(err)
					return
				}
			}
		} else {
			err = createTables(db, config)
			if err != nil {
				err = errors.New(fmt.Sprintf("Database create tables error: %s", err.Error()))
				log.Println(err)
				return
			}
		}
	} else if config.DB_TYPE == "postgres" {

		psqlConnectDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s pool_max_conns=10", config.PG_HOST, config.PG_PORT, config.PG_USER, config.PG_PASS, config.PG_DB_NAME, config.PG_SSL)
		db, err = sql.Open("pgx", psqlConnectDSN)
		if err != nil {
			err = errors.New(fmt.Sprintf("Database config error: %s", err.Error()))
			log.Println(err)
			return
		}
		err = db.Ping()
		if err != nil {
			err = errors.New(fmt.Sprintf("Database connect error: %s", err.Error()))
			log.Println(err)
			return
		} else {
			err = createTables(db, config)
			if err != nil {
				err = errors.New(fmt.Sprintf("Database create tables error: %s", err.Error()))
				log.Println(err)
				return
			}
		}
	} else {
		err = errors.New("Please define valid dbtype (genji / sqlite)")
		log.Println(err)
		return
	}
	return
}


func createTables(db *sql.DB, config Config.Settings) (err error) {

	_, err = db.Exec("CREATE TABLE if not exists DBWatchDog(Id INT PRIMARY KEY, UnixTime INT)")
	if err != nil {
		return
	} else {
		//populate DBWatchDog with data (one row with only one Id=1)
		var id int64
		// Create a sql/database DB instance
		err = db.QueryRow("SELECT Id FROM DBWatchDog").Scan(&id)
		if err != nil  {
			_, err = db.Exec("INSERT INTO DBWatchDog (Id,UnixTime) VALUES (?,?)", 1, time.Now().UnixMilli())
			if err != nil {
				return
			} else {
				log.Printf("Created new %s database file: %s \n", config.DB_TYPE, config.DB_FULL_FILE_PATH)
			}
		}
	}

	_, err = db.Exec("CREATE TABLE if not exists ServerData(Id INT NOT NULL, IP TEXT, Label TEXT,  Info TEXT, PRIMARY KEY(Id))")
	if err != nil {
		return
	}

	return
}



func InsertServerDataInDB (databaseConnection *sql.DB, serverData Data.ServerData) (id int64, err error) {
	err = databaseConnection.QueryRow("SELECT Id FROM ServerData order by Id desc limit 1").Scan(&id)
	if err != nil {
		id=1
	} else {
		id++
	}

	_, err = databaseConnection.Exec("INSERT INTO ServerData(Id,IP,Label,Info) VALUES (?, ?, ?, ?)", id, serverData.IP, serverData.Label, serverData.Info)
	return
}
