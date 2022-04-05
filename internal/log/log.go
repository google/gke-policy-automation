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
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	levelVarName    = "GKE_POLICY_LOG"
	pathVarName     = "GKE_POLICY_LOG_PATH"
	formatVarName   = "GKE_POLICY_LOG_FORMAT"
	defaultLogLevel = logrus.InfoLevel
)

var log = NewLogger()

type envProvider interface {
	Getenv(key string) string
}

type OsEnvProvider struct{}

func (p OsEnvProvider) Getenv(key string) string {
	return os.Getenv(key)
}

func NewLogger() *logrus.Logger {
	envProvider := OsEnvProvider{}
	logger := logrus.New()
	level, err := getLogLevel(envProvider)
	if err != nil {
		logger.SetOutput(io.Discard)
	}
	logger.SetLevel(level)
	return logger
}

func getLogLevel(p envProvider) (logrus.Level, error) {
	switch value := p.Getenv(levelVarName); strings.ToLower(value) {
	case "trace":
		return logrus.TraceLevel, nil
	case "debug":
		return logrus.DebugLevel, nil
	case "info":
		return logrus.InfoLevel, nil
	case "warn":
		return logrus.WarnLevel, nil
	case "error":
		return logrus.ErrorLevel, nil
	case "fatal":
		return logrus.FatalLevel, nil
	}
	return defaultLogLevel, fmt.Errorf("unsupported or missing log level")
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Printf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Warningf(format string, args ...interface{}) {
	log.Warningf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	log.Panicf(format, args...)
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Print(args ...interface{}) {
	log.Print(args...)
}

func Warn(args ...interface{}) {
	log.Warn(args...)
}

func Warning(args ...interface{}) {
	log.Warning(args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func Panic(args ...interface{}) {
	log.Panic(args...)
}
