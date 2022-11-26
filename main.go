package main

import (
	"log"
	"os"
	"strings"

	"github.com/bitfield/script"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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

	tagRefs, err := r.Tags()
	if err != nil {
		log.Fatal(err)
	}

	allTagNames := []plumbing.ReferenceName{}
	err = tagRefs.ForEach(func(t *plumbing.Reference) error {
		allTagNames = append(allTagNames, t.Name())
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	w, err := r.Worktree()
	if err != nil {
		log.Fatal(err)
	}

	os.Chdir(tempPath)

	for _, t := range allTagNames {
		err = w.Checkout(&git.CheckoutOptions{
			Branch: t,
		})
		if err != nil {
			log.Fatal(err)
		}

		output, err := script.Exec("gocyclo -avg .").String()
		if err != nil {
			log.Fatal(err)
		}

		lines := strings.Split(output, "\n")
		log.Println(lines[len(lines)-2])
	}
}
