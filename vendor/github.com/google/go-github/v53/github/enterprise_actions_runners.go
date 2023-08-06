// Copyright 2020 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
)

// ListRunnerApplicationDownloads lists self-hosted runner application binaries that can be downloaded and run.
//
// GitHub API docs: https://docs.github.com/en/rest/actions/self-hosted-runners#list-runner-applications-for-an-enterprise
func (s *EnterpriseService) ListRunnerApplicationDownloads(ctx context.Context, enterprise string) ([]*RunnerApplicationDownload, *Response, error) {
	u := fmt.Sprintf("enterprises/%v/actions/runners/downloads", enterprise)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var rads []*RunnerApplicationDownload
	resp, err := s.client.Do(ctx, req, &rads)
	if err != nil {
		return nil, resp, err
	}

	return rads, resp, nil
}

// CreateRegistrationToken creates a token that can be used to add a self-hosted runner.
//
// GitHub API docs: https://docs.github.com/en/rest/actions/self-hosted-runners#create-a-registration-token-for-an-enterprise
func (s *EnterpriseService) CreateRegistrationToken(ctx context.Context, enterprise string) (*RegistrationToken, *Response, error) {
	u := fmt.Sprintf("enterprises/%v/actions/runners/registration-token", enterprise)

	req, err := s.client.NewRequest("POST", u, nil)
	if err != nil {
		return nil, nil, err
	}

	registrationToken := new(RegistrationToken)
	resp, err := s.client.Do(ctx, req, registrationToken)
	if err != nil {
		return nil, resp, err
	}

	return registrationToken, resp, nil
}

// ListRunners lists all the self-hosted runners for a enterprise.
//
// GitHub API docs: https://docs.github.com/en/rest/actions/self-hosted-runners#list-self-hosted-runners-for-an-enterprise
func (s *EnterpriseService) ListRunners(ctx context.Context, enterprise string, opts *ListOptions) (*Runners, *Response, error) {
	u := fmt.Sprintf("enterprises/%v/actions/runners", enterprise)
	u, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	runners := &Runners{}
	resp, err := s.client.Do(ctx, req, &runners)
	if err != nil {
		return nil, resp, err
	}

	return runners, resp, nil
}

// RemoveRunner forces the removal of a self-hosted runner from an enterprise using the runner id.
//
// GitHub API docs: https://docs.github.com/en/rest/actions/self-hosted-runners#delete-a-self-hosted-runner-from-an-enterprise
func (s *EnterpriseService) RemoveRunner(ctx context.Context, enterprise string, runnerID int64) (*Response, error) {
	u := fmt.Sprintf("enterprises/%v/actions/runners/%v", enterprise, runnerID)

	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}
