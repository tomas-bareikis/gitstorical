package main

import (
	"log"
	"os"

	"github.com/go-git/go-git/v5"
)

func main() {
	tempPath, err := os.MkdirTemp("", "gitstorical")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("created dir", tempPath)

	defer os.RemoveAll(tempPath)

	_, err = git.PlainClone(tempPath, false, &git.CloneOptions{
		URL:      "https://github.com/go-git/go-git",
		Progress: os.Stdout,
	})

	if err != nil {
		log.Fatal(err)
	}
}
