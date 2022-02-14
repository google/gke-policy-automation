package app

import (
	"testing"

	"github.com/mikouaj/gke-review/internal/policy"
)

func TestCreateReviewApp(t *testing.T) {
	clusterName := "testCluster"
	clusterLocation := "europe-warsaw2"
	projectName := "testProject"
	credsFile := "./creds"
	gitRepo := "https://github.com/user/repo"
	gitBranch := "my-branch"
	gitDirectory := "rego-remote"
	localDirectory := "rego-local"

	args := []string{"gke-review",
		"-c", clusterName, "-l", clusterLocation,
		"-p", projectName,
		"-creds", credsFile,
		"-git-policy-repo", gitRepo,
		"-git-policy-branch", gitBranch,
		"-git-policy-dir", gitDirectory,
		"-local-policy-dir", localDirectory,
		"-s",
	}
	reviewMock := func(c *Config) {
		if c.ClusterName != clusterName {
			t.Errorf("clusterName = %s; want %s", c.ClusterName, clusterName)
		}
		if c.ClusterLocation != clusterLocation {
			t.Errorf("clusterLocation = %s; want %s", c.ClusterLocation, clusterLocation)
		}
		if c.ProjectName != projectName {
			t.Errorf("projectName = %s; want %s", c.ProjectName, projectName)
		}
		if c.CredentialsFile != credsFile {
			t.Errorf("CredentialsFile = %s; want %s", c.CredentialsFile, credsFile)
		}
		if !c.SilentMode {
			t.Errorf("SilentMode = %v; want true", c.SilentMode)
		}
		if c.GitRepository != gitRepo {
			t.Errorf("GitRepository = %s; want %s", c.GitRepository, gitRepo)
		}
		if c.GitBranch != gitBranch {
			t.Errorf("GitBranch = %s; want %s", c.GitBranch, gitBranch)
		}
		if c.GitDirectory != gitDirectory {
			t.Errorf("GitDirectory = %s; want %s", c.GitDirectory, gitDirectory)
		}
		if c.LocalDirectory != localDirectory {
			t.Errorf("LocalDirectory = %s; want %s", c.LocalDirectory, localDirectory)
		}
	}
	err := CreateReviewApp(reviewMock).Run(args)
	if err != nil {
		t.Fatalf("error when running the review application: %v", err)
	}
}

func TestCreateReviewApp_Defaults(t *testing.T) {
	args := []string{"gke-review",
		"-c", "testCluster", "-l", "europe-warsaw2",
		"-p", "testProject"}

	reviewMock := func(c *Config) {
		if c.GitRepository != DefaultGitRepository {
			t.Errorf("GitRepository = %s; want %s", c.GitRepository, DefaultGitRepository)
		}
		if c.GitBranch != DefaultGitBranch {
			t.Errorf("GitBranch = %s; want %s", c.GitBranch, DefaultGitBranch)
		}
		if c.GitDirectory != DefaultGitPolicyDir {
			t.Errorf("GitDirectory = %s; want %s", c.GitDirectory, DefaultGitPolicyDir)
		}
	}
	err := CreateReviewApp(reviewMock).Run(args)
	if err != nil {
		t.Fatalf("error when running the review application: %v", err)
	}
}

func TestGetPolicySource(t *testing.T) {
	c := &Config{
		GitRepository: DefaultGitRepository,
		GitBranch:     DefaultGitBranch,
		GitDirectory:  DefaultGitPolicyDir,
	}
	src := getPolicySource(c)
	if _, ok := src.(*policy.GitPolicySource); !ok {
		t.Errorf("policySource is not *GitPolicySource; want not *GitPolicySource")
	}
}

func TestGetPolicySource_local(t *testing.T) {
	c := &Config{
		GitRepository:  DefaultGitRepository,
		GitBranch:      DefaultGitBranch,
		GitDirectory:   DefaultGitPolicyDir,
		LocalDirectory: "some-local-dir",
	}
	src := getPolicySource(c)
	if _, ok := src.(*policy.LocalPolicySource); !ok {
		t.Errorf("policySource is not *LocalPolicySource; want not *LocalPolicySource")
	}
}
