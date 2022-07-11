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

package scc

import (
	"context"
	"fmt"
	"time"

	scc "cloud.google.com/go/securitycenter/apiv1"
	"github.com/dchest/uniuri"
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/version"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	sccpb "google.golang.org/genproto/googleapis/cloud/securitycenter/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	sourceDisplayName        = "GKE Policy Automation"
	sourceDescription        = "Validates GKE clusters against configuration best practices and scalability limits"
	defaultSourceSearchLimit = 1000

	FINDING_STATE_STRING_ACTIVE      = "ACTIVE"
	FINDING_STATE_STRING_INACTIVE    = "INACTIVE"
	FINDING_STATE_STRING_UNSPECIFIED = "UNSPECIFIED"
)

type SecurityCommandCenterClient interface {
	CreateSource() (string, error)
	FindSource() (*string, error)
	UpsertFinding(sourceName string, finding *Finding) (string, error)
}

type sccResource interface {
	*sccpb.Source | *sccpb.ListFindingsResponse_ListFindingsResult
}

type sccResourceIterator[R sccResource] interface {
	Next() (R, error)
}

type sccResourceFilter[R sccResource] func(R) bool

type sccSourceIterator interface {
	Next() (*sccpb.Source, error)
}

type Finding struct {
	Time         time.Time
	SourceName   string
	ResourceName string
	Category     string
	Description  string
	State        string
}

type securityCommandCenterClientImpl struct {
	ctx                context.Context
	organizationNumber string
	sourcesSearchLimit int
	client             *scc.Client
}

func NewSecurityCommandCenterClient(ctx context.Context, organizationNumber string) (SecurityCommandCenterClient, error) {
	opts := []option.ClientOption{option.WithUserAgent(version.UserAgent)}
	c, err := scc.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &securityCommandCenterClientImpl{
		ctx:                ctx,
		organizationNumber: organizationNumber,
		sourcesSearchLimit: defaultSourceSearchLimit,
		client:             c,
	}, nil
}

func (c *securityCommandCenterClientImpl) CreateSource() (string, error) {
	req := &sccpb.CreateSourceRequest{
		Parent: "organizations/" + c.organizationNumber,
		Source: &sccpb.Source{
			DisplayName: sourceDisplayName,
			Description: sourceDescription,
		},
	}
	source, err := c.client.CreateSource(c.ctx, req)
	if err != nil {
		return "", err
	}
	return source.Name, nil
}

func (c *securityCommandCenterClientImpl) FindSource() (*string, error) {
	listReq := &sccpb.ListSourcesRequest{
		Parent: "organizations/" + c.organizationNumber,
	}
	sourcesIterator := c.client.ListSources(c.ctx, listReq)
	return c.findSourceNameByDisplayName(sourceDisplayName, sourcesIterator)
}

func (c *securityCommandCenterClientImpl) UpsertFinding(sourceName string, finding *Finding) (string, error) {
	req := &sccpb.ListFindingsRequest{
		Parent: finding.SourceName,
		Filter: fmt.Sprintf("resourceName=%q AND category=%q AND state=%q", finding.ResourceName, finding.Category, FINDING_STATE_STRING_ACTIVE),
	}
	it := c.client.ListFindings(c.ctx, req)
	findings, err := resourceIteratorToSlice[*sccpb.ListFindingsResponse_ListFindingsResult](
		it,
		c.sourcesSearchLimit,
		func(lfr *sccpb.ListFindingsResponse_ListFindingsResult) bool { return true })
	if err != nil {
		return "", err
	}
	if len(findings) > 0 {
		//update finding
	}
	if finding.State == FINDING_STATE_STRING_ACTIVE {
		return c.createFinding(sourceName, finding)
	}
	return "", nil
}

func (c *securityCommandCenterClientImpl) findSourceNameByDisplayName(displayName string, it sccSourceIterator) (*string, error) {
	results, err := resourceIteratorToSlice[*sccpb.Source](it, c.sourcesSearchLimit, func(s *sccpb.Source) bool {
		return s.DisplayName == displayName
	})
	if err != nil {
		return nil, err
	}
	if len(results) < 1 {
		return nil, nil
	}
	if len(results) > 1 {
		log.Warnf("found more than one GKE Policy Automation SCC source")
	}
	name := results[0].Name
	return &name, nil
}

func (c *securityCommandCenterClientImpl) createFinding(sourceName string, finding *Finding) (string, error) {
	req := &sccpb.CreateFindingRequest{
		Parent:    sourceName,
		FindingId: uniuri.NewLen(32),
		Finding: &sccpb.Finding{
			Description:  finding.Description,
			ResourceName: finding.ResourceName,
			State:        sccpb.Finding_ACTIVE,
			Category:     finding.Category,
			Severity:     sccpb.Finding_HIGH, //mapping here
			FindingClass: sccpb.Finding_MISCONFIGURATION,
			EventTime:    timestamppb.New(finding.Time),
		},
	}
	result, err := c.client.CreateFinding(c.ctx, req)
	if err != nil {
		return "", err
	}
	return result.Name, nil
}

func resourceIteratorToSlice[R sccResource](it sccResourceIterator[R], limit int, filter sccResourceFilter[R]) ([]R, error) {
	results := make([]R, 0)
	i := 0
	for ; i < limit; i++ {
		result, err := it.Next()
		if err == iterator.Done {
			log.Debugf("search iterator done")
			break
		}
		if err != nil {
			return nil, err
		}
		log.Debugf("search iterator result: %+v", result)
		if filter(result) {
			results = append(results, result)
		} else {
			log.Debugf("filtering out result: %+v", result)
		}
	}
	if i == limit {
		log.Warnf("search limit of %d was reached", limit)
	}
	return results, nil
}

/*
	listReq := &sccpb.ListSourcesRequest{
		//Parent: "organizations/153963171798",
		//Parent: "projects/gke-policy-demo",
		Parent: "folders/426539704670",
	}
	sources := c.ListSources(ctx, listReq)
	for i := 0; i < 10; i++ {
		result, err := sources.Next()
		if err == iterator.Done {
			log.Debugf("search iterator done")
			break
		}
		if err != nil {
			return nil, err
		}
		log.Debugf("search iterator result: %s", result)
	}
	return c, nil
}
*/
