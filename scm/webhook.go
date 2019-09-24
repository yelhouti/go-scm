// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scm

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	// ErrSignatureInvalid is returned when the webhook
	// signature is invalid or cannot be calculated.
	ErrSignatureInvalid = errors.New("Invalid webhook signature")
)

type (
	// Webhook defines a webhook for repository events.
	Webhook interface {
		Repository() Repository
	}

	// WebhookUnmarshaler wraps Webhook and assigns a type for unmarshalling.
	// Use this if you need to deserialize Webhooks of uknown concrete type.
	WebhookUnmarshaler struct {
		Type    string
		Webhook Webhook
	}

	// Label on a PR
	Label struct {
		URL         string
		Name        string
		Description string
		Color       string
	}

	// PushCommit represents general info about a commit.
	PushCommit struct {
		ID       string
		Message  string
		Added    []string
		Removed  []string
		Modified []string
	}

	// PushHook represents a push hook, eg push events.
	PushHook struct {
		Ref     string
		BaseRef string
		Repo    Repository
		Before  string
		After   string
		Created bool
		Deleted bool
		Forced  bool
		Compare string
		Commits []PushCommit
		Commit  Commit
		Sender  User
		GUID    string
	}

	// BranchHook represents a branch or tag event,
	// eg create and delete github event types.
	BranchHook struct {
		Ref    Reference
		Repo   Repository
		Action Action
		Sender User
	}

	// TagHook represents a tag event, eg create and delete
	// github event types.
	TagHook struct {
		Ref    Reference
		Repo   Repository
		Action Action
		Sender User
	}

	// IssueHook represents an issue event, eg issues.
	IssueHook struct {
		Action Action
		Repo   Repository
		Issue  Issue
		Sender User
	}

	// IssueCommentHook represents an issue comment event,
	// eg issue_comment.
	IssueCommentHook struct {
		Action  Action
		Repo    Repository
		Issue   Issue
		Comment Comment
		Sender  User
	}

	PullRequestHookBranchFrom struct {
		From string
	}

	PullRequestHookBranch struct {
		Ref  PullRequestHookBranchFrom
		Sha  PullRequestHookBranchFrom
		Repo Repository
	}

	PullRequestHookChanges struct {
		Base PullRequestHookBranch
	}

	// PullRequestHook represents an pull request event,
	// eg pull_request.
	PullRequestHook struct {
		Action      Action
		Repo        Repository
		Label       Label
		PullRequest PullRequest
		Sender      User
		Changes     PullRequestHookChanges
		GUID        string
	}

	// PullRequestCommentHook represents an pull request
	// comment event, eg pull_request_comment.
	PullRequestCommentHook struct {
		Action      Action
		Repo        Repository
		PullRequest PullRequest
		Comment     Comment
		Sender      User
	}

	// ReviewCommentHook represents a pull request review
	// comment, eg pull_request_review_comment.
	ReviewCommentHook struct {
		Action      Action
		Repo        Repository
		PullRequest PullRequest
		Review      Review
	}

	// DeployHook represents a deployment event. This is
	// currently a GitHub-specific event type.
	DeployHook struct {
		Data      interface{}
		Desc      string
		Ref       Reference
		Repo      Repository
		Sender    User
		Target    string
		TargetURL string
		Task      string
	}

	// SecretFunc provides the Webhook parser with the
	// secret key used to validate webhook authenticity.
	SecretFunc func(webhook Webhook) (string, error)

	// WebhookService provides abstract functions for
	// parsing and validating webhooks requests.
	WebhookService interface {
		// Parse returns the parsed the repository webhook payload.
		Parse(req *http.Request, fn SecretFunc) (Webhook, error)
	}
)

// Repository() defines the repository webhook and provides
// a convenient way to get the associated repository without
// having to cast the type.

func (h *PushHook) Repository() Repository               { return h.Repo }
func (h *BranchHook) Repository() Repository             { return h.Repo }
func (h *DeployHook) Repository() Repository             { return h.Repo }
func (h *TagHook) Repository() Repository                { return h.Repo }
func (h *IssueHook) Repository() Repository              { return h.Repo }
func (h *IssueCommentHook) Repository() Repository       { return h.Repo }
func (h *PullRequestHook) Repository() Repository        { return h.Repo }
func (h *PullRequestCommentHook) Repository() Repository { return h.Repo }
func (h *ReviewCommentHook) Repository() Repository      { return h.Repo }

// MarshalJSON implements custom JSON marshaling logic.
func (h *PushHook) MarshalJSON() ([]byte, error) {
	hook := make(map[string]interface{})
	hook["type"] = "pushHook"

	hook["ref"] = h.Ref
	hook["baseRef"] = h.BaseRef
	hook["repo"] = h.Repo
	hook["before"] = h.Before
	hook["after"] = h.After
	hook["created"] = h.Created
	hook["deleted"] = h.Deleted
	hook["forced"] = h.Forced
	hook["compare"] = h.Compare
	hook["commits"] = h.Commits
	hook["commit"] = h.Commit
	hook["sender"] = h.Sender
	hook["guid"] = h.GUID

	return json.Marshal(hook)
}

// MarshalJSON implements custom JSON marshaling logic.
func (h *BranchHook) MarshalJSON() ([]byte, error) {
	hook := make(map[string]interface{})
	hook["type"] = "branchHook"

	hook["ref"] = h.Ref
	hook["repo"] = h.Repo
	hook["action"] = h.Action
	hook["sender"] = h.Sender

	return json.Marshal(hook)
}

