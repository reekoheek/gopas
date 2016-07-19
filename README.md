# gopas
gopas.sublime-workspace

Build tool for Golang to build project outside GOPATH

## Install

```
go get github.com/reekoheek/gopas
```

## Usage

```
$ gopas -h
Usage: gopas <action> [<args...>]

Actions:
  list     List all dependencies
  install  Install dependencies
  run      Run go code
  help     Show help
```

## gopas.yml

Specify configuration for project by `gopas.yml` file

```
name: my.co/myname/mypackage

pre-build:
    - ["do", "something", "cli", "thing"]

dependencies:
    - other.co/unknown/app
    - another.id/some/other
```