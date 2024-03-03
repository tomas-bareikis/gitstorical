# `gitstorical`

Run commands on different git repo versions.

`gitstorical` is a command-line tool that allows you to run a  command on different versions (tags) of a git repository. It provides a way to analyze or compare the output of the same command across various points in the repository's history.

## Installation

```sh
go install github.com/tomas-bareikis/gitstorical@latest
```

## Usage

List all files in the repository on each tag:

```sh
gitstorical --gitURL https://github.com/go-git/go-git \
  --command ls tags
```

Print the output in jsonl format:

```sh
gitstorical --gitURL https://github.com/go-git/go-git \
  --command ls \
  --outputFormat jsonl tags
```

Run only on tags newer than v5:

```sh
gitstorical --gitURL https://github.com/go-git/go-git \
  --command ls \
  --outputFormat jsonl tags \
  --tagFilter '>v5.0.0'
```

Checkout git repository to specific location. If a git repo already exists in the that location, it will not be overwritten.

```sh
gitstorical --checkoutDir /tmp/gitstorical/go-git \
   --gitURL https://github.com/go-git/go-git \
   --command ls \
   --outputFormat jsonl \
   tags --tagFilter '>v5.0.0'
```