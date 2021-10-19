package main

import (
	"fmt"
	"os"

	"github.com/mikouaj/gke-review/gke"
)

func main() {
	if err := gke.CreateReviewApp(gke.GkeReview).Run(os.Args); err != nil {
		fmt.Printf("error %v", err)
	}
}
