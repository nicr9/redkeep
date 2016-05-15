package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

var usage string = "Usage: redkeep <command> [<args>]"

func OpenEditor(filepath string) {
	cmd := exec.Command("vim", filepath, "+set filetype=yaml")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println(usage)
		os.Exit(2)
	}

	switch os.Args[1] {
	case "new":
		note := Note{}
		note.validate()
		notes := []Note{note}
		filepath := ToTempFile(notes)
		defer os.Remove(filepath)
		OpenEditor(filepath)
	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}
}
