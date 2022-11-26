package main

import (
	"log"
	"os"
	"time"

	"github.com/bitfield/script"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:     "gitstorical",
		Usage:    "runs a command on different versions of a git repo",
		Version:  "v0.01",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name:  "Tomas Bareikis",
				Email: "tomas.bareikis@pm.me",
			},
		},
		Action: do,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

	// Example: go run main.go https://github.com/go-git/go-git 'gocyclo -avg .' 'grep "Average"'
}

func do(cCtx *cli.Context) error {
	args := cCtx.Args()
	gitRepo := args.Get(0)
	command := args.Get(1)
	filter := args.Get(2)

	tempPath, err := os.MkdirTemp("", "gitstorical")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("created dir", tempPath)

	defer os.RemoveAll(tempPath)

	r, err := git.PlainClone(tempPath, false, &git.CloneOptions{
		URL:      gitRepo,
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

		output, err := script.Exec(command).Exec(filter).String()
		if err != nil {
			log.Fatal(err)
		}

		log.Println(output)
	}

	return nil
}
