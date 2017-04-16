package main

import (
	"fmt"
	redkeep "github.com/nicr9/redkeep/pkg"
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

func EditNotes(notes *[]redkeep.Note) (results *[]redkeep.Note) {
	filepath := redkeep.ToTempFile(*notes)
	defer os.Remove(filepath)
	OpenEditor(filepath)
	var err error
	results, err = redkeep.FromFile(filepath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return results
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println(usage)
		os.Exit(2)
	}

	client.SetNX("redkeep:id-counter", 0, 0)

	switch os.Args[1] {
	case "new":
		note := redkeep.Note{}
		notes := &[]redkeep.Note{note}
		edited := EditNotes(notes)
		redkeep.ToRedis(edited, client)

	case "list-tags":
		keys := client.Keys("redkeep:tags:*").Val()
		var tag string
		for i, key := range keys {
			tag = strings.TrimLeft(key, "redkeep:tags:")
			fmt.Printf("%d) %s\n", i, tag)
		}

	case "search-tags":
		tags := strings.Split(os.Args[2], ",")
		for i, tag := range tags {
			tags[i] = fmt.Sprintf("redkeep:tags:%s", tag)
		}
		noteIds := client.SDiff(tags...).Val()
		notes, err := redkeep.FromRedis(noteIds, client)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		edited := EditNotes(&notes)
		redkeep.ToRedis(edited, client)

	case "delete":
		ids := os.Args[2:]
		redkeep.DeleteById(client, ids...)

	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}
}
