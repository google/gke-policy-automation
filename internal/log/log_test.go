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
		level := getLogLevel(m)
		if level != mappings[k] {
			t.Errorf("value = %v, level = %v; want %v", k, level, v)
		}
	}
}
