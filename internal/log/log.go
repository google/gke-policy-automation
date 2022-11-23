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
	LookupEnv(key string) (string, bool)
}

type fileProvider interface {
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
}

type osEnvProvider struct{}

func (p osEnvProvider) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

type osFileProvider struct{}

func (p osFileProvider) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

func NewLogger() *logrus.Logger {
	envProvider := osEnvProvider{}
	fileProvider := osFileProvider{}
	logger := logrus.New()
	output, level := getLogOutputLevel(envProvider, fileProvider)
	logger.SetLevel(level)
	logger.SetOutput(output)
	logger.SetFormatter(getLogFormatter(envProvider))
	return logger
}

func getLogOutputLevel(e envProvider, f fileProvider) (io.Writer, logrus.Level) {
	lvl, err := getLogLevel(e)
	if err != nil {
		return io.Discard, lvl
	}
	return getLogOutput(e, f), lvl
}

func getLogLevel(e envProvider) (logrus.Level, error) {
	value, ok := e.LookupEnv(levelVarName)
	if !ok {
		return defaultLogLevel, fmt.Errorf("env variable %q not set", levelVarName)
	}
	var level logrus.Level
	switch strings.ToLower(value) {
	case "trace":
		level = logrus.TraceLevel
	case "debug":
		level = logrus.DebugLevel
	case "info":
		level = logrus.InfoLevel
	case "warn":
		level = logrus.WarnLevel
	case "error":
		level = logrus.ErrorLevel
	case "fatal":
		level = logrus.FatalLevel
	default:
		fmt.Printf("unknown log level %q, using defaults", value)
		level = defaultLogLevel
	}
	return level, nil
}

func getLogOutput(e envProvider, f fileProvider) io.Writer {
	logFilePath, ok := e.LookupEnv(pathVarName)
	if !ok {
		return os.Stderr
	}
	file, err := f.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("couldn't open log file %q, using default log output; %s\n", logFilePath, err)
		return os.Stderr
	}
	return file
}

func getLogFormatter(e envProvider) logrus.Formatter {
	switch value, _ := e.LookupEnv(formatVarName); strings.ToLower(value) {
	case "json":
		return &logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyLevel: "severity",
				logrus.FieldKeyMsg:   "message",
			},
		}
	default:
		return &logrus.TextFormatter{}
	}
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
