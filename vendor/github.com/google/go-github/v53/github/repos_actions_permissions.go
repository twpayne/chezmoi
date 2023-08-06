// Copyright 2022 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
)

// ActionsPermissionsRepository represents a policy for repositories and allowed actions in a repository.
//
// GitHub API docs: https://docs.github.com/en/rest/actions/permissions
type ActionsPermissionsRepository struct {
	Enabled            *bool   `json:"enabled,omitempty"`
	AllowedActions     *string `json:"allowed_actions,omitempty"`
	SelectedActionsURL *string `json:"selected_actions_url,omitempty"`
}

func (a ActionsPermissionsRepository) String() string {
	return Stringify(a)
}

// GetActionsPermissions gets the GitHub Actions permissions policy for repositories and allowed actions in a repository.
//
// GitHub API docs: https://docs.github.com/en/rest/actions/permissions#get-github-actions-permissions-for-a-repository
func (s *RepositoriesService) GetActionsPermissions(ctx context.Context, owner, repo string) (*ActionsPermissionsRepository, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/actions/permissions", owner, repo)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	permissions := new(ActionsPermissionsRepository)
	resp, err := s.client.Do(ctx, req, permissions)
	if err != nil {
		return nil, resp, err
	}

	return permissions, resp, nil
}

// EditActionsPermissions sets the permissions policy for repositories and allowed actions in a repository.
//
// GitHub API docs: https://docs.github.com/en/rest/actions/permissions#set-github-actions-permissions-for-a-repository
func (s *RepositoriesService) EditActionsPermissions(ctx context.Context, owner, repo string, actionsPermissionsRepository ActionsPermissionsRepository) (*ActionsPermissionsRepository, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/actions/permissions", owner, repo)
	req, err := s.client.NewRequest("PUT", u, actionsPermissionsRepository)
	if err != nil {
		return nil, nil, err
	}

	permissions := new(ActionsPermissionsRepository)
	resp, err := s.client.Do(ctx, req, permissions)
	if err != nil {
		return nil, resp, err
	}

	return permissions, resp, nil
}
