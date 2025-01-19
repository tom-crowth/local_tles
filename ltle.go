package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"ltle/spacetrack"
	// "io"
)

const config_file = "config/ltle.config"

type Config struct {
	firstDate         string // or time object?
	readRate          uint8
	serverPort        uint16
	datastore         string
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
	fmt.Println("Reading config file...")

	config, err := parse_config_file(config_file)
	check(err)
	fmt.Println(config)
	jconfig, err := json.Marshal(config)
	fmt.Println(string(jconfig))

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

		query_date := "2025-01-05"
		resp, err := spacetrack.Get_data(client, query_date)
		check(err)
		log.Printf("Status code from response is %s", resp.Status)
		filename := config.datastore + query_date
		file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o666)
		check(err)
		written, err := io.Copy(file, resp.Body)
		check(err)
		resp.Body.Close()
		log.Printf("%d bytes written to %s", written, filename)
	} else {
		fmt.Println("Spacetrack access disabled in config file. Not accessing")
	}

	// file_results, err := os.ReadFile("testfile")
	// check(err)
	// fmt.Println(string(file_results))

	fmt.Println("\nEnd of program")
}
