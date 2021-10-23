package main

import (
	"fmt"
	"os"

	"github.com/mikouaj/gke-review/internal/app"
)

func main() {
	if err := app.CreateReviewApp(app.GkeReview).Run(os.Args); err != nil {
		fmt.Printf("error %v", err)
	}
}
