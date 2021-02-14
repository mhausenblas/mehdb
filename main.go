package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	leaderShard  = "mehdb-0"
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
		errd := os.Mkdir(datadir, os.ModePerm)
		if errd != nil {
			log.Printf("Can't create data dir due to %v", err)
		}
	}
	log.Printf("mehdb serving from %v:%v using %v as the data directory", host, port, datadir)
	role = discover(host)
	go syncdata(port)
	r := mux.NewRouter()
	r.HandleFunc("/set/{key:[a-z]+}", writedata).Methods("PUT")
	r.HandleFunc("/get/{key:[a-z]+}", readdata).Methods("GET")
	r.HandleFunc("/keys", listkeys).Methods("GET")
	r.HandleFunc("/status", status).Methods("GET")
	http.Handle("/", r)
	srv := &http.Server{Handler: r, Addr: "0.0.0.0:" + port}
	log.Fatal(srv.ListenAndServe())
}

func discover(host string) string {
	switch host {
	case leaderShard:
		log.Printf("I am the leading shard, accepting both WRITES and READS")
		return roleLeader
	default:
		log.Printf("I am a follower shard, accepting READS")
		return roleFollower
	}
}

func syncdata(port string) {
	if role == roleLeader {
		return
	}
	client := &http.Client{Timeout: 5 * time.Second}
	ns := currentns()
	url := "http://" + leaderShard + "." + ns + ":" + port + "/keys"
	if local := os.Getenv("MEHDB_LOCAL"); local != "" {
		url = "http://localhost:9999/keys"
	}
	for {
		time.Sleep(10 * time.Second)
		log.Printf("Checking for new data from leader")
		// list keys
		r, err := client.Get(url)
		if err != nil {
			log.Printf("Can't get keys from leader due to %v", err)
			continue
		}
		keys := []string{}
		err = json.NewDecoder(r.Body).Decode(&keys)
		if err != nil {
			log.Printf("Can't decode keys due to %v", err)
		}
		_ = r.Body.Close()
		// for each key, get the data
		for _, k := range keys {
			keydir := filepath.Join(datadir, k)
			if _, err = os.Stat(keydir); os.IsNotExist(err) {
				_ = os.Mkdir(keydir, os.ModePerm)
			}
			kurl := "http://" + leaderShard + "." + ns + ":" + port + "/get/" + k
			if local := os.Getenv("MEHDB_LOCAL"); local != "" {
				kurl = "http://localhost:9999/get/" + k
			}
			r, err := client.Get(kurl)
			if err != nil {
				log.Printf("Can't get key %v from leader due to %v", k, err)
				continue
			}
			c, err := ioutil.ReadAll(r.Body)
			err = ioutil.WriteFile(filepath.Join(keydir, "content"), c, 0644)
			if err != nil {
				log.Printf("Can't sync key %v from leader due to %v", k, err)
				continue
			}
			log.Printf("Synced key %v from leader", k)
		}
	}
}

func writedata(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	if role == roleFollower { // redirect WRITES to leader
		port := "9876"
		if p := os.Getenv("MEHDB_PORT"); p != "" {
			port = p
		}
		// assemble leader FQDN based on pod DNS entry:
		lurl := "http://" + leaderShard + "." + currentns() + ":" + port + "/set/" + key
		if local := os.Getenv("MEHDB_LOCAL"); local != "" {
			lurl = "http://localhost:9999/set/" + key
		}
		log.Printf("Redirecting WRITE to %v", lurl)
		http.Redirect(w, r, lurl, 307)
		return
	}
	defer r.Body.Close()
	c, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Can't parse key %s due to %v", key, err)
		return
	}
	keydir := filepath.Join(datadir, key)
	if _, err = os.Stat(keydir); os.IsNotExist(err) {
		_ = os.Mkdir(keydir, os.ModePerm)
	}
	err = ioutil.WriteFile(filepath.Join(keydir, "content"), c, 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Can't write key %s due to %v", key, err)
		return
	}
	log.Printf("Done writing key %s", key)
	fmt.Fprint(w, "WRITE completed")
}

func readdata(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	c, err := ioutil.ReadFile(filepath.Join(datadir, key, "content"))
	if err != nil {
		http.Error(w, "", http.StatusNotFound)
		log.Printf("Can't read key %s due to %v", key, err)
		return
	}
	log.Printf("Done reading key %s", key)
	fmt.Fprint(w, string(c))
}

func listkeys(w http.ResponseWriter, r *http.Request) {
	keys, err := ioutil.ReadDir(datadir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Can't list keys due to %v", err)
		return
	}
	klist := []string{}
	for _, k := range keys {
		key := k.Name()
		if !strings.HasPrefix(key, ".") {
			if _, err := os.Stat(filepath.Join(datadir, key, "content")); err == nil {
				klist = append(klist, k.Name())
			}
		}
	}
	_ = json.NewEncoder(w).Encode(klist)
}

func status(w http.ResponseWriter, r *http.Request) {
	level := r.URL.Query().Get("level")
	switch level {
	case "full":
		keys, err := ioutil.ReadDir(datadir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		knum := 0
		for _, k := range keys {
			key := k.Name()
			if !strings.HasPrefix(key, ".") {
				if _, err := os.Stat(filepath.Join(datadir, key, "content")); err == nil {
					knum++
				}
			}
		}
		_ = json.NewEncoder(w).Encode(knum)
		return
	default:
		fmt.Fprint(w, role)
		return
	}
}

func currentns() string {
	ns := "default"
	if n := os.Getenv("MEHDB_NAMESPACE"); n != "" {
		ns = n
	}
	return ns
}
