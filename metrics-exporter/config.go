package main

import "os"

const (
	KubeConfigClusterEnvName = "KUBE_CONFIG_GKE"
)

type config struct {
	KubeConfigGKE string
}

func NewConfigFromEnv() *config {
	return &config{
		KubeConfigGKE: os.Getenv(KubeConfigClusterEnvName),
	}
}
