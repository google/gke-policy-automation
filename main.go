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

package main

import (
	"fmt"
	"os"

	"github.com/google/gke-policy-automation/internal/app"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("\nError: %s\n", err)
			os.Exit(1)
		}
	}()

	if err := app.NewPolicyAutomationCli(app.NewPolicyAutomationApp()).Run(os.Args); err != nil {
		fmt.Printf("\nError: %s\n", err)
		os.Exit(1)
	}
}
