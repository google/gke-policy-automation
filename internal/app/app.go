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

// Package app implements application management features
package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"

	cfg "github.com/google/gke-policy-automation/internal/config"
	"github.com/google/gke-policy-automation/internal/gke"
	"github.com/google/gke-policy-automation/internal/inputs"
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/outputs"
	"github.com/google/gke-policy-automation/internal/policy"
	"github.com/google/gke-policy-automation/internal/version"
	"golang.org/x/exp/maps"
)

const (
	regoPackageBaseBestPractices = "gke.policy"
	regoPackageBaseScalability   = "gke.scalability"
)

var errNoPolicies = errors.New("no policies to check against")

type PolicyAutomation interface {
	LoadCliConfig(cliConfig *CliConfig, defaultsFn setConfigDefaults, validateFn validateConfig) error
	Close() error
	Check() error
	CheckBestPractices() error
	CheckScalability() error
	ClusterJSONData() error
	Version() error
	PolicyCheck() error
	PolicyGenerateDocumentation() error
	ConfigureSCC(orgNumber string) error
}

type evaluationResults struct {
	m map[string]*policy.PolicyEvaluationResult
}

func (r *evaluationResults) Add(result *policy.PolicyEvaluationResult) *evaluationResults {
	if r.m == nil {
		r.m = make(map[string]*policy.PolicyEvaluationResult)
	}
	currentResult, ok := r.m[result.ClusterID]
	if !ok {
		r.m[result.ClusterID] = result
		return r
	}
	currentResult.Policies = append(currentResult.Policies, result.Policies...)
	return r
}

func (r *evaluationResults) List() []*policy.PolicyEvaluationResult {
	return maps.Values(r.m)
}

type PolicyAutomationApp struct {
	ctx                   context.Context
	config                *cfg.Config
	out                   *outputs.Output
	inputs                []inputs.Input
	collectors            []outputs.ValidationResultCollector
	clusterDumpCollectors []outputs.ClusterDumpCollector
	discovery             gke.DiscoveryClient
	policyDocsFile        string
}

func NewPolicyAutomationApp() PolicyAutomation {
	out := outputs.NewSilentOutput()
	return &PolicyAutomationApp{
		ctx:        context.Background(),
		config:     &cfg.Config{},
		out:        out,
		collectors: []outputs.ValidationResultCollector{outputs.NewConsoleResultCollector(out)},
	}
}

