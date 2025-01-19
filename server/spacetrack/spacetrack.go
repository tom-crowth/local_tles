package spacetrack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
)

const (
	creds_file = "config/ltle.credentials"
)

const (
	base_url    = "https://www.space-track.org"
	controller  = "/basicspacedata"
	gph_class   = "/query/class/gp_history"
	norad       = "/norad_cat_id"
	metadata    = "/metadata/true"
	distinct    = "/distinct/true"
	json_format = "/format/json"
	login       = "/ajaxauth/login"
	logout      = "/ajaxauth/logout"
)

type Credentials struct {
	username string
	password string
}

type SpacetrackSession struct {
	httpClient    *http.Client
	credentials   Credentials
	timeout       int // what is this supposed to be
	authenticated bool
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func build_request(date string) (request string) {
	// base gp_history request
	request = base_url + controller
	request += gph_class

	// add overall record count to response, remove duplicates
	request += metadata + distinct

	// time format required for single day range, %20 between date and time
	request += "/EPOCH/" + date + " 00:00:00--" + date + " 23:59:59"

	// other additional options
	// request += "/decay_date/null-val"
	// request += "/PERIAPSIS/" + "<600"

	// order by multiple things
	request += "/orderby" + norad + " asc"
	request += ",EPOCH asc"

	request += json_format

	fmt.Print("\nAPI Request: ")
	fmt.Println(request)
	return request
}

func parse_credential_file(file string) Credentials {
	data, err := os.ReadFile(file)
	check(err)

	user_creds := Credentials{"", ""}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		items := strings.Split(line, " ")
		switch items[0] {
		case "username":
			user_creds.username = items[1]
		case "password":
			user_creds.password = items[1]
		default:
			fmt.Println("Something wrong in the creds file?")
		}

	}
	return user_creds
}

func read_credentials() Credentials {
	user_creds := parse_credential_file(creds_file)
	return user_creds
}

func Setup_session() (client *http.Client, err error) {
	login_addr := base_url + login

	jar, err := cookiejar.New(nil)
	check(err)
	client = &http.Client{
		Jar: jar,
	}

	creds := read_credentials()
	// fmt.Println(creds)

	request := map[string]string{"identity": creds.username, "password": creds.password}
	jsonVal, err := json.Marshal(request)

	body := []byte(jsonVal)
	result, err := client.Post(login_addr, "application/json", bytes.NewBuffer(body))
	check(err)
	check_response(result)

	return client, nil
}

func End_session(client *http.Client) (err error) {
	fmt.Println("Logging out...")
	logout_addr := base_url + logout

	resp, err := client.Get(logout_addr)
	if err != nil {
		return err
	}
	check_response(resp)
	return nil
}

func Get_data(client *http.Client, date string) (response *http.Response, err error) {
	date = "2025-01-05"
	request := build_request(date)

	resp, err := client.Get(request)
	check(err)
	// defer resp.Body.Close()

	check_response(resp)

	// io.Copy from Reader to Writer
	// file, err := os.Create("/home/tom/dev/go/ltle/testfile")
	// check(err)
	// written, err := io.Copy(file, resp.Body)
	// check(err)
	// log.Printf("File had %d bytes written", written)

	// fmt.Print("Get Response: ")
	// fmt.Println(resp.Status)
	// fmt.Println()

	return resp, nil
}

func check_response(response *http.Response) {
	if response.StatusCode != 200 {
		log.Printf("Incorrect code received in response: %s", response.Status)
	}
}
