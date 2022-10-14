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
	"errors"
	"fmt"
	"strings"
	"time"

	scc "cloud.google.com/go/securitycenter/apiv1"
	"github.com/dchest/uniuri"
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/version"
	gax "github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	sccpb "google.golang.org/genproto/googleapis/cloud/securitycenter/v1"
	fieldmaskpb "google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	sourceDisplayName        = "GKE Policy Automation"
	sourceDescription        = "Validates GKE clusters against configuration best practices and scalability limits"
	defaultSourceSearchLimit = 1000

	FINDING_STATE_STRING_ACTIVE      = "ACTIVE"
	FINDING_STATE_STRING_INACTIVE    = "INACTIVE"
	FINDING_STATE_STRING_UNSPECIFIED = "UNSPECIFIED"

	FINDING_SEVERITY_STRING_CRITICAL = "CRITICAL"
	FINDING_SEVERITY_STRING_HIGH     = "HIGH"
	FINDING_SEVERITY_STRING_MEDIUM   = "MEDIUM"
	FINDING_SEVERITY_STRING_LOW      = "LOW"
)

type SecurityCommandCenterClient interface {
	CreateSource() (string, error)
	FindSource() (*string, error)
	UpsertFinding(sourceName string, finding *Finding) error
	Close() error
}

type sccApiClient interface {
	ListSources(ctx context.Context, req *sccpb.ListSourcesRequest, opts ...gax.CallOption) *scc.SourceIterator
	CreateSource(ctx context.Context, req *sccpb.CreateSourceRequest, opts ...gax.CallOption) (*sccpb.Source, error)
	ListFindings(ctx context.Context, req *sccpb.ListFindingsRequest, opts ...gax.CallOption) *scc.ListFindingsResponse_ListFindingsResultIterator
	UpdateFinding(ctx context.Context, req *sccpb.UpdateFindingRequest, opts ...gax.CallOption) (*sccpb.Finding, error)
	Close() error
}

type sccResource interface {
	*sccpb.Source | *sccpb.ListFindingsResponse_ListFindingsResult
}

type sccResourceIterator[R sccResource] interface {
	Next() (R, error)
}

type sccResourceFilter[R sccResource] func(R) bool
type MultipleErrors []error

func (m MultipleErrors) Error() error {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d errors occurred:", len(m))
	for _, err := range m {
		fmt.Fprintf(&sb, " %s;", err)
	}
	return errors.New(sb.String())
}

type Finding struct {
	Time              time.Time
	ResourceName      string
	Category          string
	Description       string
	State             string
	Severity          string
	CisVersion        string
	CisID             string
	SourcePolicyName  string
	SourcePolicyFile  string
	SourcePolicyGroup string
	ExternalURI       string
	Recommendation    string
}

type securityCommandCenterClientImpl struct {
	ctx                context.Context
	organizationNumber string
	sourcesSearchLimit int
	client             sccApiClient
}

func NewSecurityCommandCenterClient(ctx context.Context, organizationNumber string) (SecurityCommandCenterClient, error) {
	return newSecurityCommandCenterClient(ctx, organizationNumber)
}

func NewSecurityCommandCenterClientWithCredentialsFile(ctx context.Context, organizationNumber string, credsFile string) (SecurityCommandCenterClient, error) {
	return newSecurityCommandCenterClient(ctx, organizationNumber, option.WithCredentialsFile(credsFile))
}

