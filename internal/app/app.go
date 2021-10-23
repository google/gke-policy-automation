package app

import (
	"context"
	"fmt"

	"github.com/mikouaj/gke-review/internal/gke"
	"github.com/mikouaj/gke-review/internal/policy"
	cli "github.com/urfave/cli/v2"
)

type Config struct {
	ClusterName     string
	ClusterLocation string
	ProjectName     string
	PolicyDirectory string
}

type Review func(c *Config)

func CreateReviewApp(review Review) *cli.App {
	config := &Config{}
	app := &cli.App{
		Name:  "gke-review",
		Usage: "Review GKE cluster against set of policies",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "project",
				Aliases:     []string{"p"},
				Usage:       "Name of a GCP project",
				Required:    true,
				Destination: &config.ProjectName,
			},
			&cli.StringFlag{
				Name:        "cluster",
				Aliases:     []string{"c"},
				Usage:       "Name of a GKE cluster to review",
				Required:    true,
				Destination: &config.ClusterName,
			},
			&cli.StringFlag{
				Name:        "location",
				Aliases:     []string{"l"},
				Usage:       "GKE cluster location (region or zone)",
				Required:    true,
				Destination: &config.ClusterLocation,
			},
			&cli.StringFlag{
				Name:        "directory",
				Aliases:     []string{"d"},
				Usage:       "Directory with GKE policies",
				Required:    true,
				Destination: &config.PolicyDirectory,
			},
		},
		Action: func(c *cli.Context) error {
			review(config)
			return nil
		},
	}
	return app
}

func GkeReview(c *Config) {
	ctx := context.Background()
	gke, err := gke.NewGKEClient(ctx)
	if err != nil {
		fmt.Printf("error when creating GKE client: %s", err)
		return
	}
	defer gke.Close()
	cluster, err := gke.GetCluster(c.ProjectName, c.ClusterLocation, c.ClusterName)
	if err != nil {
		fmt.Printf("error when fetching a cluster: %s", err)
		return
	}

	pa := policy.NewPolicyAgent(ctx, c.PolicyDirectory)
	results, err := pa.EvaluatePolicies(cluster)
	if err != nil {
		fmt.Printf("error when evaluating policies: %s", err)
	}
	fmt.Printf("Evaluated %d policies, %d errored\n\n", results.TotalCount, results.ErroredCount)
	for _, policy := range results.SuccessFull {
		fmt.Printf("Successful policy: %+v\n\n", policy)
	}
	for _, policy := range results.Failed {
		fmt.Printf("Failed policy: %+v\n\n", policy)
	}

	/*
		r := rego.New(
			rego.Load([]string{c.PolicyDirectory}, nil),
			rego.Input(cluster),
			//rego.Query("data.gke.policies"),
			//rego.Query("data.gke.violations"),
			rego.Query("data.gke.policies_data"),
		)
		fmt.Printf("Got rego %+v\n", r)
		rs, err := r.Eval(context.Background())
		if err != nil {
			fmt.Printf("Rego eval error %+v\n", err)
		}
		fmt.Printf("Rego eval result %+v\n", rs)

		// Inspect results.
		fmt.Println("len:", len(rs))
		fmt.Println("value:", rs[0].Expressions[0].Value)
		fmt.Println("allowed:", rs.Allowed()) // helper method
	*/

}
