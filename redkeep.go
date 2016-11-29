package main

import (
	"fmt"
	"gopkg.in/redis.v3"
	"log"
	"os"
	"os/exec"
	"strings"
)

var usage string = "Usage: redkeep <command> [<args>]"

var client *redis.Client = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
})

func OpenEditor(filepath string) {
	cmd := exec.Command("vim", filepath, "+set filetype=yaml")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}

func EditNotes(notes []Note) {
	filepath := ToTempFile(notes)
	defer os.Remove(filepath)
	OpenEditor(filepath)
	var err error
	notes, err = FromFile(filepath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if err := ValidateNotes(notes, *client); err != nil {
		log.Fatalf("validation error: %s", err)
	}
	for i := range notes {
		fmt.Printf("Saving '%s'...\n", notes[i].Title)
		notes[i].toRedis(*client)
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println(usage)
		os.Exit(2)
	}

	client.SetNX("redkeep:id-counter", 0, 0)

	switch os.Args[1] {
	case "new":
		note := Note{}
		notes := []Note{note}
		EditNotes(notes)

	case "list-tags":
		keys := client.Keys("redkeep:tags:*").Val()
		var tag string
		for i, key := range keys {
			tag = strings.TrimLeft(key, "redkeep:tags:")
			fmt.Printf("%d) %s\n", i, tag)
		}
	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}
}
