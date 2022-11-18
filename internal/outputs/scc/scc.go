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
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	scc "cloud.google.com/go/securitycenter/apiv1"
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/version"
	gax "github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	sccpb "google.golang.org/genproto/googleapis/cloud/securitycenter/v1"
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
	apiFinding := mapFindingToAPI(sourceName, finding)
	curFinding, err := c.getFinding(apiFinding.Parent, apiFinding.Name)
	if err != nil {
		return err
	}
	if curFinding == nil && finding.State == FINDING_STATE_STRING_INACTIVE {
		log.Debugf("Skipping inactive finding that does not exist in SCC")
		return nil
	}
	if _, err := c.upsertFinding(apiFinding); err != nil {
		return err
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

// getFinding returns the finding for a given source with a given name.
// When not found, nil is returned.
func (c *securityCommandCenterClientImpl) getFinding(source, name string) (*sccpb.Finding, error) {
	req := &sccpb.ListFindingsRequest{
		Parent: source,
		Filter: fmt.Sprintf("name=%q", name),
	}
	it := c.client.ListFindings(c.ctx, req)
	results, err := resourceIteratorToSlice[*sccpb.ListFindingsResponse_ListFindingsResult](
		it,
		c.sourcesSearchLimit,
		func(lfr *sccpb.ListFindingsResponse_ListFindingsResult) bool { return true })
	if err != nil {
		return nil, err
	}
	if len(results) < 1 {
		log.Debugf("No finding for source = %v with name %v found", source, name)
		return nil, nil
	}
	if len(results) > 1 {
		log.Warnf("Multiple findings for source = %v with name %v found", source, name)
	}
	return results[0].Finding, nil
}

// upsertFinding creates or updates given SCC finding using patch operation.
// In case of update operation, all mutable fields are replaced.
func (c *securityCommandCenterClientImpl) upsertFinding(finding *sccpb.Finding) (*sccpb.Finding, error) {
	req := &sccpb.UpdateFindingRequest{
		Finding: finding,
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
	result["GKEPolicyAutomationVersion"] = structpb.NewStringValue(version.Version)

	if finding.CisID != "" && finding.CisVersion != "" {
		standards := map[string]interface{}{
			"cis_gke": []interface{}{
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
		Standard: "cis_gke",
		Version:  finding.CisVersion,
		Ids:      []string{finding.CisID}},
	}
}

// calculateFindingID generates identifier (hash) from resource name and finding category
func calculateFindingID(resourceName, findingCategory string) string {
	val := resourceName + "/" + findingCategory
	hash := md5.Sum([]byte(val))
	return hex.EncodeToString(hash[:])
}

// mapFindingToAPI maps the finding model to finding protobuf struct
func mapFindingToAPI(sourceName string, finding *Finding) *sccpb.Finding {
	name := fmt.Sprintf("%s/findings/%s", sourceName, calculateFindingID(finding.ResourceName, finding.Category))
	return &sccpb.Finding{
		Name:             name,
		Parent:           sourceName,
		Description:      finding.Description,
		ResourceName:     finding.ResourceName,
		State:            mapFindingStateString(finding.State),
		Category:         finding.Category,
		Severity:         mapFindingSeverityString(finding.Severity),
		FindingClass:     sccpb.Finding_MISCONFIGURATION,
		EventTime:        timestamppb.New(finding.Time),
		SourceProperties: mapFindingSourceProperties(finding),
		Compliances:      mapFindingCompliances(finding),
		ExternalUri:      finding.ExternalURI,
	}
}

// mapFindingStateString maps state string to SCC protobuf state int32
func mapFindingStateString(state string) sccpb.Finding_State {
	switch state {
	case FINDING_STATE_STRING_ACTIVE:
		return sccpb.Finding_ACTIVE
	case FINDING_STATE_STRING_INACTIVE:
		return sccpb.Finding_INACTIVE
	default:
		return sccpb.Finding_STATE_UNSPECIFIED
	}
}
