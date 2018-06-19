package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	host := "localhost"
	if h, err := os.Hostname(); err == nil {
		host = h
	}
	port := "9876"
	if p := os.Getenv("MEHDB_PORT"); p != "" {
		port = p
	}
	r := mux.NewRouter()
	r.HandleFunc("/set", writedata).Methods("POST")
	r.HandleFunc("/get/{key:[a-z]+}", readdata).Methods("GET")
	log.Printf("mehdb serving from: %s:%s/", host, port)
	http.Handle("/", r)
	srv := &http.Server{Handler: r, Addr: "0.0.0.0:" + port}
	log.Fatal(srv.ListenAndServe())
}

func writedata(w http.ResponseWriter, r *http.Request) {
}

func readdata(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	c, err := ioutil.ReadFile(key + ".zip")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Can't read data from key %s due to %v", key, err)
		return
	}
	fmt.Fprint(w, string(c))
}
