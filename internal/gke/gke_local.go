package gke

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	containerpb "google.golang.org/genproto/googleapis/container/v1"
)

type GKELocalClient struct {
	ctx context.Context
}

func NewLocalClient(ctx context.Context, dumpFile string) (*GKELocalClient, error) {
	return nil, nil
}

func (c *GKELocalClient) GetClusterName(name string) {

}

// to add json file path
func (c *GKELocalClient) GetCluster() (*containerpb.Cluster, error) {
	var cluster containerpb.Cluster

	clusterDumpFile, err := os.Open("test.json")
	if err != nil {
		return &cluster, err
	}
	defer clusterDumpFile.Close()

	byteValue, err := ioutil.ReadAll(clusterDumpFile)
	if err != nil {
		return &cluster, err
	}

	err = json.Unmarshal(byteValue, &cluster)
	if err != nil {
		return &cluster, err
	}
	// if we os.Open returns an error then handle it
	if err != nil {
		return &cluster, err
	}
	fmt.Println("Successfully Opened cluster data file")
	// defer the closing of our jsonFile so that we can parse it later on

	return &cluster, err
}
