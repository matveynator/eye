package Config

import (
	"os"
	"fmt"
	"flag"
)


type Settings struct {
	APP_NAME, VERSION, DB_TYPE, DB_FILE_PATH, DB_FULL_FILE_PATH, PG_HOST, PG_USER, PG_PASS, PG_DB_NAME, PG_SSL, HETZNER_ROBOT_USER, HETZNER_ROBOT_PASS string
	PG_PORT int
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func ParseFlags() (config Settings)  { 
	config.APP_NAME = "eye"
	flagVersion := flag.Bool("version", false, "Output version information")

	//db
	flag.StringVar(&config.DB_FILE_PATH, "dbpath", ".", "Provide path to writable directory to store database data.")
	flag.StringVar(&config.DB_TYPE, "dbtype", "sqlite", "Select db type: sqlite / genji / postgres")

	//PostgreSQL related start
	flag.StringVar(&config.PG_HOST, "pghost", "127.0.0.1", "PostgreSQL DB host.")
	flag.IntVar(&config.PG_PORT, "pgport", 5432, "PostgreSQL DB port.")
	flag.StringVar(&config.PG_USER, "pguser", "postgres", "PostgreSQL DB user.")
	flag.StringVar(&config.PG_PASS, "pgpass", "", "PostgreSQL DB password.")
	flag.StringVar(&config.PG_DB_NAME, "pgdbname", "eye", "PostgreSQL DB name.")
	flag.StringVar(&config.PG_SSL, "pgssl", "prefer", "disable / allow / prefer / require / verify-ca / verify-full - PostgreSQL ssl modes: https://www.postgresql.org/docs/current/libpq-ssl.html")

	//hetzner 
	flag.StringVar(&config.HETZNER_ROBOT_USER, "hetzner-user", "", "Hetzner robot user name.")
	flag.StringVar(&config.HETZNER_ROBOT_PASS, "hetzner-pass", "", "Hetzner robot password.")

	//process all flags
	flag.Parse()


	//путь к файлу бд
	config.DB_FULL_FILE_PATH = fmt.Sprintf(config.DB_FILE_PATH+"/"+config.APP_NAME+".db."+config.DB_TYPE)

	if *flagVersion  {
		if config.VERSION != "" {
			fmt.Println("Version:", config.VERSION)
		}
		os.Exit(0)
	}

	return
}
