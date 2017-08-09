package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/publicsuffix"
)

type export struct {
	URL       string `json:"url"`
	Job       string `json:"job"`
	Finished  bool   `json:"finished"`
	Processed int    `json:"processed"`
}

func prompt() (username, password string) {
	// Get the username
	fmt.Print("Username: ")
	reader := bufio.NewReader(os.Stdin)
	usernameinput, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalln("Error reading username.", err)
	}

	// Get the password
	fmt.Print("Password: ")
	passwordinput, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalln("Error reading password.", err)
	}
	fmt.Println()

	return strings.TrimSpace(usernameinput), string(passwordinput)
}

func login(client *http.Client) {
	for {
		username, password := prompt()

		resp, err := client.PostForm("https://ir.library.carleton.ca/login",
			url.Values{"userid": {username}, "password": {password}})
		if err != nil {
			log.Fatalln("Error logging in.", err)
		}
		resp.Body.Close()

		if resp.Request.URL.Path == "/" {
			break
		} else {
			fmt.Println("Username or password incorrect, please retry.")
		}
	}

}

func exportData(irtype string, client *http.Client, waitgroup *sync.WaitGroup) {
	defer waitgroup.Done()

	path := fmt.Sprintf("/%v/export", irtype)
	var curExport = &export{}

	for !curExport.Finished {

		url := fmt.Sprintf("https://ir.library.carleton.ca%v?job=%v", path, curExport.Job)

		resp, err := client.Get(url)
		if err != nil {
			log.Fatalln("Error GETing at URL ", url, err)
		}

		if resp.StatusCode != 200 || resp.Request.URL.Path != path {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			resp.Body.Close()
			log.Fatalln(buf.String())
		}

		err = json.NewDecoder(resp.Body).Decode(curExport)
		resp.Body.Close()
		if err != nil {
			log.Fatalln("Unable to parse JSON response.", err)
		}

		log.Printf("%+v\n", curExport)
		time.Sleep(1 * time.Second)
	}

	resp, err := client.Get(curExport.URL)
	if err != nil {
		log.Fatalln("Error GETing at URL ", curExport.URL, err)
	}

	mediaTypeFromHeader := resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(mediaTypeFromHeader)
	if err != nil {
		log.Fatalln("Error parsing media type.", mediaTypeFromHeader, err)
	}

	filename := fmt.Sprintf("%v-%v", time.Now().Format("2006-01-02"), params["filename"])

	log.Println("Deleting old file if it exists ", filename)

	err = os.Remove(filename)
	if err != nil {
		log.Println("Could not delete existing file ", filename, err)
	}

	log.Println("Saving to filename ", filename)

	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		log.Fatalln("Could not create file ", filename, err)
	}

	err = file.Chmod(0600)
	if err != nil {
		log.Println("Could not change permissions on file ", filename, err)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatalln("Could not copy contents into file ", filename, err)
	}

	err = file.Chmod(0400)
	if err != nil {
		log.Println("Could not change permissions on file ", filename, err)
	}
}

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 {
		log.Fatalln("Please provide at least one short code for a type to export, like col or pub.")
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatalln("Could not create cookie storage.", err)
	}
	client := &http.Client{
		Jar: jar,
	}

	login(client)
	log.Println("Login Complete!")

	var waitgroup sync.WaitGroup
	for _, shortcode := range flag.Args() {
		waitgroup.Add(1)
		go exportData(shortcode, client, &waitgroup)
	}
	waitgroup.Wait()
}
