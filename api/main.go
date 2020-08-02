package main

import (
	"io"
	"errors"
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
	serveMux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("../web/dist"))))

	// Start server
	log.Println("Serving http at localhost:5000")
	log.Fatal(http.ListenAndServe(":5000", serveMux))
}

func handler(res http.ResponseWriter, req *http.Request) {
	var result string
	var err error

	switch req.Method {
	case "GET":
		id := strings.TrimLeft(req.URL.Path, "msg/")
		result, err = getText(id)

	case "POST":
		req.ParseForm()
		text := req.FormValue("text")
		expiration := req.FormValue("expiration")
		result, err = saveText(text, expiration)
	}

	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		io.WriteString(res, err.Error())
	} else {
		io.WriteString(res, result)
	}
}

func getText(id string) (string, error) {
	if id == "" {
		return "", errors.New("Missing id")
	}

	log.Println(id)
	return "OK", nil
}

func saveText(text string, expiration string) (result string, err error) {
	if text == "" {
		return "", errors.New("Missing field: text")
	}
	if !stringInSlice(expiration, expirations) {
		return "", errors.New("Missing or invalid field: expiration")
	}

	return "OK", nil
}

func stringInSlice(toFind string, list []string) bool {
	for _, val := range list {
		if val == toFind {
			return true
		}
	}
	return false
}
