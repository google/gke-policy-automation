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
	"testing"

	"github.com/sirupsen/logrus"
)

type MockEnvProvider struct {
	GetenvFn func(string) string
}

func (m MockEnvProvider) Getenv(s string) string {
	return m.GetenvFn(s)
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
			GetenvFn: func(s string) string {
				if s != levelVarName {
					t.Fatalf("env variable = %v; want %v", s, levelVarName)
				}
				return k
			},
		}
		level, _ := getLogLevel(m)
		if level != mappings[k] {
			t.Errorf("value = %v, level = %v; want %v", k, level, v)
		}
	}
}
