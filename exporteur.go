package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
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
		fmt.Println("Unexpected error reading username.")
		fmt.Println(err)
		os.Exit(1)
	}

	// Get the password
	fmt.Print("Password: ")
	passwordinput, err := terminal.ReadPassword(0)
	if err != nil {
		fmt.Println("Unexpected error reading password.")
		fmt.Println(err)
		os.Exit(1)
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
			fmt.Println("Unexpected error logging in.")
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()

		if resp.Request.URL.Path == "/" {
			break
		} else {
			fmt.Println("Error logging in, retry...")
		}
	}

}

func exportData(irtype string, client *http.Client) {

	path := fmt.Sprintf("/%v/export", irtype)
	var curExport = &export{}

	for !curExport.Finished {

		url := fmt.Sprintf("https://ir.library.carleton.ca%v?job=%v", path, curExport.Job)

		resp, err := client.Get(url)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if resp.StatusCode != 200 || resp.Request.URL.Path != path {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			fmt.Println(buf.String())
			resp.Body.Close()
			os.Exit(1)
		}

		err = json.NewDecoder(resp.Body).Decode(curExport)
		resp.Body.Close()

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("%+v\n", curExport)
		time.Sleep(1 * time.Second)
	}

	resp, err := client.Get(curExport.URL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	filename := fmt.Sprintf("%v-%v", time.Now().Format("2006-01-02"), params["filename"])

	fmt.Printf("Saving to filename %v\n", filename)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	file.Close()

}

func main() {

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		fmt.Println("Unexpected error creating cookie storage.")
		fmt.Println(err)
		os.Exit(1)
	}

	client := &http.Client{
		Jar: jar,
	}

	login(client)

	fmt.Println("Login complete!")

	exportData("col", client)

}
