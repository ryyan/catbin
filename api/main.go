package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// Port to run the server on
	Port = ":5000"

	// TextDir is the directory that texts will be saved to
	// The filename will be the text ID, first line the expiration, second line the text itself
	TextDir = ".text"

	// Expiration times and format
	Day        = time.Hour * 24
	Week       = Day * 7
	Month      = Day * 31
	Year       = Day * 365
	DateFormat = time.RFC3339
)

var (
	// expirations holds the possible expiration (enum) values
	expirations = []string{"hour", "day", "week", "month", "year"}

	// textCache holds the map of texts, key=ID, value=expirationDate
	// This way we don't have to read all the text files over and over to find those expired
	textCache = make(map[string]time.Time)
)

// response is the json type returned when getting a text
type response struct {
	Expiration string `json:"expiration"`
	Text       string `json:"text"`
}

func main() {
	// Create text directory
	os.Mkdir(TextDir, 0755)

	// Scan text directory, clean old texts, and cache current
	initialScan()

	// Start scanner goroutine
	go cleaner()

	// Set endpoints
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/msg", handler)
	serveMux.HandleFunc("/msg/", handler)
	serveMux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("../web/dist"))))

	// Start server
	log.Printf("Server started at localhost%s\n", Port)
	log.Fatal(http.ListenAndServe(Port, serveMux))
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
		return "", errors.New("Missing field: id")
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
	check(err)

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
	id := generateId(33)

	// create text file
	file, err := os.Create(TextDir + "/" + id)
	check(err)
	defer file.Close()

	// save text file
	_, err = file.WriteString(expirationStr + "\n" + text)
	check(err)

	// save text to cache
	textCache[id] = expirationDate

	return id, nil
}

func initialScan() {
	deleted := 0

	// get current datetime
	now := time.Now().UTC()

	// get path to text directory
	path, err := os.Getwd()
	check(err)
	path = path + "/" + TextDir
	log.Printf("Scanning %s\n", path)

	// iterate over files in text directory
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		check(err)
		if info.Name() == TextDir {
			return nil
		}

		// the filename is the text ID
		id := info.Name()

		// read the first line of the file to get the expiration date
		file, err := os.Open(path)
		check(err)
		defer file.Close()
		scanner := bufio.NewScanner(file)
		scanner.Scan()
		expiration := scanner.Text()
		err = scanner.Err()
		check(err)
		expirationDate, err := time.Parse(DateFormat, expiration)
		check(err)

		// delete text if expired
		if expirationDate.Before(now) {
			err := os.Remove(path)
			check(err)
			deleted += 1
			return nil
		}

		// save text to cache
		textCache[id] = expirationDate

		return nil
	})

	log.Printf("Deleted %d\n", deleted)
	log.Printf("Cached %d\n", len(textCache))
}

func cleaner() {
	// infinite loop since this is a background job
	for {
		// sleep between scans
		time.Sleep(time.Minute * 11)

		deleted := 0

		// get current datetime
		now := time.Now().UTC()

		// get path to text directory
		path, err := os.Getwd()
		check(err)
		path = path + "/" + TextDir

		// iterate over text cache and delete expired texts
		for id, expirationDate := range textCache {
			if expirationDate.Before(now) {
				err := os.Remove(path + "/" + id)
				check(err)
				delete(textCache, id)
				deleted += 1
			}
		}

		if deleted > 0 {
			log.Printf("Cleaned %d\n", deleted)
		}
	}
}

// check reduces the amount of if err != nil spam
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
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
