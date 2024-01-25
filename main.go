package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/bitfield/script"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/urfave/cli/v2"
	giturls "github.com/whilp/git-urls"
)

func main() {
	log.SetHandler(text.New(os.Stderr))
	log.SetLevel(log.WarnLevel)

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
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "verbose",
				Value: false,
				Usage: "verbose mode",
				Action: func(ctx *cli.Context, b bool) error {
					log.SetLevel(log.DebugLevel)
					return nil
				},
			},
		},
		Action: do,
	}

	if err := app.Run(os.Args); err != nil {
		log.WithError(err).Fatal("gitstorcal error")
	}

	// Example: go run main.go git@github.com:go-git/go-git.git 'gocyclo -avg .' 'grep "Average"'
}

func do(cCtx *cli.Context) error {
	args := cCtx.Args()
	gitRepo := args.Get(0)
	command := args.Get(1)
	filter := args.Get(2)

	l := log.WithFields(
		log.Fields{
			"gitRepo": gitRepo,
			"command": command,
			"filter":  filter,
		},
	)

	tempPath, err := os.MkdirTemp("", "gitstorical")
	if err != nil {
		return err
	}
	l.WithField("tmpDir", tempPath).Debug("created temp dir")

	defer os.RemoveAll(tempPath)

	cloneOptions := &git.CloneOptions{
		URL:      gitRepo,
		Progress: os.Stdout,
		Depth:    1,
	}

	parsedGitURL, err := giturls.Parse(gitRepo)
	if err != nil {
		return err
	}

	if parsedGitURL.Scheme == "ssh" {
		sshAuthMethod, err := ssh.NewSSHAgentAuth("git")
		if err != nil {
			return err
		}
		cloneOptions.Auth = sshAuthMethod
	}

	r, err := git.PlainClone(tempPath, false, cloneOptions)
	if err != nil {
		return err
	}

	tagRefs, err := r.Tags()
	if err != nil {
		return err
	}

	allTagNames := []plumbing.ReferenceName{}
	err = tagRefs.ForEach(func(t *plumbing.Reference) error {
		allTagNames = append(allTagNames, t.Name())
		return nil
	})
	if err != nil {
		return err
	}

	l.WithField("refs", allTagNames).Debug("found refs")

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	os.Chdir(tempPath)

	for _, t := range allTagNames {
		out, err := processReference(w, t, command, filter)
		if err != nil {
			return err
		}

		out = strings.TrimSpace(out)
		fmt.Printf("%s,%s\n", t.String(), out)
	}

	return nil
}

func processReference(
	wt *git.Worktree,
	ref plumbing.ReferenceName,
	command, filter string,
) (string, error) {
	err := wt.Checkout(&git.CheckoutOptions{
		Branch: ref,
	})
	if err != nil {
		return "", err
	}

	out, err := script.Exec(command).String()
	if err != nil {
		return out, err
	}

	return script.Echo(out).Exec(filter).String()
}
