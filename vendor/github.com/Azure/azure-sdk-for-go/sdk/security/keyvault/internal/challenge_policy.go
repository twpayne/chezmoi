//go:build go1.18
// +build go1.18

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package internal

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/internal/errorinfo"
)

const challengeMatchError = `challenge resource "%s" doesn't match the requested domain. Set DisableChallengeResourceVerification to true in your client options to disable. See https://aka.ms/azsdk/blog/vault-uri for more information`

type KeyVaultChallengePolicyOptions struct {
	// DisableChallengeResourceVerification controls whether the policy requires the
	// authentication challenge resource to match the Key Vault or Managed HSM domain
	DisableChallengeResourceVerification bool
}

type keyVaultAuthorizer struct {
	// tro is the policy's authentication parameters. These are discovered from an authentication challenge
	// elicited ahead of the first client request.
	tro policy.TokenRequestOptions
	// TODO: move into tro once it has a tenant field (https://github.com/Azure/azure-sdk-for-go/issues/19841)
	tenantID                string
	verifyChallengeResource bool
}

type reqBody struct {
	body        io.ReadSeekCloser
	contentType string
}

func NewKeyVaultChallengePolicy(cred azcore.TokenCredential, opts *KeyVaultChallengePolicyOptions) policy.Policy {
	if opts == nil {
		opts = &KeyVaultChallengePolicyOptions{}
	}
	kv := keyVaultAuthorizer{
		verifyChallengeResource: !opts.DisableChallengeResourceVerification,
	}
	return runtime.NewBearerTokenPolicy(cred, nil, &policy.BearerTokenOptions{
		AuthorizationHandler: policy.AuthorizationHandler{
			OnRequest:   kv.authorize,
			OnChallenge: kv.authorizeOnChallenge,
		},
	})
}

func (k *keyVaultAuthorizer) authorize(req *policy.Request, authNZ func(policy.TokenRequestOptions) error) error {
	if len(k.tro.Scopes) == 0 || k.tenantID == "" {
		if body := req.Body(); body != nil {
			// We don't know the scope or tenant ID because we haven't seen a challenge yet. We elicit one now by sending
			// the request without authorization, first removing its body, if any. authorizeOnChallenge will reattach the
			// body, authorize the request, and send it again.
			rb := reqBody{body, req.Raw().Header.Get("content-type")}
			req.SetOperationValue(rb)
			if err := req.SetBody(nil, ""); err != nil {
				return err
			}
		}
		// returning nil indicates the bearer token policy should send the request
		return nil
	}
	// else we know the auth parameters and can authorize the request as normal
	return authNZ(k.tro)
}

func (k *keyVaultAuthorizer) authorizeOnChallenge(req *policy.Request, res *http.Response, authNZ func(policy.TokenRequestOptions) error) error {
	// parse the challenge
	if err := k.updateTokenRequestOptions(res, req.Raw()); err != nil {
		return err
	}
	// reattach the request's original body, if it was removed by authorize(). If a bug prevents recovering
	// the body, this policy will send the request without it and get a 400 response from Key Vault.
	var rb reqBody
	if req.OperationValue(&rb) {
		if err := req.SetBody(rb.body, rb.contentType); err != nil {
			return err
		}
	}
	// authenticate with the parameters supplied by Key Vault, authorize the request, send it again
	return authNZ(k.tro)
}

// parses Tenant ID from auth challenge
// https://login.microsoftonline.com/00000000-0000-0000-0000-000000000000
func parseTenant(url string) string {
	if url == "" {
		return ""
	}
	parts := strings.Split(url, "/")
	tenant := parts[3]
	tenant = strings.ReplaceAll(tenant, ",", "")
	return tenant
}

type challengePolicyError struct {
	err error
}

func (c *challengePolicyError) Error() string {
	return c.err.Error()
}

func (*challengePolicyError) NonRetriable() {
	// marker method
}

func (c *challengePolicyError) Unwrap() error {
	return c.err
}

var _ errorinfo.NonRetriable = (*challengePolicyError)(nil)

// updateTokenRequestOptions parses authentication parameters from Key Vault's challenge
func (k *keyVaultAuthorizer) updateTokenRequestOptions(resp *http.Response, req *http.Request) error {
	authHeader := resp.Header.Get("WWW-Authenticate")
	if authHeader == "" {
		return &challengePolicyError{err: errors.New("response has no WWW-Authenticate header for challenge authentication")}
	}

	// Strip down to auth and resource
	// Format is "Bearer authorization=\"<site>\" resource=\"<site>\"" OR
	// "Bearer authorization=\"<site>\" scope=\"<site>\" resource=\"<resource>\""
	authHeader = strings.ReplaceAll(authHeader, "Bearer ", "")

	parts := strings.Split(authHeader, " ")

	vals := map[string]string{}
	for _, part := range parts {
		subParts := strings.Split(part, "=")
		if len(subParts) == 2 {
			stripped := strings.ReplaceAll(subParts[1], "\"", "")
			stripped = strings.TrimSuffix(stripped, ",")
			vals[subParts[0]] = stripped
		}
	}

	k.tenantID = parseTenant(vals["authorization"])
	scope := ""
	if v, ok := vals["scope"]; ok {
		scope = v
	} else if v, ok := vals["resource"]; ok {
		scope = v
	}
	if scope == "" {
		return &challengePolicyError{err: errors.New("could not find a valid resource in the WWW-Authenticate header")}
	}
	if k.verifyChallengeResource {
		// the challenge resource's host must match the requested vault's host
		parsed, err := url.Parse(scope)
		if err != nil {
			return &challengePolicyError{err: fmt.Errorf(`invalid challenge resource "%s": %v`, scope, err)}
		}
		if !strings.HasSuffix(req.URL.Host, "."+parsed.Host) {
			return &challengePolicyError{err: fmt.Errorf(challengeMatchError, scope)}
		}
	}
	if !strings.HasSuffix(scope, "/.default") {
		scope += "/.default"
	}
	k.tro.Scopes = []string{scope}
	return nil
}
