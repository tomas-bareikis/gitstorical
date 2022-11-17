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

	r, err := git.PlainClone(tempPath, false, &git.CloneOptions{
		URL:      "https://github.com/go-git/go-git",
		Progress: os.Stdout,
	})
	if err != nil {
		log.Fatal(err)
	}

	w, err := r.Worktree()
	if err != nil {
		log.Fatal(err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: "refs/tags/v5.4.1",
	})
	if err != nil {
		log.Fatal(err)
	}
}
