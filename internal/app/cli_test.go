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
	"testing"

	cli "github.com/urfave/cli/v2"
)

func TestNewPolicyAutomationCli(t *testing.T) {
	app := NewPolicyAutomationApp()
	cmd := NewPolicyAutomationCli(app)
	validateCommandsExist(t, cmd.Commands, []string{"check", "dump", "version"})
}

func TestCheckClusterCommand(t *testing.T) {
	app := NewPolicyAutomationApp()
	cmd := createCheckCommand(app)
	validateCommandsExist(t, cmd.Subcommands, []string{"best-practices", "scalability", "policies"})
}

func validateCommandsExist(t *testing.T, commands []*cli.Command, expected []string) {
	expectedCmds := make(map[string]bool)
	for _, expectedCmd := range expected {
		expectedCmds[expectedCmd] = false
	}
	for _, cmd := range commands {
		if _, ok := expectedCmds[cmd.Name]; ok {
			expectedCmds[cmd.Name] = true
		}
	}
	for expectedSubCmd, present := range expectedCmds {
		if !present {
			t.Errorf("expected (sub)command %s is missing", expectedSubCmd)
		}
	}
}
