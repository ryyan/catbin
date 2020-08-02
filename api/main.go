package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// Directory that texts will be saved to
	TextDir = ".text"

	// Expiration times and format
	Day        = time.Hour * 24
	Week       = Day * 7
	Month      = Day * 31
	Year       = Day * 365
	DateFormat = time.RFC3339
)

var (
	expirations = []string{"hour", "day", "week", "month", "year"}
)

type response struct {
	Expiration string
	Text       string
}

func main() {
	// Create texts directory
	os.Mkdir(TextDir, 0755)

	// Set endpoints
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

	// read text file
	data, err := ioutil.ReadFile(TextDir + "/" + id)
	if err != nil {
		return "", errors.New("Text not found")
	}

	// parse file
	dataArr := strings.Split(string(data), "\n")
	result, err := json.Marshal(&response{
		Expiration: dataArr[0],
		Text:       dataArr[1],
	})
	if err != nil {
		log.Fatalf("Failed to marshal json: %s", err)
	}

	return string(result), nil
}

func saveText(text string, expiration string) (string, error) {
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
	expirationStr := expirationDate.Format(DateFormat)

	// generate ID
	id := generateId(36)

	// create text file
	file, err := os.Create(TextDir + "/" + id)
	if err != nil {
		log.Fatalf("Failed creating file: %s", err)
	}
	defer file.Close()

	// save text file
	_, err = file.WriteString(expirationStr + "\n" + text)
	if err != nil {
		log.Fatalf("Failed writing to file: %s", err)
	}

	return id, nil
}

// stringInSlice returns true if the given string is in the list
func stringInSlice(toFind string, list []string) bool {
	for _, val := range list {
		if val == toFind {
			return true
		}
	}
	return false
}

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var (
	random = rand.NewSource(time.Now().UTC().UnixNano())
)

// generateId generates a random string
// https://stackoverflow.com/questions/22892120
func generateId(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	for i, cache, remain := n-1, random.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = random.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}
