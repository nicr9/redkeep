package main

import (
	"fmt"
	"gopkg.in/redis.v3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

type Note struct {
	Id      string   `yaml:"id"`
	Title   string   `yaml:"title"`
	Created string   `yaml:"created"`
	Updated string   `yaml:"updated"`
	Tags    []string `yaml:"tags"`
	Open    []string `yaml:"open"`
	Closed  []string `yaml:"closed"`
	Body    string   `yaml:"body"`
}

func (n *Note) validate(client redis.Client) error {
	// Initialise timestamps
	now := strconv.FormatInt(time.Now().Unix(), 10)
	if n.Created == "" {
		n.Created = now
	}
	n.Updated = now

	if n.Id == "" {
		n.Id = strconv.FormatInt(client.Incr("redkeep:id-counter").Val(), 10)
	}
	return nil
}

func (n *Note) toRedis(client redis.Client) error {
	// Update timestamp
	now := time.Now().Unix()
	n.Updated = now

	key := fmt.Sprintf("redkeep:note:%d", n.Id)

	client.Set(fmt.Sprintf("%s:title", key), n.Title, 0)
	client.Set(fmt.Sprintf("%s:created", key), n.Created, 0)
	client.Set(fmt.Sprintf("%s:updated", key), n.Updated, 0)
	client.Set(fmt.Sprintf("%s:body", key), n.Body, 0)

	n.updateTags(client)

	return nil
}

func (n *Note) updateTags(client redis.Client) {
	key := fmt.Sprintf("redkeep:note:%d", n.Id)

	tags_key := fmt.Sprintf("%s:tags", key)
	temp_key := fmt.Sprintf("%s:temp", key)
	remove_key := fmt.Sprintf("%s:remove", key)
	update_key := fmt.Sprintf("%s:update", key)

	client.SAdd(temp_key, n.Tags...)

	client.SDiffStore(remove_key, tags_key, temp_key)
	client.SDiffStore(update_key, temp_key, tags_key)

	// for tag in remove_key
	for _, tag := range client.SMembers(remove_key).Val() {
		client.SRem(fmt.Sprintf("redkeep:tags:%s", tag), fmt.Sprintf("%s", n.Id))
	}

	// for tag in update_key
	for _, tag := range client.SMembers(update_key).Val() {
		client.SAdd(fmt.Sprintf("redkeep:tags:%s", tag), fmt.Sprintf("%s", n.Id))
	}

	client.Del(tags_key)
	client.Rename(temp_key, tags_key)

	client.Del(remove_key)
	client.Del(update_key)
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

func FromFile(filepath string) ([]Note, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	notes := []Note{}
	err = yaml.Unmarshal(data, &notes)
	if err != nil {
		return nil, err
	}

	return notes, nil
}

func ValidateNotes(notes []Note, client redis.Client) error {
	for i := range notes {
		notes[i].validate(client)
	}

	return nil
}
