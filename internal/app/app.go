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
	SilentMode      bool
	out             *Output
}

func (c *Config) init() {
	if c.SilentMode {
		c.out = NewSilentOutput()
	} else {
		c.out = NewStdOutOutput()
	}
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
			&cli.BoolFlag{
				Name:        "silent",
				Aliases:     []string{"s"},
				Usage:       "Silent mode",
				Required:    false,
				Destination: &config.SilentMode,
			},
		},
		Action: func(c *cli.Context) error {
			review(config)
			return nil
		},
	}
	config.init()
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
	c.out.Printf(c.out.Color("[white][bold]Fetching GKE cluster details... [projects/%s/locations/%s/clusters/%s]\n"),
		c.ProjectName,
		c.ClusterLocation,
		c.ClusterName)
	cluster, err := gke.GetCluster(c.ProjectName, c.ClusterLocation, c.ClusterName)
	if err != nil {
		c.out.ErrorPrint("could not fetch the cluster details", err)
		return
	}
	c.out.Printf(c.out.Color("[white][bold]Evaluating REGO policies... [source: %q directory]\n"),
		c.PolicyDirectory)
	pa := policy.NewPolicyAgent(ctx, c.PolicyDirectory)
	results, err := pa.EvaluatePolicies(cluster)
	if err != nil {
		c.out.ErrorPrint("could not parse policies", err)
		return
	}

	if len(results.Errored()) > 0 {
		c.out.Printf(c.out.Color("\n[white][bold]Policy parsing errors:\n\n"))
		for _, errored := range results.Errored() {
			c.out.Printf(c.out.Color("[light_yellow][bold]- %s: [reset][yellow]%s\n"), errored.Name, errored.ProcessingErrors[0])
		}
	}

	for _, group := range results.Groups() {
		c.out.Printf(c.out.Color("\n[white][bold]Group %q:\n\n"), group)
		for _, policy := range results.Policies(group) {
			if policy.Valid {
				c.out.Printf(c.out.Color("[bold][green][\u2713] %s: [reset][green]%s\n"), policy.FullName, policy.Description)
			}
			if !policy.Valid {
				c.out.Printf(c.out.Color("[bold][red][x] %s: [reset][red]%s. [bold]Violations:[reset][red] %s\n"), policy.FullName, policy.Description, policy.Violations[0])
			}
		}
	}

	c.out.Printf(c.out.Color("\n[bold][green]Review complete! Policies: %d valid, %d violated, %d errored.\n"),
		results.ValidCount(),
		results.ViolatedCount(),
		results.ErroredCount())
}
