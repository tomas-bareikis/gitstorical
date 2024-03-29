package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"github.com/tomas-bareikis/gitstorical/files"
	"github.com/tomas-bareikis/gitstorical/format"
	"github.com/tomas-bareikis/gitstorical/ref"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bitfield/script"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/urfave/cli/v2"
	giturls "github.com/whilp/git-urls"
)

var gitURL string
var command string
var checkoutDir string
var outputFormat = format.Plain
var semverConstraints *semver.Constraints
var log *zap.SugaredLogger

func main() {
	logLevel := zap.NewAtomicLevel()

	zapLog := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(os.Stderr),
		logLevel,
	))

	//nolint errcheck
	defer zapLog.Sync()
	log = zapLog.Sugar()

	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print only the version",
	}

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
		Commands: []*cli.Command{
			{
				Name:   "tags",
				Usage:  "run on repository tags",
				Action: runOnTags,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "tagFilter",
						Aliases: []string{"t"},
						Value:   "",
						Usage:   "semver constraint to filter tags by `FILTER`, e.g. '>=1.0.0 <2.0.0'",
						Action: func(ctx *cli.Context, s string) error {
							var err error

							semverConstraints, err = semver.NewConstraint(s)
							return err
						},
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "gitURL",
				Aliases:     []string{"u"},
				Value:       "",
				Usage:       "fetch repository from `URL`",
				Required:    true,
				Destination: &gitURL,
			},
			&cli.StringFlag{
				Name:        "command",
				Aliases:     []string{"c"},
				Value:       "",
				Usage:       "On ech ref, `COMMAND` will be executed",
				Required:    true,
				Destination: &command,
			},
			&cli.StringFlag{
				Name:        "checkoutDir",
				Aliases:     []string{"d"},
				Value:       "",
				Usage:       "The git repo will be checked out at `DIR`",
				Destination: &checkoutDir,
			},
			&cli.StringFlag{
				Name:    "outputFormat",
				Aliases: []string{"f"},
				Value:   "plain",
				Usage:   "output `FORMAT` to use [plain, jsonl]",
				Action: func(ctx *cli.Context, s string) error {
					var err error

					outputFormat, err = format.ParseType(s)
					return err
				},
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Value:   false,
				Usage:   "verbose mode",
				Action: func(ctx *cli.Context, b bool) error {
					logLevel.SetLevel(zapcore.DebugLevel)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.With(err).Fatal("gitstorcal error")
	}
}

func runOnTags(cCtx *cli.Context) error {
	l := log.With(
		"gitURL", gitURL,
		"command", command,
		"checkoutDir", checkoutDir,
	)

	var err error

	if checkoutDir == "" {
		checkoutDir, err = os.MkdirTemp("", "gitstorical")
		if err != nil {
			return errors.Wrap(err, "failed to create temp dir")
		}
		l.With("tmpDir", checkoutDir).Debug("created temp dir")

		defer os.RemoveAll(checkoutDir)
	}

	if !files.Exists(checkoutDir) {
		err := os.MkdirAll(checkoutDir, os.ModePerm)
		if err != nil {
			return errors.Wrap(err, "failed to create checkout dir")
		}

		l.Debug("created checkout dir")
	}

	dirEmpty, err := files.IsDirEmpty(checkoutDir)
	if err != nil {
		return errors.Wrap(err, "failed to check if checkout dir is empty")
	}

	var repo *git.Repository
	if dirEmpty {
		l.Debug("starting repo clone")

		err = cloneToPath(checkoutDir, gitURL)
		if err != nil {
			return errors.Wrap(err, "failed to clone repo")
		}

		l.Debug("repo cloning complete")
	} else {
		l.Debug("checkoutDir not empty, skipping clone")
	}

	repo, err = git.PlainOpen(checkoutDir)
	if err != nil {
		return errors.Wrap(err, "failed to open repo")
	}

	tagRefs, err := repo.Tags()
	if err != nil {
		return errors.Wrap(err, "failed to retrieve tags")
	}
	l.Debug("retrieved all tags")

	allTagNames := []plumbing.ReferenceName{}
	err = tagRefs.ForEach(func(t *plumbing.Reference) error {
		allTagNames = append(allTagNames, t.Name())
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to iterate over tags")
	}

	if semverConstraints != nil {
		allTagNames, err = ref.TagsFilter(allTagNames, semverConstraints, l)
		if err != nil {
			return errors.Wrap(err, "failed to filter tags")
		}
	}

	ref.SortTags(allTagNames, log)

	l.With("refs", allTagNames).Debug("found refs")

	w, err := repo.Worktree()
	if err != nil {
		return errors.Wrap(err, "failed to get worktree")
	}

	err = os.Chdir(checkoutDir)
	if err != nil {
		return errors.Wrap(err, "failed to change dir")
	}

	for _, t := range allTagNames {
		out, err := processReference(w, t, command)
		if err != nil {
			return errors.Wrapf(err, "failed to process ref %s", t.Short())
		}

		formatted, err := format.String(outputFormat, t, out)
		if err != nil {
			return errors.Wrapf(err, "failed to format output for ref %s", t.Short())
		}

		fmt.Print(formatted)
	}

	return nil
}

func processReference(
	wt *git.Worktree,
	ref plumbing.ReferenceName,
	command string,
) (string, error) {
	err := wt.Checkout(&git.CheckoutOptions{
		Branch: ref,
		Force:  true,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to checkout ref")
	}

	command = fmt.Sprintf("bash -c '%s'", command)
	out, err := script.Exec(command).String()
	if err != nil {
		return out, errors.Wrap(err, "failed to execute command")
	}

	return script.Echo(out).String()
}

func cloneToPath(cloneTo, gitURL string) error {
	cloneOptions := &git.CloneOptions{
		URL:      gitURL,
		Progress: os.Stderr,
	}

	parsedGitURL, err := giturls.Parse(gitURL)
	if err != nil {
		return errors.Wrap(err, "failed to parse git url")
	}

	if parsedGitURL.Scheme == "ssh" {
		sshAuthMethod, err := ssh.NewSSHAgentAuth("git")
		if err != nil {
			return err
		}
		cloneOptions.Auth = sshAuthMethod
	}

	_, err = git.PlainClone(cloneTo, false, cloneOptions)
	return errors.Wrap(err, "failed to clone repo")
}
