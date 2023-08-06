// Copyright 2020 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// CodeScanningService handles communication with the code scanning related
// methods of the GitHub API.
//
// GitHub API docs: https://docs.github.com/en/rest/code-scanning
type CodeScanningService service

// Rule represents the complete details of GitHub Code Scanning alert type.
type Rule struct {
	ID                    *string  `json:"id,omitempty"`
	Severity              *string  `json:"severity,omitempty"`
	Description           *string  `json:"description,omitempty"`
	Name                  *string  `json:"name,omitempty"`
	SecuritySeverityLevel *string  `json:"security_severity_level,omitempty"`
	FullDescription       *string  `json:"full_description,omitempty"`
	Tags                  []string `json:"tags,omitempty"`
	Help                  *string  `json:"help,omitempty"`
}

// Location represents the exact location of the GitHub Code Scanning Alert in the scanned project.
type Location struct {
	Path        *string `json:"path,omitempty"`
	StartLine   *int    `json:"start_line,omitempty"`
	EndLine     *int    `json:"end_line,omitempty"`
	StartColumn *int    `json:"start_column,omitempty"`
	EndColumn   *int    `json:"end_column,omitempty"`
}

// Message is a part of MostRecentInstance struct which provides the appropriate message when any action is performed on the analysis object.
type Message struct {
	Text *string `json:"text,omitempty"`
}

// MostRecentInstance provides details of the most recent instance of this alert for the default branch or for the specified Git reference.
type MostRecentInstance struct {
	Ref             *string   `json:"ref,omitempty"`
	AnalysisKey     *string   `json:"analysis_key,omitempty"`
	Environment     *string   `json:"environment,omitempty"`
	State           *string   `json:"state,omitempty"`
	CommitSHA       *string   `json:"commit_sha,omitempty"`
	Message         *Message  `json:"message,omitempty"`
	Location        *Location `json:"location,omitempty"`
	Classifications []string  `json:"classifications,omitempty"`
}

// Tool represents the tool used to generate a GitHub Code Scanning Alert.
type Tool struct {
	Name    *string `json:"name,omitempty"`
	GUID    *string `json:"guid,omitempty"`
	Version *string `json:"version,omitempty"`
}

// Alert represents an individual GitHub Code Scanning Alert on a single repository.
//
// GitHub API docs: https://docs.github.com/en/rest/code-scanning
type Alert struct {
	Number             *int                  `json:"number,omitempty"`
	Repository         *Repository           `json:"repository,omitempty"`
	RuleID             *string               `json:"rule_id,omitempty"`
	RuleSeverity       *string               `json:"rule_severity,omitempty"`
	RuleDescription    *string               `json:"rule_description,omitempty"`
	Rule               *Rule                 `json:"rule,omitempty"`
	Tool               *Tool                 `json:"tool,omitempty"`
	CreatedAt          *Timestamp            `json:"created_at,omitempty"`
	UpdatedAt          *Timestamp            `json:"updated_at,omitempty"`
	FixedAt            *Timestamp            `json:"fixed_at,omitempty"`
	State              *string               `json:"state,omitempty"`
	ClosedBy           *User                 `json:"closed_by,omitempty"`
	ClosedAt           *Timestamp            `json:"closed_at,omitempty"`
	URL                *string               `json:"url,omitempty"`
	HTMLURL            *string               `json:"html_url,omitempty"`
	MostRecentInstance *MostRecentInstance   `json:"most_recent_instance,omitempty"`
	Instances          []*MostRecentInstance `json:"instances,omitempty"`
	DismissedBy        *User                 `json:"dismissed_by,omitempty"`
	DismissedAt        *Timestamp            `json:"dismissed_at,omitempty"`
	DismissedReason    *string               `json:"dismissed_reason,omitempty"`
	DismissedComment   *string               `json:"dismissed_comment,omitempty"`
	InstancesURL       *string               `json:"instances_url,omitempty"`
}

// ID returns the ID associated with an alert. It is the number at the end of the security alert's URL.
func (a *Alert) ID() int64 {
	if a == nil {
		return 0
	}

	s := a.GetHTMLURL()

	// Check for an ID to parse at the end of the url
	if i := strings.LastIndex(s, "/"); i >= 0 {
		s = s[i+1:]
	}

	// Return the alert ID as a 64-bit integer. Unable to convert or out of range returns 0.
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}

	return id
}

