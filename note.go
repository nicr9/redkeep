package main

import (
	"fmt"
	"gopkg.in/redis.v3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"time"
)

type Note struct {
	Id      int64    `yaml:"id"`
	Title   string   `yaml:"title"`
	Created int64    `yaml:"created"`
	Updated int64    `yaml:"updated"`
	Tags    []string `yaml:"tags"`
	Open    []string `yaml:"open"`
	Closed  []string `yaml:"closed"`
	Body    string   `yaml:"body"`
}

func (n *Note) validate(client redis.Client) error {
	// Initialise timestamps
	now := time.Now().Unix()
	if n.Created == 0 {
		n.Created = now
	}
	if n.Updated == 0 {
		n.Updated = now
	}

	if n.Id == 0 {
		n.Id = client.Incr("redkeep:id-counter").Val()
	}
	return nil
}

func ToTempFile(notes []Note) string {
	// Marshall notes
	data, err := yaml.Marshal(&notes)
	if err != nil {
		log.Fatal(err)
	}

	// Open temp file
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		log.Fatal(err)
	}

	// Write data
	if _, err := tmpfile.Write(data); err != nil {
		log.Fatal(err)
	}

	// Close file
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	return tmpfile.Name()
}
