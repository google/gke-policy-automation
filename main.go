package main

import (
	"fmt"
	"os"

	"github.com/mikouaj/gke-review/internal/app"
)

func main() {
	if err := app.NewPolicyAutomationCli(app.NewPolicyAutomationApp()).Run(os.Args); err != nil {
		fmt.Printf("\nError: %s\n", err)
	}
}