// AlertListOptions specifies optional parameters to the CodeScanningService.ListAlerts
// method.
type AlertListOptions struct {
	// State of the code scanning alerts to list. Set to closed to list only closed code scanning alerts. Default: open
	State string `url:"state,omitempty"`

	// Return code scanning alerts for a specific branch reference. The ref must be formatted as heads/<branch name>.
	Ref string `url:"ref,omitempty"`

	ListCursorOptions

	// Add ListOptions so offset pagination with integer type "page" query parameter is accepted
	// since ListCursorOptions accepts "page" as string only.
	ListOptions
}

// AnalysesListOptions specifies optional parameters to the CodeScanningService.ListAnalysesForRepo method.
type AnalysesListOptions struct {
	// Return code scanning analyses belonging to the same SARIF upload.
	SarifID *string `url:"sarif_id,omitempty"`

	// Return code scanning analyses for a specific branch reference. The ref can be formatted as refs/heads/<branch name> or simply <branch name>.
	Ref *string `url:"ref,omitempty"`

	ListOptions
}

// ScanningAnalysis represents an individual GitHub Code Scanning ScanningAnalysis on a single repository.
//
// GitHub API docs: https://docs.github.com/en/rest/code-scanning
type ScanningAnalysis struct {
	ID           *int64     `json:"id,omitempty"`
	Ref          *string    `json:"ref,omitempty"`
	CommitSHA    *string    `json:"commit_sha,omitempty"`
	AnalysisKey  *string    `json:"analysis_key,omitempty"`
	Environment  *string    `json:"environment,omitempty"`
	Error        *string    `json:"error,omitempty"`
	Category     *string    `json:"category,omitempty"`
	CreatedAt    *Timestamp `json:"created_at,omitempty"`
	ResultsCount *int       `json:"results_count,omitempty"`
	RulesCount   *int       `json:"rules_count,omitempty"`
	URL          *string    `json:"url,omitempty"`
	SarifID      *string    `json:"sarif_id,omitempty"`
	Tool         *Tool      `json:"tool,omitempty"`
	Deletable    *bool      `json:"deletable,omitempty"`
	Warning      *string    `json:"warning,omitempty"`
}

// SarifAnalysis specifies the results of a code scanning job.
//
// GitHub API docs: https://docs.github.com/en/rest/code-scanning
type SarifAnalysis struct {
	CommitSHA   *string    `json:"commit_sha,omitempty"`
	Ref         *string    `json:"ref,omitempty"`
	Sarif       *string    `json:"sarif,omitempty"`
	CheckoutURI *string    `json:"checkout_uri,omitempty"`
	StartedAt   *Timestamp `json:"started_at,omitempty"`
	ToolName    *string    `json:"tool_name,omitempty"`
}

// CodeScanningAlertState specifies the state of a code scanning alert.
//
// GitHub API docs: https://docs.github.com/en/rest/code-scanning
type CodeScanningAlertState struct {
	// State sets the state of the code scanning alert and is a required field.
	// You must also provide DismissedReason when you set the state to "dismissed".
	// State can be one of: "open", "dismissed".
	State string `json:"state"`
	// DismissedReason represents the reason for dismissing or closing the alert.
	// It is required when the state is "dismissed".
	// It can be one of: "false positive", "won't fix", "used in tests".
	DismissedReason *string `json:"dismissed_reason,omitempty"`
	// DismissedComment is associated with the dismissal of the alert.
	DismissedComment *string `json:"dismissed_comment,omitempty"`
}

// SarifID identifies a sarif analysis upload.
//
// GitHub API docs: https://docs.github.com/en/rest/code-scanning
type SarifID struct {
	ID  *string `json:"id,omitempty"`
	URL *string `json:"url,omitempty"`
}

