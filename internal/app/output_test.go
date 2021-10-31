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
