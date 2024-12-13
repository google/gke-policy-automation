// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app

import (
	"encoding/json"
	"fmt"

	"github.com/google/gke-policy-automation/internal/config"
	"github.com/google/gke-policy-automation/internal/gke"
	"github.com/google/gke-policy-automation/internal/inputs"
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/outputs"
	"github.com/google/gke-policy-automation/internal/policy"
)

// getClusters retrieves lists of a clusters for further processing
// from the sources that are defined in a configuration.
func (p *PolicyAutomationApp) getClusters() ([]string, error) {
	if p.config.DumpFile != "" {
		log.Debugf("using local cluster discovery client on a file %s", p.config.DumpFile)
		dc := gke.NewLocalDiscoveryClient(p.config.DumpFile)
		return dc.GetClustersInOrg("doesn't-matter-for-local-discovery")
	}
	if p.config.ClusterDiscovery.Enabled {
		var dc gke.DiscoveryClient
		var err error
		if p.config.CredentialsFile != "" {
			log.Debugf("instantiating cluster discovery client with a credentials file")
			dc, err = gke.NewDiscoveryClientWithCredentialsFile(p.ctx, p.config.CredentialsFile)
		} else {
			log.Debugf("instantiating cluster discovery client")
			dc, err = gke.NewDiscoveryClient(p.ctx)
		}
		if err != nil {
			return nil, err
		}
		p.discovery = dc
		return p.discoverClusters()
	}
	clusters := make([]string, 0, len(p.config.Clusters))
	for _, configCluster := range p.config.Clusters {
		clusterID, err := getClusterID(configCluster)
		log.Debugf("cluster id from config: %s", clusterID)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, clusterID)
	}
	return clusters, nil
}

// discoverClusters discovers clusters according to the cluster discovery configuration.
func (p *PolicyAutomationApp) discoverClusters() ([]string, error) {
	if p.config.ClusterDiscovery.Organization != "" {
		log.Infof("Discovering clusters in organization %s", p.config.ClusterDiscovery.Organization)
		p.out.Printf("%s %s\n",
			outputs.IconInfo,
			consoleInfoColorF("Discovering clusters in for organization... [%s]", p.config.ClusterDiscovery.Organization),
		)
		return p.discovery.GetClustersInOrg(p.config.ClusterDiscovery.Organization)
	}
	clusters := make([]string, 0)
	for _, folder := range p.config.ClusterDiscovery.Folders {
		log.Infof("Discovering clusters in folder %s", folder)
		p.out.Printf("%s %s\n",
			outputs.IconInfo,
			consoleInfoColorF("Discovering clusters in folder... [%s]", folder),
		)
		results, err := p.discovery.GetClustersInFolder(folder)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, results...)
	}
	for _, project := range p.config.ClusterDiscovery.Projects {
		log.Infof("Discovering clusters in project %s", project)
		p.out.Printf("%s %s\n",
			outputs.IconInfo,
			consoleInfoColorF("Discovering clusters in project... [%s]", project),
		)
		results, err := p.discovery.GetClustersInProject(project)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, results...)
	}
	log.Debugf("discovered %v clusters in projects and folders", len(clusters))
	return clusters, nil
}

func (p *PolicyAutomationApp) evaluateClusters(regoPackageBases []string) error {
	log.Info("Cluster review starting")
	files, err := p.loadPolicyFiles()
	if err != nil {
		return err
	}
	if len(files) == 0 {
		p.out.Printf("%s\n", consoleWarnColorF("No policies to check against"))
		log.Errorf("No policies to check against")
		return errNoPolicies
	}
	// create a PolicyAgent client instance
	pa := policy.NewPolicyAgent(p.ctx)
	p.out.Printf("%s %s\n",
		outputs.IconInfo,
		consoleInfoColorF("Parsing REGO policies..."),
	)
	log.Info("Parsing rego policies")
	// parsing policies before running checks
	if err := pa.WithFiles(files, p.config.PolicyExclusions); err != nil {
		p.out.ErrorPrint("could not parse policy files", err)
		log.Errorf("could not parse policy files: %s", err)
		return err
	}

	clusterIds, err := p.getClusters()
	if err != nil {
		p.out.ErrorPrint("could not identify clusters", err)
		log.Errorf("could not identify clusters: %s", err)
		return err
	}
	if len(clusterIds) < 1 {
		p.out.Printf("%s\n", consoleWarnColorF("No clusters to check, finishing..."))
		log.Info("Cluster review finished")
		p.out.Printf("%s %s\n",
			outputs.IconInfo,
			consoleInfoColorF("Cluster review finished"),
		)
		return nil
	}
	p.out.Printf("%s %s\n",
		outputs.IconInfo,
		consoleInfoColorF("Fetching data from %d input(s) for %d cluster(s)", len(p.inputs), len(clusterIds)),
	)
	clusterData, errors := inputs.GetAllInputsData(p.inputs, clusterIds)
	if len(errors) > 0 {
		p.out.ErrorPrint("could not fetch the cluster details", errors[0])
		log.Errorf("could not fetch cluster details: %s", errors[0])
		return errors[0]
	}
	val, _ := json.MarshalIndent(clusterData, "", "    ")
	log.Debugf("[DEBUG] cluster: %s", string(val))

	evalResults := &evaluationResults{}
	for _, cluster := range clusterData {
		p.out.Printf("%s %s\n",
			outputs.IconInfo,
			consoleInfoColorF("Evaluating policies against GKE cluster... [%s]", cluster.Name),
		)
		log.Infof("Evaluating policies against GKE cluster %s", cluster.Name)
		for _, pkgBase := range regoPackageBases {
			evalResult, err := pa.Evaluate(cluster, pkgBase)
			if err != nil {
				p.out.ErrorPrint("failed to evaluate policies", err)
				log.Errorf("could not evaluate rego policies on cluster %s: %s", cluster.Name, err)
				return err
			}
			evalResult.ClusterID = cluster.Name
			evalResults.Add(evalResult)
		}
	}

	for _, c := range p.collectors {
		log.Infof("Collector %s registering the results", c.Name())
		p.out.Printf("%s %s\n",
			outputs.IconInfo,
			consoleInfoColorF("Writing evaluation results ... [%s]", c.Name()),
		)
		if err = c.RegisterResult(evalResults.List()); err != nil {
			p.out.ErrorPrint("failed to register evaluation results", err)
			log.Errorf("could not register evaluation results: %s", err)
			return err
		}
		if err = c.Close(); err != nil {
			p.out.ErrorPrint("failed to close results registration", err)
			log.Errorf("could not finalize registering evaluation results: %s", err)
			return err
		}
		log.Infof("Collector %s processing closed", c.Name())
	}
	log.Info("Cluster review finished")
	p.out.Printf("%s %s\n",
		outputs.IconInfo,
		consoleInfoColorF("Cluster review finished"),
	)
	return nil
}

func getClusterID(c config.ConfigCluster) (string, error) {
	if c.ID != "" {
		return c.ID, nil
	}
	if c.Name != "" && c.Location != "" && c.Project != "" {
		return fmt.Sprintf("projects/%s/locations/%s/clusters/%s", c.Project, c.Location, c.Name), nil
	}
	return "", fmt.Errorf("cluster mandatory parameters not set (project, name, location)")
}