// MarshalJSON implements custom JSON marshaling logic.
func (h *DeployHook) MarshalJSON() ([]byte, error) {
	hook := make(map[string]interface{})
	hook["type"] = "deployHook"

	hook["data"] = h.Data
	hook["desc"] = h.Desc
	hook["ref"] = h.Ref
	hook["repo"] = h.Repo
	hook["sender"] = h.Sender
	hook["target"] = h.Target
	hook["targetUrl"] = h.TargetURL
	hook["task"] = h.Task

	return json.Marshal(hook)
}

// MarshalJSON implements custom JSON marshaling logic.
func (h *TagHook) MarshalJSON() ([]byte, error) {
	hook := make(map[string]interface{})
	hook["type"] = "tagHook"

	hook["ref"] = h.Ref
	hook["repo"] = h.Repo
	hook["action"] = h.Action
	hook["sender"] = h.Sender

	return json.Marshal(hook)
}

// MarshalJSON implements custom JSON marshaling logic.
func (h *IssueHook) MarshalJSON() ([]byte, error) {
	hook := make(map[string]interface{})
	hook["type"] = "issueHook"

	hook["action"] = h.Action
	hook["repo"] = h.Repo
	hook["issue"] = h.Issue
	hook["sender"] = h.Sender

	return json.Marshal(hook)
}

// MarshalJSON implements custom JSON marshaling logic.
func (h *IssueCommentHook) MarshalJSON() ([]byte, error) {
	hook := make(map[string]interface{})
	hook["type"] = "issueCommentHook"

	hook["action"] = h.Action
	hook["repo"] = h.Repo
	hook["issue"] = h.Issue
	hook["comment"] = h.Comment
	hook["sender"] = h.Sender

	return json.Marshal(hook)
}

// MarshalJSON implements custom JSON marshaling logic.
func (h *PullRequestHook) MarshalJSON() ([]byte, error) {
	hook := make(map[string]interface{})
	hook["type"] = "pullRequestHook"

	hook["action"] = h.Action
	hook["repo"] = h.Repo
	hook["label"] = h.Label
	hook["pullRequest"] = h.PullRequest
	hook["sender"] = h.Sender
	hook["changes"] = h.Changes
	hook["guid"] = h.GUID

	return json.Marshal(hook)
}

// MarshalJSON implements custom JSON marshaling logic.
func (h *PullRequestCommentHook) MarshalJSON() ([]byte, error) {
	hook := make(map[string]interface{})
	hook["type"] = "pullRequestCommentHook"

	hook["action"] = h.Action
	hook["repo"] = h.Repo
	hook["pullRequest"] = h.PullRequest
	hook["comment"] = h.Comment
	hook["sender"] = h.Sender

	return json.Marshal(hook)
}

// MarshalJSON implements custom JSON marshaling logic.
func (h *ReviewCommentHook) MarshalJSON() ([]byte, error) {
	hook := make(map[string]interface{})
	hook["type"] = "reviewCommentHook"

	hook["action"] = h.Action
	hook["repo"] = h.Repo
	hook["pullRequest"] = h.PullRequest
	hook["review"] = h.Review

	return json.Marshal(hook)
}

// UnmarshalJSON supports deserialization of GitEventSpec.ParsedWebhook into a concrete implementation of scm.Webhook
func (wu *WebhookUnmarshaler) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}

	var rawMessage *json.RawMessage
	var webhookMap map[string]string
	err = json.Unmarshal(*rawMessage, &webhookMap)
	if err != nil {
		return err
	}

	if webhookMap["type"] == "pushHook" {

		var h *PushHook
		err = json.Unmarshal(*rawMessage, h)
		if err != nil {
			return err
		}
		wu.Webhook = h

	} else if webhookMap["type"] == "branchHook" {

		var h *BranchHook
		err = json.Unmarshal(*rawMessage, h)
		if err != nil {
			return err
		}
		wu.Webhook = h

	} else if webhookMap["type"] == "deployHook" {

		var h *DeployHook
		err = json.Unmarshal(*rawMessage, h)
		if err != nil {
			return err
		}
		wu.Webhook = h

	} else if webhookMap["type"] == "tagHook" {

		var h *TagHook
		err = json.Unmarshal(*rawMessage, h)
		if err != nil {
			return err
		}
		wu.Webhook = h

	} else if webhookMap["type"] == "issueHook" {

		var h *IssueHook
		err = json.Unmarshal(*rawMessage, h)
		if err != nil {
			return err
		}
		wu.Webhook = h

	} else if webhookMap["type"] == "issueCommentHook" {

		var h *IssueHook
		err = json.Unmarshal(*rawMessage, h)
		if err != nil {
			return err
		}
		wu.Webhook = h

	} else if webhookMap["type"] == "pullRequestHook" {

		var h *IssueHook
		err = json.Unmarshal(*rawMessage, h)
		if err != nil {
			return err
		}
		wu.Webhook = h

	} else if webhookMap["type"] == "pullRequestCommentHook" {

		var h *IssueHook
		err = json.Unmarshal(*rawMessage, h)
		if err != nil {
			return err
		}
		wu.Webhook = h

	} else if webhookMap["type"] == "reviewCommentHook" {

		var h *IssueHook
		err = json.Unmarshal(*rawMessage, h)
		if err != nil {
			return err
		}
		wu.Webhook = h

	}

	return nil
}
