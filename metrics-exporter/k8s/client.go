package k8s

import (
	"context"
	"encoding/base64"
	"fmt"

	"cloud.google.com/go/container/apiv1/containerpb"
	"github.com/google/gke-policy-automation/internal/inputs"
	"github.com/google/gke-policy-automation/internal/inputs/clients"
	"github.com/google/gke-policy-automation/internal/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func NewInClusterClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func NewClientFromGKE(ctx context.Context, clusterID string) (*kubernetes.Clientset, error) {
	ts, err := clients.NewGoogleTokenSource(ctx)
	if err != nil {
		return nil, err
	}
	token, err := ts.GetAuthToken()
	if err != nil {
		return nil, err
	}
	gkeInput, err := inputs.NewGKEApiInput(ctx)
	if err != nil {
		return nil, err
	}
	defer gkeInput.Close()
	data, err := gkeInput.GetData(clusterID)
	if err != nil {
		return nil, err
	}
	gkeData := data.(*containerpb.Cluster)
	kubeConfig, err := createKubeConfig(gkeData, token)
	if err != nil {
		return nil, err
	}
	config, _ := clientcmd.BuildConfigFromKubeconfigGetter("", func() (*clientcmdapi.Config, error) {
		return kubeConfig, nil
	})
	return kubernetes.NewForConfig(config)
}

func createKubeConfig(clusterData *containerpb.Cluster, clusterToken string) (*clientcmdapi.Config, error) {
	k8sKubeConfigContextName := "gke"
	clusterMasterAuth := clusterData.MasterAuth.ClusterCaCertificate
	clusterEndpoint := clusterData.Endpoint
	config := clientcmdapi.NewConfig()

	caCert, err := base64.StdEncoding.DecodeString(clusterMasterAuth)
	if err != nil {
		log.Debugf("Unable to retrieve clusterMasterAuth %s:", err)
		return nil, err
	}
	config.APIVersion = "v1"
	config.Kind = "Config"
	config.Clusters = map[string]*clientcmdapi.Cluster{
		k8sKubeConfigContextName: {
			CertificateAuthorityData: caCert,
			Server:                   fmt.Sprintf("https://%v", clusterEndpoint),
		},
	}
	config.AuthInfos = map[string]*clientcmdapi.AuthInfo{
		k8sKubeConfigContextName: {Token: clusterToken},
	}
	config.Contexts = map[string]*clientcmdapi.Context{
		k8sKubeConfigContextName: {
			Cluster:  k8sKubeConfigContextName,
			AuthInfo: k8sKubeConfigContextName,
		},
	}
	config.CurrentContext = k8sKubeConfigContextName
	log.Info("Local kubernetes cluster configuration created")
	return config, nil
}
