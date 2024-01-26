package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tomas-bareikis/gitstorical/files"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/bitfield/script"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/urfave/cli/v2"
	giturls "github.com/whilp/git-urls"
)

var checkoutDir string

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
			&cli.StringFlag{
				Name:  "checkoutDir",
				Value: "",
				Usage: "directory where the git repo will be checked out",
				Action: func(ctx *cli.Context, s string) error {
					checkoutDir = s
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
	gitURL := args.Get(0)
	command := args.Get(1)
	filter := args.Get(2)

	l := log.WithFields(
		log.Fields{
			"gitURL":      gitURL,
			"command":     command,
			"filter":      filter,
			"checkoutDir": checkoutDir,
		},
	)

	var err error

	if checkoutDir == "" {
		checkoutDir, err = os.MkdirTemp("", "gitstorical")
		if err != nil {
			return err
		}
		l.WithField("tmpDir", checkoutDir).Debug("created temp dir")

		defer os.RemoveAll(checkoutDir)
	}

	if !files.Exists(checkoutDir) {
		err := os.MkdirAll(checkoutDir, os.ModePerm)
		if err != nil {
			return err
		}

		l.Debug("created checkout dir")
	}

	dirEmpty, err := files.IsDirEmpty(checkoutDir)
	if err != nil {
		return err
	}

	var repo *git.Repository
	if dirEmpty {
		l.Debug("starting repo clone")

		err = cloneToPath(checkoutDir, gitURL)
		if err != nil {
			return err
		}

		l.Debug("repo cloning complete")
	} else {
		l.Debug("checkoutDir not empty, skipping clone")
	}
	
	repo, err = git.PlainOpen(checkoutDir)
	if err != nil {
		return err
	}

	tagRefs, err := repo.Tags()
	if err != nil {
		return err
	}
	l.Debug("retrieved all tags")

	allTagNames := []plumbing.ReferenceName{}
	err = tagRefs.ForEach(func(t *plumbing.Reference) error {
		allTagNames = append(allTagNames, t.Name())
		return nil
	})
	if err != nil {
		return err
	}

	l.WithField("refs", allTagNames).Debug("found refs")

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	os.Chdir(checkoutDir)

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
		Force:  true,
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

func cloneToPath(path, gitURL string) error {
	cloneOptions := &git.CloneOptions{
		URL:      gitURL,
		Progress: os.Stdout,
	}

	parsedGitURL, err := giturls.Parse(gitURL)
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

	_, err = git.PlainClone(checkoutDir, false, cloneOptions)
	return err
}
