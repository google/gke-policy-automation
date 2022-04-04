// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package policy

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type WalkFn func(root string, fn filepath.WalkFunc) error
type ReadFn func(name string) ([]byte, error)

type LocalPolicySource struct {
	directory     string
	policyFileExt string
}

func NewLocalPolicySource(rootDirectory string) PolicySource {
	return &LocalPolicySource{
		directory:     rootDirectory,
		policyFileExt: "rego",
	}
}

func (src LocalPolicySource) String() string {
	return fmt.Sprintf("local directory: %s", src.directory)
}

func (src LocalPolicySource) GetPolicyFiles() ([]*PolicyFile, error) {
	files, err := src.getPolicyFiles(filepath.Walk)
	if err != nil {
		return nil, err
	}
	if err := readPolicyFiles(files, os.ReadFile); err != nil {
		return nil, err
	}
	return files, nil
}

func (src LocalPolicySource) getPolicyFiles(walkFn WalkFn) ([]*PolicyFile, error) {
	files := make([]*PolicyFile, 0)
	err := walkFn(src.directory, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, "."+src.policyFileExt) {
			files = append(files, &PolicyFile{
				Name:     info.Name(),
				FullName: path})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func readPolicyFiles(files []*PolicyFile, readFn ReadFn) error {
	for _, file := range files {
		data, err := readFn(file.FullName)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %s", file.FullName, err)
		}
		file.Content = string(data)
	}
	return nil
}