// ListAlertsForOrg lists code scanning alerts for an org.
//
// You must use an access token with the security_events scope to use this endpoint. GitHub Apps must have the security_events
// read permission to use this endpoint.
//
// GitHub API docs: https://docs.github.com/en/rest/code-scanning#list-code-scanning-alerts-for-an-organization
func (s *CodeScanningService) ListAlertsForOrg(ctx context.Context, org string, opts *AlertListOptions) ([]*Alert, *Response, error) {
	u := fmt.Sprintf("orgs/%v/code-scanning/alerts", org)
	u, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var alerts []*Alert
	resp, err := s.client.Do(ctx, req, &alerts)
	if err != nil {
		return nil, resp, err
	}

	return alerts, resp, nil
}

// ListAlertsForRepo lists code scanning alerts for a repository.
//
// Lists all open code scanning alerts for the default branch (usually master) and protected branches in a repository.
// You must use an access token with the security_events scope to use this endpoint. GitHub Apps must have the security_events
// read permission to use this endpoint.
//
// GitHub API docs: https://docs.github.com/en/rest/code-scanning#list-code-scanning-alerts-for-a-repository
func (s *CodeScanningService) ListAlertsForRepo(ctx context.Context, owner, repo string, opts *AlertListOptions) ([]*Alert, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/code-scanning/alerts", owner, repo)
	u, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var alerts []*Alert
	resp, err := s.client.Do(ctx, req, &alerts)
	if err != nil {
		return nil, resp, err
	}

	return alerts, resp, nil
}

// GetAlert gets a single code scanning alert for a repository.
//
// You must use an access token with the security_events scope to use this endpoint.
// GitHub Apps must have the security_events read permission to use this endpoint.
//
// The security alert_id is the number at the end of the security alert's URL.
//
// GitHub API docs: https://docs.github.com/en/rest/code-scanning#get-a-code-scanning-alert
func (s *CodeScanningService) GetAlert(ctx context.Context, owner, repo string, id int64) (*Alert, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/code-scanning/alerts/%v", owner, repo, id)

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	a := new(Alert)
	resp, err := s.client.Do(ctx, req, a)
	if err != nil {
		return nil, resp, err
	}

	return a, resp, nil
}

// UpdateAlert updates the state of a single code scanning alert for a repository.
//
// You must use an access token with the security_events scope to use this endpoint.
// GitHub Apps must have the security_events read permission to use this endpoint.
//
// The security alert_id is the number at the end of the security alert's URL.
//
// GitHub API docs: https://docs.github.com/en/rest/code-scanning?apiVersion=2022-11-28#update-a-code-scanning-alert
func (s *CodeScanningService) UpdateAlert(ctx context.Context, owner, repo string, id int64, stateInfo *CodeScanningAlertState) (*Alert, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/code-scanning/alerts/%v", owner, repo, id)

	req, err := s.client.NewRequest("PATCH", u, stateInfo)
	if err != nil {
		return nil, nil, err
	}

	a := new(Alert)
	resp, err := s.client.Do(ctx, req, a)
	if err != nil {
		return nil, resp, err
	}

	return a, resp, nil
}

// UploadSarif uploads the result of code scanning job to GitHub.
//
// For the parameter sarif, you must first compress your SARIF file using gzip and then translate the contents of the file into a Base64 encoding string.
// You must use an access token with the security_events scope to use this endpoint. GitHub Apps must have the security_events
// write permission to use this endpoint.
//
// GitHub API docs: https://docs.github.com/en/rest/code-scanning#upload-an-analysis-as-sarif-data
func (s *CodeScanningService) UploadSarif(ctx context.Context, owner, repo string, sarif *SarifAnalysis) (*SarifID, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/code-scanning/sarifs", owner, repo)

	req, err := s.client.NewRequest("POST", u, sarif)
	if err != nil {
		return nil, nil, err
	}

	sarifID := new(SarifID)
	resp, err := s.client.Do(ctx, req, sarifID)
	if err != nil {
		return nil, resp, err
	}

	return sarifID, resp, nil
}

// ListAnalysesForRepo lists code scanning analyses for a repository.
//
// Lists the details of all code scanning analyses for a repository, starting with the most recent.
// You must use an access token with the security_events scope to use this endpoint.
// GitHub Apps must have the security_events read permission to use this endpoint.
//
// GitHub API docs: https://docs.github.com/en/rest/code-scanning#list-code-scanning-analyses-for-a-repository
func (s *CodeScanningService) ListAnalysesForRepo(ctx context.Context, owner, repo string, opts *AnalysesListOptions) ([]*ScanningAnalysis, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/code-scanning/analyses", owner, repo)
	u, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var analyses []*ScanningAnalysis
	resp, err := s.client.Do(ctx, req, &analyses)
	if err != nil {
		return nil, resp, err
	}

	return analyses, resp, nil
}

