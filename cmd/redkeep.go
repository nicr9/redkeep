package main

import (
	"fmt"
	redkeep "github.com/nicr9/redkeep/pkg"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"
)

var usage string = "RedKeep Usage: redkeep <command> [<args>]"
var redkeepHost *url.URL

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

func saveNotes(notes *[]redkeep.Note, client *http.Client) {
	endpoint, err := url.Parse("/api/v1/save")
	if err != nil {
		log.Fatal(err)
	} else {
		if buff, err := redkeep.ToJsonReader(notes); err == nil {
			target := redkeepHost.ResolveReference(endpoint).String()
			response, err := client.Post(target, "application/json", buff)
			if err != nil {
				log.Fatal(err)
			} else if response.StatusCode < 400 {
				log.Fatal(ioutil.ReadAll(response.Body))
			}
		} else {
			log.Fatal(err)
		}
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println(usage)
		os.Exit(2)
	}

	var err error
	redkeepHost, err = url.Parse(os.Getenv("REDKEEP_HOST"))
	if err != nil {
		log.Fatal(err)
	}

	var client = &http.Client{
		Timeout: time.Second * 10,
	}

	switch os.Args[1] {
	case "new":
		note := redkeep.Note{}
		notes := &[]redkeep.Note{note}
		edited := EditNotes(notes)
		saveNotes(edited, client)

		//	case "list-tags":
		//		keys := client.Keys("redkeep:tags:*").Val()
		//		var tag string
		//		for i, key := range keys {
		//			tag = strings.TrimLeft(key, "redkeep:tags:")
		//			fmt.Printf("%d) %s\n", i, tag)
		//		}
		//
		//	case "search-tags":
		//		tags := strings.Split(os.Args[2], ",")
		//		for i, tag := range tags {
		//			tags[i] = fmt.Sprintf("redkeep:tags:%s", tag)
		//		}
		//		noteIds := client.SDiff(tags...).Val()
		//		notes, err := redkeep.FromRedis(noteIds, client)
		//		if err != nil {
		//			log.Fatalf("error: %v", err)
		//		}
		//		edited := EditNotes(&notes)
		//		saveNotes(edited, client)
		//
		//	case "delete":
		//		ids := os.Args[2:]
		//		redkeep.DeleteById(client, ids...)

	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}
}
