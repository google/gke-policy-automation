// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"os"
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type MockEnvProvider struct {
	LookupEnvFn func(key string) (string, bool)
}

func (m MockEnvProvider) LookupEnv(key string) (string, bool) {
	return m.LookupEnvFn(key)
}

type MockFileProvider struct {
	OpenFileFn func(name string, flag int, perm os.FileMode) (*os.File, error)
}

func (m MockFileProvider) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return m.OpenFileFn(name, flag, perm)
}

func TestGetLogLevel(t *testing.T) {
	mappings := map[string]logrus.Level{
		"TRACE": logrus.TraceLevel,
		"DEBUG": logrus.DebugLevel,
		"INFO":  logrus.InfoLevel,
		"WARN":  logrus.WarnLevel,
		"ERROR": logrus.ErrorLevel,
		"FATAL": logrus.FatalLevel,
		"bla":   defaultLogLevel,
	}
	for k, v := range mappings {
		m := MockEnvProvider{
			LookupEnvFn: func(key string) (string, bool) {
				if key != levelVarName {
					t.Fatalf("env variable = %v; want %v", key, levelVarName)
				}
				return k, true
			},
		}
		level, _ := getLogLevel(m)
		if level != mappings[k] {
			t.Errorf("value = %v, level = %v; want %v", k, level, v)
		}
	}
}

func TestGetLogLevelVarNotSet(t *testing.T) {
	m := MockEnvProvider{
		LookupEnvFn: func(key string) (string, bool) {
			return "", false
		},
	}
	_, err := getLogLevel(m)
	if err == nil {
		t.Errorf("err = nil, want error")
	}
}

func TestGetLogOutput(t *testing.T) {
	path := "/some/test/file"
	file := &os.File{}
	e := MockEnvProvider{
		LookupEnvFn: func(key string) (string, bool) {
			if key != pathVarName {
				t.Fatalf("env variable = %v; want %v", key, pathVarName)
			}
			return path, true
		},
	}
	f := MockFileProvider{
		OpenFileFn: func(name string, flag int, perm os.FileMode) (*os.File, error) {
			return file, nil
		},
	}
	output := getLogOutput(e, f)
	if !reflect.DeepEqual(output, file) {
		t.Fatalf("output file = %+v; want %+v", output, file)
	}
}

func TestGetLogOutputVarNotSet(t *testing.T) {
	file := os.Stderr
	e := MockEnvProvider{
		LookupEnvFn: func(key string) (string, bool) {
			return "", false
		},
	}
	f := MockFileProvider{
		OpenFileFn: func(name string, flag int, perm os.FileMode) (*os.File, error) {
			return file, nil
		},
	}
	output := getLogOutput(e, f)
	if !reflect.DeepEqual(output, file) {
		t.Fatalf("output file = %+v; want %+v", output, file)
	}
}

func TestGetLogFormatter(t *testing.T) {
	e := MockEnvProvider{
		LookupEnvFn: func(key string) (string, bool) {
			if key != formatVarName {
				t.Fatalf("env variable = %v; want %v", key, formatVarName)
			}
			return "json", true
		},
	}
	formatter := getLogFormatter(e)
	assert.IsType(t, &logrus.JSONFormatter{}, formatter)
}

func TestGetLogFormatterVarNotSet(t *testing.T) {
	e := MockEnvProvider{
		LookupEnvFn: func(key string) (string, bool) {
			return "", false
		},
	}
	formatter := getLogFormatter(e)
	assert.IsType(t, &logrus.TextFormatter{}, formatter)
}
