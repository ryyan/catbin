package main

import (
	"log"
	"net/http"
	"strings"
)

func main() {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/msg", handler)
	serveMux.HandleFunc("/msg/", handler)
	//serveMux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("../web/dist"))))
	log.Println("Serving http at localhost:5000")
	log.Fatal(http.ListenAndServe(":5000", serveMux))
}

func handler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {

	case "GET":
		id := strings.TrimLeft(req.URL.Path, "msg/")
		log.Println(id)

	case "POST":
		req.ParseForm()
		log.Println(req.FormValue("text"))
		log.Println(req.FormValue("expiration"))
	}
}
