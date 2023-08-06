// Copyright 2022 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
)

// CodeownersErrors represents a list of syntax errors detected in the CODEOWNERS file.
type CodeownersErrors struct {
	Errors []*CodeownersError `json:"errors"`
}

// CodeownersError represents a syntax error detected in the CODEOWNERS file.
type CodeownersError struct {
	Line       int     `json:"line"`
	Column     int     `json:"column"`
	Kind       string  `json:"kind"`
	Source     string  `json:"source"`
	Suggestion *string `json:"suggestion,omitempty"`
	Message    string  `json:"message"`
	Path       string  `json:"path"`
}

// GetCodeownersErrors lists any syntax errors that are detected in the CODEOWNERS file.
//
// GitHub API docs: https://docs.github.com/en/rest/repos/repos#list-codeowners-errors
func (s *RepositoriesService) GetCodeownersErrors(ctx context.Context, owner, repo string) (*CodeownersErrors, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/codeowners/errors", owner, repo)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	codeownersErrors := &CodeownersErrors{}
	resp, err := s.client.Do(ctx, req, codeownersErrors)
	if err != nil {
		return nil, resp, err
	}

	return codeownersErrors, resp, nil
}