func (p *PolicyAutomationApp) Close() error {
	errors := make([]error, 0)
	for _, i := range p.inputs {
		if err := i.Close(); err != nil {
			log.Warnf("error when closing input %s: %s", i.GetID(), err)
			errors = append(errors, err)
		}
	}
	if p.discovery != nil {
		if err := p.discovery.Close(); err != nil {
			log.Warnf("error when closing discovery client: %s", err)
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

func (p *PolicyAutomationApp) Check() error {
	return p.evaluateClusters([]string{regoPackageBaseBestPractices})
}

func (p *PolicyAutomationApp) CheckBestPractices() error {
	return p.evaluateClusters([]string{regoPackageBaseBestPractices})
}

func (p *PolicyAutomationApp) CheckScalability() error {
	p.out.ColorPrintf("%s [yellow][bold]Running scalability check requires metrics from kube-state-metrics!\n", outputs.IconInfo)
	docsTitle := fmt.Sprintf("%s \x1b]8;;%s\x07%s\x1b]8;;\x07", outputs.IconHyperlink, "https://github.com/google/gke-policy-automation/blob/scalability-docs/docs/user-guide.md#checking-scalability-limits", "documentation")
	p.out.ColorPrintf("%s [yellow][bold]Check the %s for more details.\n", outputs.IconInfo, docsTitle)
	return p.evaluateClusters([]string{regoPackageBaseScalability})
}

func (p *PolicyAutomationApp) ClusterJSONData() error {
	clusterIds, err := p.getClusters()
	if err != nil {
		p.out.ErrorPrint("could not get clusters", err)
		log.Errorf("could not get clusters: %s", err)
	}

	//for dumping JSON data - create gkeInput
	var gkeInput inputs.Input
	if p.config.CredentialsFile != "" {
		gkeInput, err = inputs.NewGKEApiInputWithCredentials(p.ctx, p.config.CredentialsFile)
		if err != nil {
			return err
		}
	} else {
		gkeInput, err = inputs.NewGKEApiInput(p.ctx)
		if err != nil {
			return err
		}
	}
	p.inputs = append(p.inputs, gkeInput)

	clusterData, errors := inputs.GetAllInputsData(p.inputs, clusterIds)
	if len(errors) > 0 {
		p.out.ErrorPrint("could not fetch the cluster details", errors[0])
		log.Errorf("could not fetch cluster details: %s", errors[0])
		return errors[0]
	}
	val, err := json.MarshalIndent(clusterData, "", "    ")
	log.Debugf("[DEBUG] cluster: " + string(val))

	for _, cluster := range clusterData {

		if err != nil {
			p.out.ErrorPrint("could not fetch the cluster details", err)
			log.Errorf("could not fetch cluster details: %s", err)
			return err
		}
		val, err := json.MarshalIndent(cluster, "", "    ")
		if err != nil {
			log.Debugf("could not format cluster details: %s", err)
		}
		log.Debugf("cluster: " + string(val))

		for _, dumpCollector := range p.clusterDumpCollectors {
			log.Debugf("registering cluster data with cluster dump collector %s", reflect.TypeOf(dumpCollector).String())
			dumpCollector.RegisterCluster(cluster)
		}
	}
	for _, dumpCollector := range p.clusterDumpCollectors {
		colType := reflect.TypeOf(dumpCollector).String()
		log.Debugf("closing cluster dump collector %s", colType)
		p.out.ColorPrintf("%s [light_gray][bold]closing cluster dump collector ...\n", outputs.IconInfo)
		if err := dumpCollector.Close(); err != nil {
			log.Errorf("failed to close cluster dump collector %s due to %s", colType, err)
			return err
		}
	}

	return nil
}

func (p *PolicyAutomationApp) Version() error {
	p.out.Printf("%s\n", version.Version)
	return nil
}

func (p *PolicyAutomationApp) PolicyCheck() error {
	files, err := p.loadPolicyFiles()
	if err != nil {
		p.out.ErrorPrint("loading policy files failed: ", err)
		log.Errorf("loading policy files failed: %s", err)
		return err
	}
	pa := policy.NewPolicyAgent(p.ctx)
	if err := pa.WithFiles(files, p.config.PolicyExclusions); err != nil {
		p.out.ErrorPrint("could not parse policy files", err)
		log.Errorf("could not parse policy files: %s", err)
		return err
	}
	p.out.ColorPrintf("%s [bold][green] All policies validated correctly\n", outputs.IconInfo)
	log.Info("All policies validated correctly")
	return nil
}

func (p *PolicyAutomationApp) PolicyGenerateDocumentation() error {
	w, err := os.OpenFile(p.policyDocsFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		p.out.ErrorPrint("could not open output file for writing: ", err)
		log.Errorf("could not open output file for writing: %s", err)
		return err
	}
	defer w.Close()

	files, err := p.loadPolicyFiles()
	if err != nil {
		p.out.ErrorPrint("loading policy files failed: ", err)
		log.Errorf("loading policy files failed: %s", err)
		return err
	}

	pa := policy.NewPolicyAgent(p.ctx)
	if err := pa.WithFiles(files, p.config.PolicyExclusions); err != nil {
		p.out.ErrorPrint("could not parse policy files", err)
		log.Errorf("could not parse policy files: %s", err)
		return err
	}

	documentationGenerator := outputs.NewMarkdownPolicyDocumentation(pa.GetPolicies())
	p.out.ColorPrintf("%s [light_gray][bold]Writing policy documentation ... [%s]\n", outputs.IconInfo, p.policyDocsFile)
	log.Infof("Writing policy documentation to file %s", p.policyDocsFile)
	if _, err := w.Write([]byte(documentationGenerator.GenerateDocumentation())); err != nil {
		p.out.ErrorPrint("could not write documentation file", err)
		log.Errorf("could not write documentation file: %s", err)
		return err
	}
	return nil
}

func (p *PolicyAutomationApp) loadPolicyFiles() ([]*policy.PolicyFile, error) {
	policyFiles := make([]*policy.PolicyFile, 0)
	for _, policyConfig := range p.config.Policies {
		var policySrc policy.PolicySource
		if policyConfig.LocalDirectory != "" {
			policySrc = policy.NewLocalPolicySource(policyConfig.LocalDirectory)
		}
		if policyConfig.GitRepository != "" {
			policySrc = policy.NewGitPolicySource(policyConfig.GitRepository,
				policyConfig.GitBranch,
				policyConfig.GitDirectory)
		}
		p.out.ColorPrintf("%s [light_gray][bold]Reading policy files... [%s]\n", outputs.IconInfo, policySrc)
		log.Infof("Reading policy files from %s", policySrc)
		files, err := policySrc.GetPolicyFiles()
		if err != nil {
			p.out.ErrorPrint("could not read policy files", err)
			log.Errorf("could not read policy files: %s", err)
			return nil, err
		}
		policyFiles = append(policyFiles, files...)
	}
	return policyFiles, nil
}
