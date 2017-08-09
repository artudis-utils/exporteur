package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"time"
)

type export struct {
	URL       string `json:"url"`
	Job       string `json:"job"`
	Finished  bool   `json:"finished"`
	Processed int    `json:"processed"`
}

func uberHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "login":
		http.Redirect(w, r, "/", http.StatusFound)
	case path.Base(r.URL.Path) == "export":
		curExport := export{}
		curExport.Job = r.URL.Query().Get("job")
		jobInteger, _ := strconv.ParseInt(curExport.Job, 10, 64)
		if curExport.Job == "" {
			curExport.Job = strconv.FormatInt(time.Now().Unix(), 10)
		} else if (time.Now().Unix() - jobInteger) > 10 {
			curExport.Finished = true
			curExport.URL = "localhost:8080"
		} else {
			curExport.Processed = int(time.Now().Unix() - jobInteger)
		}
		returnMe, err := json.Marshal(curExport)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(returnMe)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "inline; filename=\"test.json\"")
		fmt.Fprint(w, `{"Hello":"World"}`)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusOK)
}

func main() {
	http.HandleFunc("/", uberHandler)
	http.ListenAndServe(":8080", nil)
}
