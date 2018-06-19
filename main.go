package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

const (
	roleLeader   = "leader"
	roleFollower = "follower"
)

var (
	role    string
	datadir string
)

func main() {
	host := "localhost"
	if h, err := os.Hostname(); err == nil {
		host = h
		if ho := os.Getenv("MEHDB_HOST"); ho != "" {
			host = ho
		}
	}
	port := "9876"
	if p := os.Getenv("MEHDB_PORT"); p != "" {
		port = p
	}
	cwd, _ := os.Getwd()
	datadir = filepath.Join(cwd, "data")
	if d := os.Getenv("MEHDB_DATADIR"); d != "" {
		datadir = d
	}
	if _, err := os.Stat(datadir); os.IsNotExist(err) {
		os.Mkdir(datadir, os.ModePerm)
	}
	log.Printf("mehdb serving from %v:%v using %v as the data directory", host, port, datadir)
	role = discover(host)
	go syncdata()
	r := mux.NewRouter()
	r.HandleFunc("/set/{key:[a-z]+}", writedata).Methods("PUT")
	r.HandleFunc("/get/{key:[a-z]+}", readdata).Methods("GET")
	http.Handle("/", r)
	srv := &http.Server{Handler: r, Addr: "0.0.0.0:" + port}
	log.Fatal(srv.ListenAndServe())
}

func discover(host string) string {
	switch host {
	case "meh-shard-0":
		log.Printf("I am the leading shard, accepting both WRITES and READS")
		return roleLeader
	default:
		log.Printf("I am a follower shard, accepting READS")
		return roleFollower
	}
}

func syncdata() {
	if role == roleLeader {
		return
	}
	for {
		log.Printf("Checking for new data from leader")
		time.Sleep(5 * time.Second)
	}
}

func writedata(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	defer r.Body.Close()
	c, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Can't parse key %s due to %v", key, err)
		return
	}
	keydir := filepath.Join(datadir, key)
	if _, err = os.Stat(keydir); os.IsNotExist(err) {
		os.Mkdir(keydir, os.ModePerm)
	}
	err = ioutil.WriteFile(filepath.Join(keydir, "content"), c, 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Can't write key %s due to %v", key, err)
		return
	}
	fmt.Fprint(w, "WRITE completed")
}

func readdata(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	c, err := ioutil.ReadFile(filepath.Join(datadir, key, "content"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Can't read key %s due to %v", key, err)
		return
	}
	fmt.Fprint(w, string(c))
}