// GetAnalysis gets a single code scanning analysis for a repository.
//
// You must use an access token with the security_events scope to use this endpoint.
// GitHub Apps must have the security_events read permission to use this endpoint.
//
// The security analysis_id is the ID of the analysis, as returned from the ListAnalysesForRepo operation.
//
// GitHub API docs: https://docs.github.com/en/rest/code-scanning#get-a-code-scanning-analysis-for-a-repository
func (s *CodeScanningService) GetAnalysis(ctx context.Context, owner, repo string, id int64) (*ScanningAnalysis, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/code-scanning/analyses/%v", owner, repo, id)

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	analysis := new(ScanningAnalysis)
	resp, err := s.client.Do(ctx, req, analysis)
	if err != nil {
		return nil, resp, err
	}

	return analysis, resp, nil
}

// DefaultSetupConfiguration represents a code scanning default setup configuration.
type DefaultSetupConfiguration struct {
	State      *string    `json:"state,omitempty"`
	Languages  []string   `json:"languages,omitempty"`
	QuerySuite *string    `json:"query_suite,omitempty"`
	UpdatedAt  *Timestamp `json:"updated_at,omitempty"`
}

// GetDefaultSetupConfiguration gets a code scanning default setup configuration.
//
// You must use an access token with the repo scope to use this
// endpoint with private repos or the public_repo scope for public repos. GitHub Apps must have the repo write
// permission to use this endpoint.
//
// GitHub API docs: https://docs.github.com/en/rest/code-scanning#get-a-code-scanning-default-setup-configuration
func (s *CodeScanningService) GetDefaultSetupConfiguration(ctx context.Context, owner, repo string) (*DefaultSetupConfiguration, *Response, error) {
	u := fmt.Sprintf("repos/%s/%s/code-scanning/default-setup", owner, repo)

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	cfg := new(DefaultSetupConfiguration)
	resp, err := s.client.Do(ctx, req, cfg)
	if err != nil {
		return nil, resp, err
	}

	return cfg, resp, nil
}

// UpdateDefaultSetupConfigurationOptions specifies parameters to the CodeScanningService.UpdateDefaultSetupConfiguration
// method.
type UpdateDefaultSetupConfigurationOptions struct {
	State      string   `json:"state"`
	QuerySuite *string  `json:"query_suite,omitempty"`
	Languages  []string `json:"languages,omitempty"`
}

// UpdateDefaultSetupConfigurationResponse represents a response from updating a code scanning default setup configuration.
type UpdateDefaultSetupConfigurationResponse struct {
	RunID  *int64  `json:"run_id,omitempty"`
	RunURL *string `json:"run_url,omitempty"`
}

// UpdateDefaultSetupConfiguration updates a code scanning default setup configuration.
//
// You must use an access token with the repo scope to use this
// endpoint with private repos or the public_repo scope for public repos. GitHub Apps must have the repo write
// permission to use this endpoint.
//
// This method might return an AcceptedError and a status code of 202. This is because this is the status that GitHub
// returns to signify that it has now scheduled the update of the pull request branch in a background task.
//
// GitHub API docs: https://docs.github.com/en/rest/code-scanning#update-a-code-scanning-default-setup-configuration
func (s *CodeScanningService) UpdateDefaultSetupConfiguration(ctx context.Context, owner, repo string, options *UpdateDefaultSetupConfigurationOptions) (*UpdateDefaultSetupConfigurationResponse, *Response, error) {
	u := fmt.Sprintf("repos/%s/%s/code-scanning/default-setup", owner, repo)

	req, err := s.client.NewRequest("PATCH", u, options)
	if err != nil {
		return nil, nil, err
	}

	a := new(UpdateDefaultSetupConfigurationResponse)
	resp, err := s.client.Do(ctx, req, a)
	if err != nil {
		return nil, resp, err
	}

	return a, resp, nil
}
