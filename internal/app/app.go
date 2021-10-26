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
	Printf(Color("[white][bold]Fetching GKE cluster details... [projects/%s/locations/%s/clusters/%s]\n"),
		c.ProjectName,
		c.ClusterLocation,
		c.ClusterName)
	cluster, err := gke.GetCluster(c.ProjectName, c.ClusterLocation, c.ClusterName)
	if err != nil {
		ErrorPrint("could not fetch the cluster details", err)
		return
	}
	Printf(Color("[white][bold]Evaluating REGO policies... [source: %q directory]\n"),
		c.PolicyDirectory)
	pa := policy.NewPolicyAgent(ctx, c.PolicyDirectory)
	results, err := pa.EvaluatePolicies(cluster)
	if err != nil {
		fmt.Printf("error when evaluating policies: %s", err)
	}

	Printf(Color("\n[bold][green]Review complete! Policies: %d valid, %d violated, %d errored.\n"),
		results.ValidCount(),
		results.ViolatedCount(),
		results.ErroredCount())
}