func newSecurityCommandCenterClient(ctx context.Context, organizationNumber string, opts ...option.ClientOption) (SecurityCommandCenterClient, error) {
	opts = append(opts, option.WithUserAgent(version.UserAgent))
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

func (c *securityCommandCenterClientImpl) UpsertFinding(sourceName string, finding *Finding) error {
	sccFindings, err := c.getFindings(sourceName, finding.ResourceName, finding.Category)
	if err != nil {
		return err
	}
	if len(sccFindings) < 1 && finding.State == FINDING_STATE_STRING_ACTIVE {
		_, err = c.createFinding(sourceName, finding)
		return err
	}
	if errors := c.updateFindings(sccFindings, finding); len(errors) > 0 {
		return errors.Error()
	}
	return nil
}

func (c *securityCommandCenterClientImpl) Close() error {
	log.Debugf("closing scc client")
	return c.client.Close()
}

func (c *securityCommandCenterClientImpl) findSourceNameByDisplayName(displayName string, it sccResourceIterator[*sccpb.Source]) (*string, error) {
	results, err := resourceIteratorToSlice(it, c.sourcesSearchLimit, func(s *sccpb.Source) bool {
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

// getFindings returns slice of findings for a given SCC source, resource and category.
func (c *securityCommandCenterClientImpl) getFindings(source string, resource string, category string) ([]*sccpb.ListFindingsResponse_ListFindingsResult, error) {
	req := &sccpb.ListFindingsRequest{
		Parent: source,
		Filter: fmt.Sprintf("resourceName=%q AND category=%q", resource, category),
	}
	it := c.client.ListFindings(c.ctx, req)
	return resourceIteratorToSlice[*sccpb.ListFindingsResponse_ListFindingsResult](
		it,
		c.sourcesSearchLimit,
		func(lfr *sccpb.ListFindingsResponse_ListFindingsResult) bool { return true })
}

// createFinding creates given finding under the given source.
func (c *securityCommandCenterClientImpl) createFinding(sourceName string, finding *Finding) (string, error) {
	sccFinding := &sccpb.Finding{
		Name:             fmt.Sprintf("%s/findings/%s", sourceName, uniuri.NewLen(32)),
		Parent:           sourceName,
		Description:      finding.Description,
		ResourceName:     finding.ResourceName,
		State:            sccpb.Finding_ACTIVE,
		Category:         finding.Category,
		Severity:         mapFindingSeverityString(finding.Severity),
		FindingClass:     sccpb.Finding_MISCONFIGURATION,
		EventTime:        timestamppb.New(finding.Time),
		SourceProperties: mapFindingSourceProperties(finding),
		Compliances:      mapFindingCompliances(finding),
		ExternalUri:      finding.ExternalURI,
	}
	sccFinding, err := c.upsertFinding(sccFinding, nil)
	if err != nil {
		return "", err
	}
	return sccFinding.Name, nil
}

// updateFinding updates all findings with a names from given slice
func (c *securityCommandCenterClientImpl) updateFindings(findingListResults []*sccpb.ListFindingsResponse_ListFindingsResult, finding *Finding) MultipleErrors {
	errors := MultipleErrors{}
	for _, result := range findingListResults {
		if err := c.updateFinding(result.Finding, finding); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// updateFinding updates state and event time for a given finding
func (c *securityCommandCenterClientImpl) updateFinding(result *sccpb.Finding, finding *Finding) error {
	result.EventTime = timestamppb.New(finding.Time)
	result.SourceProperties = mapFindingSourceProperties(finding)
	updateMask := []string{"state", "event_time", "source_properties"}
	log.Debugf("updating finding: data %v; updateMask %v", finding, updateMask)
	_, err := c.upsertFinding(result, updateMask)
	return err
}

// upsertFinding creates or updates given SCC finding using patch operation.
// For creation, the given finding should have valid identifier in the name field.
// For update, the updateMaskPaths should be given to indicate fields to be updated.
func (c *securityCommandCenterClientImpl) upsertFinding(finding *sccpb.Finding, updateMaskPaths []string) (*sccpb.Finding, error) {
	req := &sccpb.UpdateFindingRequest{
		Finding: finding,
	}
	if len(updateMaskPaths) > 0 {
		req.UpdateMask = &fieldmaskpb.FieldMask{
			Paths: updateMaskPaths,
		}
	}
	log.Debugf("SCC finding update with req: %+v", req)
	return c.client.UpdateFinding(c.ctx, req)
}

// resourceIteratorToSlice iterates using given resource iterator, up to the given limit, and returns list of resources.
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

// mapFindingSeverityString maps severity string to SCC severity uint32
func mapFindingSeverityString(severity string) sccpb.Finding_Severity {
	switch severity {
	case FINDING_SEVERITY_STRING_CRITICAL:
		return sccpb.Finding_CRITICAL
	case FINDING_SEVERITY_STRING_HIGH:
		return sccpb.Finding_HIGH
	case FINDING_SEVERITY_STRING_MEDIUM:
		return sccpb.Finding_MEDIUM
	case FINDING_SEVERITY_STRING_LOW:
		return sccpb.Finding_LOW
	default:
		return sccpb.Finding_SEVERITY_UNSPECIFIED
	}
}

func mapFindingSourceProperties(finding *Finding) map[string]*structpb.Value {
	result := make(map[string]*structpb.Value)
	result["PolicyName"] = structpb.NewStringValue(finding.SourcePolicyName)
	result["PolicyFile"] = structpb.NewStringValue(finding.SourcePolicyFile)
	result["PolicyGroup"] = structpb.NewStringValue(finding.SourcePolicyGroup)

	result["Recommendation"] = structpb.NewStringValue(finding.Recommendation)

	if finding.CisID != "" && finding.CisVersion != "" {
		standards := map[string]interface{}{
			"cis": []interface{}{
				map[string]interface{}{
					"version": finding.CisVersion,
					"ids":     []interface{}{finding.CisID},
				},
			},
		}
		structValue, err := structpb.NewStruct(standards)
		if err != nil {
			panic("mapping finding sourceProperties failed: cannot construct strucpb struct Value")
		}
		result["compliance_standards"] = structpb.NewStructValue(structValue)
	}
	return result
}

func mapFindingCompliances(finding *Finding) []*sccpb.Compliance {
	if finding.CisID == "" || finding.CisVersion == "" {
		return nil
	}
	return []*sccpb.Compliance{{
		Standard: "cis",
		Version:  finding.CisVersion,
		Ids:      []string{finding.CisID}},
	}
}
