package redkeep

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/redis.v5"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

type Note struct {
	Id      string   `yaml:"id"      json:"id"`
	Title   string   `yaml:"title"   json:"title"`
	Created string   `yaml:"created" json:"created"`
	Updated string   `yaml:"updated" json:"updated"`
	Tags    []string `yaml:"tags"    json:"tags"`
	Open    []string `yaml:"open"    json:"open"`
	Closed  []string `yaml:"closed"  json:"closed"`
	Body    string   `yaml:"body"    json:"body"`
}

func (n *Note) validate(client *redis.Client) error {
	// Initialise timestamps
	now := strconv.FormatInt(time.Now().Unix(), 10)
	if n.Created == "" {
		n.Created = now
	}
	n.Updated = now

	if n.Id == "" {
		n.Id = strconv.FormatInt(client.Incr("redkeep:id-counter").Val(), 10)
	} else if _, err := strconv.Atoi(n.Id); err != nil {
		return fmt.Errorf("Note ID is corrupt: %+v", n)
	}

	return nil
}

func FromRedis(noteIds []string, client *redis.Client) (notes []Note, err error) {
	for _, id := range noteIds {
		title := client.Get(fmt.Sprintf("redkeep:note:%s:title", id)).Val()
		created := client.Get(fmt.Sprintf("redkeep:note:%s:created", id)).Val()
		updated := client.Get(fmt.Sprintf("redkeep:note:%s:updated", id)).Val()
		body := client.Get(fmt.Sprintf("redkeep:note:%s:body", id)).Val()

		tags := client.SMembers(fmt.Sprintf("redkeep:note:%s:tags", id)).Val()
		open := client.LRange(fmt.Sprintf("redkeep:note:%s:open", id), 0, -1).Val()
		closed := client.LRange(fmt.Sprintf("redkeep:note:%s:closed", id), 0, -1).Val()

		obj := Note{id, title, created, updated, tags, open, closed, body}
		notes = append(notes, obj)
	}

	return notes, nil
}

func ToRedis(notes *[]Note, client *redis.Client) error {
	for _, n := range *notes {
		if err := n.validate(client); err != nil {
			fmt.Printf("Validation error: %s\n", err)
			continue
		}

		key := fmt.Sprintf("redkeep:note:%s", n.Id)

		client.Set(fmt.Sprintf("%s:title", key), n.Title, 0)
		client.Set(fmt.Sprintf("%s:created", key), n.Created, 0)
		client.Set(fmt.Sprintf("%s:updated", key), n.Updated, 0)
		client.Set(fmt.Sprintf("%s:body", key), n.Body, 0)

		n.pushTags(*client)

		updateList(n.Open, fmt.Sprintf("%s:open", key), *client)
		updateList(n.Closed, fmt.Sprintf("%s:closed", key), *client)
	}

	return nil
}

func (n *Note) pushTags(client redis.Client) {
	key := fmt.Sprintf("redkeep:note:%s", n.Id)

	tags_key := fmt.Sprintf("%s:tags", key)
	temp_key := fmt.Sprintf("%s:temp", key)
	remove_key := fmt.Sprintf("%s:remove", key)
	update_key := fmt.Sprintf("%s:update", key)

	client.SAdd(temp_key, n.Tags)

	client.SDiffStore(remove_key, tags_key, temp_key)
	client.SDiffStore(update_key, temp_key, tags_key)

	// Remove n from each tag set in remove_key
	for _, tag := range client.SMembers(remove_key).Val() {
		client.SRem(fmt.Sprintf("redkeep:tags:%s", tag), fmt.Sprintf("%s", n.Id))
	}

	// Add n to each tag set in update_key
	for _, tag := range client.SMembers(update_key).Val() {
		client.SAdd(fmt.Sprintf("redkeep:tags:%s", tag), fmt.Sprintf("%s", n.Id))
	}

	// Replace old tags_key
	client.Del(tags_key)
	client.Rename(temp_key, tags_key)

	// Clean up
	client.Del(remove_key)
	client.Del(update_key)
}

func updateList(list []string, key string, client redis.Client) {
	// Push new list to temp key
	temp := fmt.Sprintf("tmp:%s", key)
	client.LPush(temp, list)

	// Replace old key with temp key
	client.Del(key)
	client.Rename(temp, key)
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

func FromFile(filepath string) (*[]Note, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	notes := &[]Note{}
	err = yaml.Unmarshal(data, notes)
	if err != nil {
		return nil, err
	}

	return notes, nil
}

func ToFile(notes []Note, filepath string) error {

	data, err := yaml.Marshal(&notes)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(filepath, data, 0664)
	if err != nil {
		return err
	}

	return nil
}

func DeleteById(client *redis.Client, ids ...string) {
	for _, id := range ids {
		key := fmt.Sprintf("redkeep:note:%s", id)

		// Add n to each tag set in update_key
		tags_key := fmt.Sprintf("%s:tags", key)
		for _, tag := range client.SMembers(tags_key).Val() {
			client.SRem(fmt.Sprintf("redkeep:tags:%s", tag), id)
		}

		client.Del(fmt.Sprintf("%s:*", key))
	}
}

func ToJsonReader(notes *[]Note) (io.Reader, error) {
	b, err := json.Marshal(notes)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(b), nil
}
