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
	"errors"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	//log "github.com/sirupsen/logrus"
	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func createFiles(files []string) {
	for _, file := range files {
		f, _ := os.Create(file)
		f.Close()
	}
}

func removeFiles(files []string) {
	for _, file := range files {
		os.Remove(file)
	}
}

func writeFile(file string) {
	ioutil.WriteFile(file, []byte("test"), 0644)
}

func TestWatch(t *testing.T) {
	files := []string{"1.txt", "2.txt", "3.txt"}
	createFiles(files)
	defer removeFiles(files)

	wg := sync.WaitGroup{}

	w := NewWatch(WithFiles(files), WithInterval(time.Second), WithFunction(func(file string) {
		wg.Done()
		assert.Equal(t, file, files[0])
	}))
	defer w.Stop()

	for _, f := range files {
		list := w.FileList()
		assert.Contains(t, list, f)
	}

	w.RemoveFiles([]string{files[2]})
	list := w.FileList()
	assert.NotContains(t, list, files[2])

	w.Watch()

	writeFile(files[0])
	wg.Add(1)
	wg.Wait()
}

func TestWatchFailed(t *testing.T) {
	files := []string{"1.txt", "2.txt"}
	w := NewWatch(WithFiles(files), WithInterval(time.Second))
	assert.Empty(t, w.FileList())

	createFiles(files)
	defer removeFiles(files)

	w.AddFile(files[0])
	f := w.Files[files[0]]
	f.CheckSum = []byte("xyz")
	ch, err := f.CheckChange()
	assert.True(t, ch)
	assert.NoError(t, err)

	monkey.Patch(io.Copy, func(w io.Writer, r io.Reader) (int64, error) {
		return 0, errors.New("Errors")
	})
	w = NewWatch(WithFiles(files), WithInterval(time.Second))
	assert.Empty(t, w.FileList())

	ch, err = f.CheckChange()
	assert.False(t, ch)
	assert.Error(t, err)

	monkey.UnpatchAll()

	w.AddFile(files[0])

	monkey.Patch(SHA256, func(file string) ([]byte, error) {
		return nil, errors.New("Errors")
	})
	w.Watch()
	defer w.Stop()
	time.Sleep(time.Second)
}
