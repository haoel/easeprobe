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
	"crypto/sha256"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

// FileItem is the file information
type FileItem struct {
	File     string
	CheckSum []byte
}

// NewFileItem create a new file item
func NewFileItem(file string) (*FileItem, error) {
	checksum, err := SHA256(file)
	if err != nil {
		log.Errorf("[%s] sha256 file error: %v", kind, err)
		return nil, err
	}
	f := &FileItem{
		File:     file,
		CheckSum: checksum,
	}
	return f, nil
}

// SHA256 calculate the sha256 of the file
func SHA256(file string) ([]byte, error) {
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	sha := sha256.New()
	if _, err := io.Copy(sha, fd); err != nil {
		return nil, err
	}

	return sha.Sum(nil), nil
}

// CheckChange check the file change or not
func (f *FileItem) CheckChange() (bool, error) {
	newCheckSum, err := SHA256(f.File)
	if newCheckSum == nil {
		return false, err
	}

	defer func() {
		f.CheckSum = newCheckSum
	}()

	if len(newCheckSum) != len(f.CheckSum) {
		return true, nil
	}
	for i := range newCheckSum {
		if newCheckSum[i] != f.CheckSum[i] {
			return true, nil
		}
	}
	return false, nil
}
