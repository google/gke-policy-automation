package app

import (
	"context"

	"github.com/mikouaj/gke-review/internal/gke"
)

type Config struct {
	ClusterName     string
	ClusterLocation string
	ProjectName     string
	PolicyDirectory string
	SilentMode      bool
	CredentialsFile string

	ctx context.Context
	out *Output
	gke *gke.GKEClient
}

func (c *Config) Load(ctx context.Context) error {
	c.ctx = ctx
	if c.SilentMode {
		c.out = NewSilentOutput()
	} else {
		c.out = NewStdOutOutput()
	}
	var err error
	if c.CredentialsFile != "" {
		c.gke, err = gke.NewClientWithCredentialsFile(ctx, c.CredentialsFile)
	} else {
		c.gke, err = gke.NewClient(ctx)
	}
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) Close() error {
	if c.gke != nil {
		return c.gke.Close()
	}
	return nil
}
