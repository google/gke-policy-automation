package gke

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/gke-policy-automation/internal/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientauthv1b1 "k8s.io/client-go/pkg/apis/clientauthentication/v1beta1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/retry"
)

type cred struct {
	googleDefaultTokenSource func(ctx context.Context, scope ...string) (oauth2.TokenSource, error)
	k8sStartingConfig        func() (*clientcmdapi.Config, error)
}

func newCred() *cred {
	return &cred{
		googleDefaultTokenSource: google.DefaultTokenSource,
		k8sStartingConfig:        k8sStartingConfig,
	}
}

type Creds struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Status     Status `json:"status"`
}

type Status struct {
	ExpirationTimestamp time.Time `json:"expirationTimestamp"`
	Token               string    `json:"token"`
}

func getClusterToken() (string, error) {
	cred := newCred()
	var execCredential *clientauthv1b1.ExecCredential
	var creds Creds

	token, expiry, err := cred.defaultAccessToken()
	if err != nil {
		log.Errorf("unable to retrieve default access token: %w", err)
		return "", err
	}

	execCredential = &clientauthv1b1.ExecCredential{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ExecCredential",
			APIVersion: "client.authentication.k8s.io/v1beta1",
		},
		Status: &clientauthv1b1.ExecCredentialStatus{
			Token:               token,
			ExpirationTimestamp: expiry,
		},
	}

	execCredentialJSON, err := formatToJSON(execCredential)
	if err != nil {
		log.Errorf("unable to convert credentials to json: %w", err)
		return "", err
	}

	if err := json.Unmarshal([]byte(execCredentialJSON), &creds); err != nil {
		log.Errorf("unable to retrieve credentials: %w", err)
		return "", fmt.Errorf("unable to retrieve credentials: %w", err)
	}

	return creds.Status.Token, nil
}

func (c *cred) defaultAccessToken() (string, *metav1.Time, error) {
	var tok *oauth2.Token
	var defaultScopes = []string{
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/userinfo.email"}

	err := retry.OnError(retry.DefaultBackoff, func(err error) bool { return true }, func() error {
		ts, err := c.googleDefaultTokenSource(context.Background(), defaultScopes...)
		if err != nil {
			log.Errorf("cannot construct google default token source: %w", err)
			return err
		}

		tok, err = ts.Token()
		if err != nil {
			log.Errorf("cannot retrieve default token from google default token source: %w", err)
			return err
		}

		return nil
	})
	if err != nil {
		log.Errorf("getting google default token failed after multiple retries: %w", err)
		return "", nil, err
	}

	return tok.AccessToken, &metav1.Time{Time: tok.Expiry}, nil
}

func formatToJSON(i interface{}) (string, error) {
	s, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		log.Errorf("unable to unmarshal credentials: %w", err)
		return "", err
	}
	return string(s), nil
}

func k8sStartingConfig() (*clientcmdapi.Config, error) {
	po := clientcmd.NewDefaultPathOptions()
	return po.GetStartingConfig()
}
