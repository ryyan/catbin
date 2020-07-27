package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	expirations = []string{"hour", "day", "week", "month", "year"}
)

func main() {
	// Create directory that messages will be saved to
	os.Mkdir(".files", 0755)

	// Set api endpoints
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/msg", handler)
	serveMux.HandleFunc("/msg/", handler)
	//serveMux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("../web/dist"))))

	// Start server
	log.Println("Serving http at localhost:5000")
	log.Fatal(http.ListenAndServe(":5000", serveMux))
}

func handler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		id := strings.TrimLeft(req.URL.Path, "msg/")
		if id == "" {
			writeError(res, "Missing id")
			return
		}
		
		log.Println(id)

	case "POST":
		req.ParseForm()
		text := req.FormValue("text")
		expiration := req.FormValue("expiration")
		if text == "" {
			writeError(res, "Missing field: text")
			return
		}
		if !stringInSlice(expiration, expirations) {
			writeError(res, "Missing or invalid field: expiration")
			return
		}
	}
}

func stringInSlice(toFind string, list []string) bool {
	for _, val := range list {
		if val == toFind {
			return true
		}
	}
	return false
}

func write(res http.ResponseWriter, msg string) {
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, msg)
}

func writeError(res http.ResponseWriter, err string) {
	res.WriteHeader(http.StatusBadRequest)
	io.WriteString(res, err)
}
