/*
 * Copyright (c) 2022, MegaEase
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package watch

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const kind = "watch"

// NotifyFunc is the callback function for watch
type NotifyFunc func(file string)

// Watch is the file watcher structure
type Watch struct {
	Files    map[string]FileItem
	Func     NotifyFunc
	Interval time.Duration
	done     chan bool
	wg       sync.WaitGroup
}

// Option is the file watcher option
type Option func(*Watch)

// WithFiles add the files to watch
func WithFiles(files []string) Option {
	return func(w *Watch) {
		w.AddFiles(files)
	}
}

// WithFunction set the watch function
func WithFunction(fn NotifyFunc) Option {
	return func(w *Watch) {
		w.Func = fn
	}
}

// WithInterval set the interval for the file watcher
func WithInterval(interval time.Duration) Option {
	return func(w *Watch) {
		w.Interval = interval
	}
}

// NewWatch create a new file watcher
func NewWatch(options ...Option) *Watch {
	w := &Watch{
		Files:    map[string]FileItem{},
		Func:     nil,
		Interval: 30 * time.Second,
		done:     make(chan bool, 1),
		wg:       sync.WaitGroup{},
	}
	// Loop through each option
	for _, opt := range options {
		// Call the option giving the instantiated
		// *House as the argument
		opt(w)
	}
	return w
}

// Stop the file watcher
func (w *Watch) Stop() {
	w.done <- true
	w.wg.Wait()
}

// AddFile add the file to watch
func (w *Watch) AddFile(file string) error {
	fileItem, err := NewFileItem(file)
	if err != nil {
		return err
	}
	w.Files[file] = *fileItem
	return nil
}

// AddFiles add the files to watch
func (w *Watch) AddFiles(file []string) error {
	for _, f := range file {
		if err := w.AddFile(f); err != nil {
			return err
		}
	}
	return nil
}

// RemoveFile remove the file from watch
func (w *Watch) RemoveFile(file string) {
	delete(w.Files, file)
}

// RemoveFiles remove the files from watch
func (w *Watch) RemoveFiles(file []string) {
	for _, f := range file {
		w.RemoveFile(f)
	}
}

// FileList return the file watcher list
func (w *Watch) FileList() map[string]FileItem {
	return w.Files
}

// Watch watch the files
func (w *Watch) Watch() {

	check := func() {
		for file, fileItem := range w.Files {
			if changed, err := fileItem.CheckChange(); err != nil {
				log.Errorf("[%s] check file change error: %v", kind, err)
				continue
			} else if changed {
				log.Infof("[%s] file changed: %s", kind, file)
				if w.Func != nil {
					w.Func(file)
				}
			}
		}
	}

	go func() {
		w.wg.Add(1)
		defer w.wg.Done()
		interval := time.NewTimer(w.Interval)
		defer interval.Stop()
		for {
			select {
			case <-w.done:
				log.Info("[watch] stopped")
				return
			case <-interval.C:
				check()
			}
		}
	}()
}
