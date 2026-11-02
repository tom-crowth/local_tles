package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"local_tles/spacetrack"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const config_file = "ltle.config"

var DB *sql.DB

func initDB(config *Config) {
	log.Println("In DB init function")
	var err error
	DB, err = sql.Open("sqlite3", config.database)
	check(err)

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		title TEXT
	);`

	_, err = DB.Exec(sqlStmt)
	check(err)
	log.Println("End of DB init function")
}

type Config struct {
	firstDate         string // or time object?
	readRate          uint8
	serverPort        uint16
	datastore         string
	database          string
	access_spacetrack bool
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func parse_config_file(file string) (Config, error) {
	data, err := os.ReadFile(file)
	check(err)

	config := Config{}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		items := strings.Split(line, " ")
		switch items[0] {
		case "READ_RATE":
			readRate, err := strconv.Atoi(items[1])
			check(err)
			config.readRate = uint8(readRate)
		case "PORT":
			serverPort, err := strconv.Atoi(items[1])
			check(err)
			config.serverPort = uint16(serverPort)
		case "FIRST_DATE":
			config.firstDate = items[1]
		case "DATASTORE":
			config.datastore = items[1]
		case "DATABASE":
			config.database = items[1]
		case "ACCESS_ST":
			access_flag, err := strconv.Atoi(items[1])
			check(err)
			if access_flag == 1 {
				config.access_spacetrack = true
			} else {
				config.access_spacetrack = false
			}
		default:
			err := "Something wrong in the config file?"
			fmt.Println(err)
			// return nil, errors.
		}

	}
	return config, nil
}

func main() {
	default_query_date := "2026-10-26"
	args := os.Args
	query_date := default_query_date
	if len(args) == 2 {
		query_date = args[1]
	}

	ltle_config_dir := os.Getenv("LTLE_CONFIG")
	fmt.Println(ltle_config_dir)

	fmt.Println("Reading config file...")

	config_path := path.Join(ltle_config_dir, config_file)
	// check for existence of file here
	config, err := parse_config_file(config_path)
	check(err)
	fmt.Println(config)
	jconfig, err := json.Marshal(config)
	fmt.Println(string(jconfig))

	// check for query date file
	date_filename := config.datastore + query_date + ".data"
	if _, err := os.Stat(date_filename); err == nil {
		log.Printf("File for specified date already exists:\n\t%s", date_filename)
		return
	}

	// client, err := spacetrack.Setup_session(creds, creds)
	if config.access_spacetrack {
		client, err := spacetrack.Setup_session()
		check(err)
		if err != nil {
			fmt.Println(client)
		}

		defer func() {
			err := spacetrack.End_session(client)
			if err != nil {
				panic(err)
			}
		}()

		resp, err := spacetrack.Get_data(client, query_date)
		check(err)
		log.Printf("Status code from response is %s", resp.Status)
		file, err := os.OpenFile(date_filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o666)
		check(err)
		written, err := io.Copy(file, resp.Body)
		check(err)
		// check if written == 0 here, raise error and delete file if so
		// we should not get an empty file
		// The status code should also have been checked already and needs to be 200
		resp.Body.Close()
		log.Printf("%d bytes written to %s", written, date_filename)
	} else {
		fmt.Println("Spacetrack access disabled in config file. Not accessing")
	}

	initDB(&config)
	defer DB.Close()

	// file_results, err := os.ReadFile("testfile")
	// check(err)
	// fmt.Println(string(file_results))

	fmt.Println("\nEnd of program")
}
