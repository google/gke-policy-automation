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

package app

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

func TestPrintf(t *testing.T) {
	var buff bytes.Buffer
	testString := "some test message"
	out := Output{w: &buff}

	out.Printf(testString)
	result := buff.String()
	if buff.String() != testString {
		t.Errorf("Printf produced %s: want %s", result, testString)
	}
}

func TestErrorPrint(t *testing.T) {
	var buff bytes.Buffer
	errMsg := "could not test"
	cause := errors.New("test cause")
	out := Output{w: &buff}

	out.ErrorPrint(errMsg, cause)
	result := buff.String()
	expected := fmt.Sprintf("Error: %s: %s\n", errMsg, cause.Error())
	if result != expected {
		t.Errorf("ErrorPrint produced %s: want %s", result, expected)
	}
}
