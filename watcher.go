package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Watcher struct {
	Watches      []string
	Extensions   []string
	Ignores      []string
	LogI         func(s string, args ...interface{})
	cb           func() (*Runner, error)
	runner       *Runner
	modifiedTime time.Time
}

func (w *Watcher) Watch(cb func() (*Runner, error)) error {
	w.cb = cb

	for {
		for _, dir := range w.Watches {
			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if w.isIgnorable(path) {
					return filepath.SkipDir
				}

				if w.isAcceptable(path) && info.ModTime().After(w.modifiedTime) {
					w.modifiedTime = time.Now()
					w.Start()
				}
				return nil
			})
		}

		time.Sleep(1 * time.Second)
	}
	return nil
}

func (w *Watcher) Start() error {
	if w.runner != nil && !w.runner.IsExited() {
		w.LogI("Killing last watched process ...")
		w.runner.Kill()
	}
	w.LogI("Starting watched process ...")

	runner, err := w.cb()
	if err != nil {
		return err
	}
	w.runner = runner
	return nil
}

/**
 * Check whether path is ignorable
 *
 * @param {string} path
 */
func (w *Watcher) isIgnorable(path string) bool {
	ignorable := false
	for _, ignore := range w.Ignores {
		if matched, _ := filepath.Match(ignore, path); matched {
			ignorable = true
			break
		}
	}
	return ignorable
}

/**
 * Check whether path having acceptable extension
 *
 * @param {string} path
 */
func (w *Watcher) isAcceptable(path string) bool {
	for _, ext := range w.Extensions {
		if strings.HasSuffix(path, "."+ext) {
			return true
		}
	}

	return false
}
