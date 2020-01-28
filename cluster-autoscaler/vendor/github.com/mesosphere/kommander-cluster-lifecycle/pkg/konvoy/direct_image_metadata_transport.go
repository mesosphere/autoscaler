package konvoy

/*
 * Copyright 2019 Mesosphere, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/auth/challenge"
	"github.com/docker/distribution/registry/client/transport"
)

// NOTE: Some of this code has been taken from https://github.com/docker/distribution/blob/master/registry/client/auth
// packages to be able to talk to docker registry api.

type userpass struct {
	username string
	password string
}

type credentials struct {
	creds         map[string]userpass
	refreshTokens map[string]string
}

type remoteAuthChallenger struct {
	remoteURL url.URL
	sync.Mutex
	cm challenge.Manager
	cs auth.CredentialStore
}

func (r *remoteAuthChallenger) challengeManager() challenge.Manager {
	return r.cm
}

// tryEstablishChallenges will attempt to get a challenge type for the upstream if none currently exist
func (r *remoteAuthChallenger) tryEstablishChallenges() error {
	r.Lock()
	defer r.Unlock()

	remoteURL := r.remoteURL
	remoteURL.Path = "/v2/"
	challenges, err := r.cm.GetChallenges(remoteURL)
	if err != nil {
		return err
	}

	if len(challenges) > 0 {
		return nil
	}

	// establish challenge type with upstream
	if err := ping(r.cm, remoteURL.String()); err != nil {
		return err
	}

	return nil
}

func NewTokenAuthTransport(opts ImageMetadataOptions) (http.RoundTripper, error) {
	var defaultTransport http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: opts.DockerRegistryAuthSkipVerify},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	cs, err := configureAuth(opts.DockerRegistryUsername, opts.DockerRegistryPassword, opts.DockerRegistryURL)
	if err != nil {
		return nil, err
	}

	remoteURL, err := url.Parse(opts.DockerRegistryURL)
	if err != nil {
		return nil, err
	}

	challengeAuth := &remoteAuthChallenger{
		remoteURL: *remoteURL,
		cm:        challenge.NewSimpleManager(),
		cs:        cs,
	}

	if err := challengeAuth.tryEstablishChallenges(); err != nil {
		return nil, err
	}

	authorizer := auth.NewAuthorizer(challengeAuth.challengeManager(), auth.NewTokenHandler(defaultTransport, cs, opts.DockerRegistryRepository, "pull"), auth.NewBasicHandler(cs))

	return transport.NewTransport(defaultTransport, authorizer), nil
}

// configureAuth stores credentials for challenge responses
func configureAuth(username, password, remoteURL string) (auth.CredentialStore, error) {
	creds := map[string]userpass{}

	authURLs, err := getAuthURLs(remoteURL)
	if err != nil {
		return nil, err
	}

	for _, url := range authURLs {
		creds[url] = userpass{
			username: username,
			password: password,
		}
	}

	return credentials{creds: creds}, nil
}

func (c credentials) Basic(u *url.URL) (string, string) {
	up := c.creds[u.String()]

	return up.username, up.password
}

func (c credentials) RefreshToken(u *url.URL, service string) string {
	return c.refreshTokens[service]
}

func (c credentials) SetRefreshToken(u *url.URL, service, token string) {
	if c.refreshTokens != nil {
		c.refreshTokens[service] = token
	}
}

func getAuthURLs(remoteURL string) ([]string, error) {
	authURLs := []string{}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	resp, err := http.Get(remoteURL + "/v2/")
	if err != nil {
		return nil, fmt.Errorf("error getting the autho urls for the docker registry, %v", err)
	}
	defer resp.Body.Close()

	for _, c := range challenge.ResponseChallenges(resp) {
		if strings.EqualFold(c.Scheme, "bearer") {
			authURLs = append(authURLs, c.Parameters["realm"])
		}
	}

	return authURLs, nil
}

func ping(manager challenge.Manager, endpoint string) error {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Get(endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return manager.AddResponse(resp)
}
