// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gitlab implements a GitLab client.
package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/shurcooL/graphql"
)

func NewWebHookService() scm.WebhookService {
	return &webhookService{nil}
}

// New returns a new GitLab API client.
func New(uri string) (*scm.Client, error) {
	base, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(base.Path, "/") {
		base.Path = base.Path + "/"
	}
	client := &wrapper{new(scm.Client)}
	client.BaseURL = base
	// initialize services
	client.Driver = scm.DriverGitlab
	client.Contents = &contentService{client}
	client.Git = &gitService{client}
	client.Issues = &issueService{client}
	client.Milestones = &milestoneService{client}
	client.Organizations = &organizationService{client}
	client.PullRequests = &pullService{client}
	client.Repositories = &repositoryService{client}
	client.Reviews = &reviewService{client}
	client.Users = &userService{client}
	client.Webhooks = &webhookService{client}

	graphqlEndpoint := scm.UrlJoin(uri, "/api/graphql")
	client.GraphQLURL, err = url.Parse(graphqlEndpoint)
	if err != nil {
		return nil, err
	}
	client.GraphQL = &dynamicGraphQLClient{client, graphqlEndpoint}
	return client.Client, nil
}

type dynamicGraphQLClient struct {
	wrapper         *wrapper
	graphqlEndpoint string
}

func (d *dynamicGraphQLClient) Query(ctx context.Context, q interface{}, vars map[string]interface{}) error {
	httpClient := d.wrapper.Client.Client
	if httpClient != nil {

		transport := httpClient.Transport
		if transport != nil {
			query := graphql.NewClient(
				d.graphqlEndpoint,
				&http.Client{
					Transport: transport,
				})
			return query.Query(ctx, q, vars)
		}
	}
	fmt.Println("WARNING: no http transport configured for GraphQL and Gitlab")
	return nil
}

// NewDefault returns a new GitLab API client using the
// default gitlab.com address.
func NewDefault() *scm.Client {
	client, _ := New("https://gitlab.com")
	return client
}

// wraper wraps the Client to provide high level helper functions
// for making http requests and unmarshaling the response.
type wrapper struct {
	*scm.Client
}

type gl_namespace struct {
	ID                          int    `json:"id"`
	Name                        string `json:"name"`
	Path                        string `json:"path"`
	Kind                        string `json:"kind"`
	FullPath                    string `json:"full_path"`
	ParentID                    int    `json:"parent_id"`
	AvatarURL                   string  `json:"avatar_url"`
	WebURL                      string  `json:"web_url"`
}

// findNamespaceByName will look up the namespace for the given name
func (c *wrapper) findNamespaceByName(ctx context.Context, name string) (*gl_namespace, error) {
	in := url.Values{}
	in.Set("search", name)
	path := fmt.Sprintf("api/v4/namespaces?%s", in.Encode())

	out := new(gl_namespace)
	_, err := c.do(ctx, "GET", path, nil, &out)

	return out, err
}

// `do_gl_namespace` is very same with `do()`, but it needs one further step to convert array to json for correct marshall
func (c *wrapper) do_gl_namespace(ctx context.Context, method, path string, in, out interface{}) (*scm.Response, error) {
	req := &scm.Request{
		Method: method,
		Path:   path,
	}
	// if we are posting or putting data, we need to
	// write it to the body of the request.
	if in != nil {
		buf := new(bytes.Buffer)
		json.NewEncoder(buf).Encode(in)
		if req.Header == nil {
			req.Header = map[string][]string{}
		}
		req.Header.Set("Content-Type", "application/json")
		req.Body = buf
	}
	// execute the http request
	res, err := c.Client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// parse the gitlab request id.
	res.ID = res.Header.Get("X-Request-Id")

	// parse the gitlab rate limit details.
	res.Rate.Limit, _ = strconv.Atoi(
		res.Header.Get("RateLimit-Limit"),
	)
	res.Rate.Remaining, _ = strconv.Atoi(
		res.Header.Get("RateLimit-Remaining"),
	)
	res.Rate.Reset, _ = strconv.ParseInt(
		res.Header.Get("RateLimit-Reset"), 10, 64,
	)

	// snapshot the request rate limit
	c.Client.SetRate(res.Rate)

	// if an error is encountered, unmarshal and return the
	// error response.
	if res.Status > 300 {
		err := new(Error)
		json.NewDecoder(res.Body).Decode(err)
		return res, err
	}

	if out == nil {
		return res, nil
	}

	// remove the bracket to convert array to json for correct marshall into gl_namespace
	buf := new(strings.Builder)
	_, err = io.Copy(buf, res.Body)
	str_buf := buf.String()
	new_str := buf.String()[1:len(str_buf)-1]

	// if a json response is expected, parse and return
	// the json response.
	return res, json.Unmarshal([]byte(new_str),&out)
}

// do wraps the Client.Do function by creating the Request and
// unmarshalling the response.
func (c *wrapper) do(ctx context.Context, method, path string, in, out interface{}) (*scm.Response, error) {
	req := &scm.Request{
		Method: method,
		Path:   path,
	}
	// if we are posting or putting data, we need to
	// write it to the body of the request.
	if in != nil {
		buf := new(bytes.Buffer)
		json.NewEncoder(buf).Encode(in)
		if req.Header == nil {
			req.Header = map[string][]string{}
		}
		req.Header.Set("Content-Type", "application/json")
		req.Body = buf
	}
	// execute the http request
	res, err := c.Client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// parse the gitlab request id.
	res.ID = res.Header.Get("X-Request-Id")

	// parse the gitlab rate limit details.
	res.Rate.Limit, _ = strconv.Atoi(
		res.Header.Get("RateLimit-Limit"),
	)
	res.Rate.Remaining, _ = strconv.Atoi(
		res.Header.Get("RateLimit-Remaining"),
	)
	res.Rate.Reset, _ = strconv.ParseInt(
		res.Header.Get("RateLimit-Reset"), 10, 64,
	)

	// snapshot the request rate limit
	c.Client.SetRate(res.Rate)

	// if an error is encountered, unmarshal and return the
	// error response.
	if res.Status > 300 {
		buf := new(strings.Builder)
		_, err = io.Copy(buf, res.Body)
		str_buf := buf.String()
		err_i := new(Error)
		json.Unmarshal([]byte(str_buf),&err_i)

		return res, err_i
	}

	if out == nil {
		return res, nil
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, res.Body)
	str_buf := buf.String()

	// if a json response is expected, parse and return
	// the json response.
	return res, json.Unmarshal([]byte(str_buf),&out)
}

// Error represents a GitLab error.
type Error struct {
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

type updateNoteOptions struct {
	Body string `json:"body"`
}

type labelEvent struct {
	ID           int        `json:"id"`
	Action       string     `json:"action"`
	CreatedAt    *time.Time `json:"created_at"`
	ResourceType string     `json:"resource_type"`
	ResourceID   int        `json:"resource_id"`
	User         user       `json:"user"`
	Label        label      `json:"label"`
}

func convertLabelEvents(src []*labelEvent) []*scm.ListedIssueEvent {
	var answer []*scm.ListedIssueEvent
	for _, from := range src {
		answer = append(answer, convertLabelEvent(from))
	}
	return answer
}

func convertLabelEvent(from *labelEvent) *scm.ListedIssueEvent {
	event := "labeled"
	if from.Action == "remove" {
		event = "unlabeled"
	}
	return &scm.ListedIssueEvent{
		Event:   event,
		Actor:   *convertUser(&from.User),
		Label:   *convertLabel(&from.Label),
		Created: *from.CreatedAt,
	}
}
