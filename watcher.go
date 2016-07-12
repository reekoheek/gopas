package main

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Watcher struct {
	*Logger
	Watches      []string
	Extensions   []string
	Ignores      []string
	cb           func() (*Runner, error)
	runner       *Runner
	modifiedTime time.Time
}

func (w *Watcher) Watch(cb func() (*Runner, error)) error {
	w.cb = cb

	quit := make(chan error, 1)

	go func() {
		for {
			for _, dir := range w.Watches {
				filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
					if w.isIgnorable(path) {
						return filepath.SkipDir
					}

					if w.isAcceptable(path) && info.ModTime().After(w.modifiedTime) {
						w.modifiedTime = time.Now()
						if err := w.Start(); err != nil {
							quit <- err
							return errors.New("End of walk")
						}
					}
					return nil
				})
			}

			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		for {
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			token := strings.Trim(text, " \t\r\n")

			switch token {
			case "rs":
				if err := w.Start(); err != nil {
					quit <- err
					break
				}
			case "q":
				quit <- nil
				break
			}
		}
	}()

	return <-quit
}

func (w *Watcher) Start() error {
	if w.runner != nil && !w.runner.IsExited() {
		w.LogI("[WATCHER] Killing last process ...")
		if err := w.runner.Kill(); err != nil {
			return err
		}
		//time.Sleep(200 * time.Millisecond)
	}
	w.LogI("[WATCHER] Starting process ...")

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
