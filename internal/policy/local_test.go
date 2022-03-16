//Copyright 2022 Google LLC
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package policy

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fileInfoMock struct {
	nameFn  func() string
	isDirFn func() bool
}

func (m fileInfoMock) Name() string {
	return m.nameFn()
}

func (m fileInfoMock) IsDir() bool {
	return m.isDirFn()
}

func (m fileInfoMock) Size() int64 {
	return 0
}

func (m fileInfoMock) Mode() fs.FileMode {
	return fs.ModeExclusive
}

func (m fileInfoMock) ModTime() time.Time {
	return time.Now()
}

func (m fileInfoMock) Sys() interface{} {
	return nil
}

func TestNewLocalPolicySource(t *testing.T) {
	directory := "someDirectory"
	src := NewLocalPolicySource(directory)
	localSrc, ok := src.(*LocalPolicySource)
	if !ok {
		t.Errorf("Result of NewLocalPolicySource is not *LocalPolicySource")
	}
	if localSrc.directory != directory {
		t.Errorf("directory = %s; want %s", localSrc.directory, directory)
	}
}

func TestGetPolicyFiles(t *testing.T) {
	type fsObj struct {
		path string
		info fs.FileInfo
	}
	mockObjects := []fsObj{
		{"rego/policyOne.rego", fileInfoMock{
			nameFn:  func() string { return "policyOne.rego" },
			isDirFn: func() bool { return false }},
		},
		{"rego/subDirectory", fileInfoMock{
			nameFn:  func() string { return "subDirectory" },
			isDirFn: func() bool { return true }},
		},
		{"rego/subDirectory/policyTwo.rego", fileInfoMock{
			nameFn:  func() string { return "policyTwo.rego" },
			isDirFn: func() bool { return false }},
		},
		{"rego/file.txt", fileInfoMock{
			nameFn:  func() string { return "file.txt" },
			isDirFn: func() bool { return false }},
		},
	}

	mockWalkFn := func(root string, fn filepath.WalkFunc) error {
		for _, mockObj := range mockObjects {
			if err := fn(mockObj.path, mockObj.info, nil); err != nil {
				return err
			}
		}
		return nil
	}

	src := LocalPolicySource{directory: "dir", policyFileExt: "rego"}
	files, err := src.getPolicyFiles(mockWalkFn)
	if err != nil {
		t.Errorf("err is not nil; want nil")
	}
	if len(files) != 2 {
		t.Errorf("len(paths) = %d; want %d", len(files), 2)
	}
	for i, file := range files {
		if !strings.HasSuffix(file.Name, "."+src.policyFileExt) {
			t.Errorf("file[%d].Name = %s; want name with prefix %s", i, file.Name, "."+src.policyFileExt)
		}
		if !strings.HasSuffix(file.FullName, "."+src.policyFileExt) {
			t.Errorf("file[%d].FullName = %s; want fullName with prefix %s", i, file.FullName, "."+src.policyFileExt)
		}
	}
}

func TestReadPolicyFiles(t *testing.T) {
	files := []*PolicyFile{
		{Name: "policy_file_one.rego", FullName: "policies/policy_file_one.rego"},
		{Name: "policy_file_two.rego", FullName: "policies/subdir/policy_file_two.rego"},
	}
	contents := map[string]string{
		"policies/policy_file_one.rego":        "policy_file_one content",
		"policies/subdir/policy_file_two.rego": "policy_file_two.rego",
	}

	readFn := func(name string) ([]byte, error) {
		content, ok := contents[name]
		if !ok {
			return nil, fmt.Errorf("could not read content for file %s", name)
		}
		return []byte(content), nil
	}

	err := readPolicyFiles(files, readFn)
	if err != nil {
		t.Errorf("err is not nil; want nil")
	}
	for _, file := range files {
		if file.Content != contents[file.FullName] {
			t.Errorf("file %s content = %s; want %s", file.FullName, file.Content, contents[file.FullName])
		}
	}
}
