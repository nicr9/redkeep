package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	redkeep "github.com/nicr9/redkeep/pkg"
	"gopkg.in/redis.v5"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

func passClient(client *redis.Client, lambda func(*redis.Client, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		lambda(client, w, r)
	}
}

func Dashboard(client *redis.Client, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if message, err := client.Get("redkeep:welcomemessage").Result(); err == nil {
		fmt.Fprintln(w, message)
	} else {
		fmt.Fprintf(w, "Can't find welcome message: %s", err)
	}
}

func Keys(client *redis.Client, w http.ResponseWriter, r *http.Request) {
	if keys, err := client.Keys("*").Result(); err == nil {
		w.WriteHeader(http.StatusOK)
		for _, message := range keys {
			fmt.Fprintln(w, message)
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Can't find welcome message: %s", err)
	}
}

func ApiSaveNotes(client *redis.Client, w http.ResponseWriter, r *http.Request) {
	var notes *[]redkeep.Note
	b, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(b, &notes)

	log.Printf("Saving '%s'...\n", n.Title)
	if err := redkeep.ToRedis(notes, client); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Failed to save notes to Redis: %v", notes)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func main() {
	log.Println("Starting up RedKeep...")

	// Validate environment variables
	dbStr := os.Getenv("REDIS_DB")
	var dbInt int64
	var err error
	if dbInt, err = strconv.ParseInt(dbStr, 10, 0); err != nil {
		log.Printf("REDIS_DB should be an int, got '%v' instead\n", dbStr)
		os.Exit(1)
	}

	var host string = os.Getenv("REDIS_HOSTPORT")
	log.Printf("Connecting to host %s, to db %d", host, dbInt)

	// Connect to Redis
	var client *redis.Client = redis.NewClient(&redis.Options{
		Addr:     host,
		Password: os.Getenv("REDIS_PASSWD"),
		DB:       int(dbInt),
	})
	client.Wait(1, 15*1000) // Wait at least 15 seconds for redis to respond

	log.Println("Ready to test connection...")
	if pong, err := client.Ping().Result(); err != nil {
		log.Println("Redis ping failed:", pong, err)
	} else {
		log.Println("Connection looks good!")
	}

	// id-counter is used to get unique ids for new notes
	client.SetNX("redkeep:id-counter", 0, 0)
	client.Set("redkeep:welcomemessage", "Hello World!", 0)

	// Create router and register endpoints
	r := mux.NewRouter()
	r.HandleFunc("/", passClient(client, Dashboard))
	r.HandleFunc("/keys", passClient(client, Keys))
	r.HandleFunc("/api/v1/save", passClient(client, ApiSaveNotes)).
		Methods("POST")

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":80", r))
}
