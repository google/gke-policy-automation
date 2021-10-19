package gke

import (
	"fmt"
	"testing"
)

func TestCreateReviewApp(t *testing.T) {
	clusterName := "testCluster"
	clusterLocation := "europe-warsaw2"
	projectName := "testProject"
	policyDirectory := "./policies"

	args := []string{"gke-review", "-c", clusterName, "-l", clusterLocation, "-p", projectName, "-d", policyDirectory}
	reviewMock := func(c *Config) {
		fmt.Printf("Got config %+v", c)
		if c.ClusterName != clusterName {
			t.Errorf("clusterName = %s; want %s", c.ClusterName, clusterName)
		}
		if c.ClusterLocation != clusterLocation {
			t.Errorf("clusterLocation = %s; want %s", c.ClusterLocation, clusterLocation)
		}
		if c.ProjectName != projectName {
			t.Errorf("projectName = %s; want %s", c.ProjectName, projectName)
		}
		if c.PolicyDirectory != policyDirectory {
			t.Errorf("policyDirectory = %s; want %s", c.PolicyDirectory, policyDirectory)
		}
	}
	err := CreateReviewApp(reviewMock).Run(args)
	if err != nil {
		t.Fatalf("error when running the review application: %v", err)
	}
}
