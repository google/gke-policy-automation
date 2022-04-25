package gke

import "context"

type GKELocalClient struct {
	ctx context.Context
}

func NewLocalClient(ctx context.Context, dumpFile string) (*GKELocalClient, error) {
	return nil, nil
}

func (c *GKELocalClient) GetClusterName(name string) {

}

// type ClusterManagerClient interface {
// 	GetCluster(ctx context.Context, req *containerpb.GetClusterRequest, opts ...gax.CallOption) (*containerpb.Cluster, error)
// 	Close() error
// }

// type GKEClient struct {
// 	ctx    context.Context
// 	client ClusterManagerClient
// }

// func NewClient(ctx context.Context) (*GKEClient, error) {
// 	return newGKEClient(ctx)
// }

// func NewClientWithCredentialsFile(ctx context.Context, credentialsFile string) (*GKEClient, error) {
// 	return newGKEClient(ctx, option.WithCredentialsFile(credentialsFile))
// }

// func newGKEClient(ctx context.Context, opts ...option.ClientOption) (*GKEClient, error) {
// 	cli, err := container.NewClusterManagerClient(ctx, opts...)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &GKEClient{
// 		ctx:    ctx,
// 		client: cli,
// 	}, nil
// }

// func (c *GKEClient) GetCluster(name string) (*containerpb.Cluster, error) {
// 	req := &containerpb.GetClusterRequest{
// 		Name: name}
// 	return c.client.GetCluster(c.ctx, req)
// }

// func (c *GKEClient) Close() error {
// 	return c.client.Close()
// }

// func GetClusterName(project string, location string, name string) string {
// 	return fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, name)
// }
