package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	Day = time.Hour * 24
	Week = Day * 7
	Month = Day * 31
	Year = Day * 365
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
	// validate
	if id == "" {
		return "", errors.New("Missing id")
	}

	log.Println(id)
	return "OK", nil
}

func saveText(text string, expiration string) (result string, err error) {
	// validate
	if text == "" {
		return "", errors.New("Missing field: text")
	}
	if !stringInSlice(expiration, expirations) {
		return "", errors.New("Missing or invalid field: expiration")
	}

	// calculate expiration date
	expirationDate := time.Now().UTC()
	switch expiration {
	case "hour":
		expirationDate = expirationDate.Add(time.Hour)
	case "day":
		expirationDate = expirationDate.Add(Day)
	case "week":
		expirationDate = expirationDate.Add(Week)
	case "month":
		expirationDate = expirationDate.Add(Month)
	case "year":
		expirationDate = expirationDate.Add(Year)
	}

	log.Println(expirationDate)

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
