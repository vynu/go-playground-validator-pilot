// Package models contains GitLab webhook payload models with comprehensive validation rules.
// This module defines all GitLab-specific data structures for webhook validation.
package models

import (
	"time"
)

// GitLabPayload represents the top-level GitLab webhook payload structure.
type GitLabPayload struct {
	ObjectKind        string              `json:"object_kind" validate:"required,oneof=push tag_push merge_request"`
	EventType         string              `json:"event_type" validate:"omitempty"`
	Before            string              `json:"before" validate:"omitempty,len=40,hexadecimal"`
	After             string              `json:"after" validate:"omitempty,len=40,hexadecimal"`
	Ref               string              `json:"ref" validate:"omitempty"`
	CheckoutSHA       string              `json:"checkout_sha" validate:"omitempty,len=40,hexadecimal"`
	Message           *string             `json:"message,omitempty"`
	UserID            int64               `json:"user_id" validate:"omitempty,gt=0"`
	UserName          string              `json:"user_name" validate:"omitempty,min=1"`
	UserUsername      string              `json:"user_username" validate:"omitempty,gitlab_username"`
	UserEmail         string              `json:"user_email" validate:"omitempty,email"`
	UserAvatar        string              `json:"user_avatar" validate:"omitempty,url"`
	ProjectID         int64               `json:"project_id" validate:"omitempty,gt=0"`
	Project           *GitLabProject      `json:"project,omitempty" validate:"omitempty"`
	Commits           []GitLabCommit      `json:"commits" validate:"omitempty,dive"`
	TotalCommitsCount int                 `json:"total_commits_count" validate:"omitempty,gte=0"`
	Repository        *GitLabRepository   `json:"repository,omitempty" validate:"omitempty"`
	ObjectAttributes  *GitLabMergeRequest `json:"object_attributes,omitempty" validate:"omitempty"`
	MergeRequest      *GitLabMergeRequest `json:"merge_request,omitempty" validate:"omitempty"`
	Assignees         []GitLabUser        `json:"assignees" validate:"omitempty,dive"`
	Assignee          *GitLabUser         `json:"assignee,omitempty" validate:"omitempty"`
	Reviewers         []GitLabUser        `json:"reviewers" validate:"omitempty,dive"`
	Labels            []GitLabLabel       `json:"labels" validate:"omitempty,dive"`
	Changes           *GitLabChanges      `json:"changes,omitempty" validate:"omitempty"`
}

// GitLabProject represents a GitLab project structure.
type GitLabProject struct {
	ID                int64   `json:"id" validate:"required,gt=0"`
	Name              string  `json:"name" validate:"required,min=1,max=100"`
	Description       *string `json:"description,omitempty" validate:"omitempty,max=2000"`
	WebURL            string  `json:"web_url" validate:"required,url"`
	AvatarURL         *string `json:"avatar_url,omitempty" validate:"omitempty,url"`
	GitSSHURL         string  `json:"git_ssh_url" validate:"required,min=1"`
	GitHTTPURL        string  `json:"git_http_url" validate:"required,url"`
	Namespace         string  `json:"namespace" validate:"required,min=1"`
	VisibilityLevel   int     `json:"visibility_level" validate:"gte=0,lte=20"`
	PathWithNamespace string  `json:"path_with_namespace" validate:"required,min=1"`
	DefaultBranch     string  `json:"default_branch" validate:"required,min=1"`
	Homepage          string  `json:"homepage" validate:"omitempty,url"`
	URL               string  `json:"url" validate:"required,min=1"`
	SSHURL            string  `json:"ssh_url" validate:"required,min=1"`
	HTTPURL           string  `json:"http_url" validate:"required,url"`
}

// GitLabRepository represents a GitLab repository structure.
type GitLabRepository struct {
	Name            string `json:"name" validate:"required,min=1"`
	URL             string `json:"url" validate:"required,min=1"`
	Description     string `json:"description" validate:"omitempty"`
	Homepage        string `json:"homepage" validate:"omitempty,url"`
	GitHTTPURL      string `json:"git_http_url" validate:"omitempty,url"`
	GitSSHURL       string `json:"git_ssh_url" validate:"omitempty,min=1"`
	VisibilityLevel int    `json:"visibility_level" validate:"gte=0,lte=20"`
}

// GitLabUser represents a GitLab user structure.
type GitLabUser struct {
	ID        int64  `json:"id" validate:"required,gt=0"`
	Name      string `json:"name" validate:"required,min=1,max=255"`
	Username  string `json:"username" validate:"required,gitlab_username"`
	Email     string `json:"email" validate:"omitempty,email"`
	AvatarURL string `json:"avatar_url" validate:"omitempty,url"`
	State     string `json:"state" validate:"omitempty,oneof=active blocked"`
}

// GitLabCommit represents a GitLab commit structure.
type GitLabCommit struct {
	ID        string           `json:"id" validate:"required,len=40,hexadecimal"`
	Message   string           `json:"message" validate:"required,min=1"`
	Title     string           `json:"title" validate:"omitempty"`
	Timestamp time.Time        `json:"timestamp" validate:"required"`
	URL       string           `json:"url" validate:"required,url"`
	Author    GitLabCommitUser `json:"author" validate:"required"`
	Added     []string         `json:"added" validate:"omitempty"`
	Modified  []string         `json:"modified" validate:"omitempty"`
	Removed   []string         `json:"removed" validate:"omitempty"`
}

// GitLabCommitUser represents a commit author/committer.
type GitLabCommitUser struct {
	Name  string `json:"name" validate:"required,min=1,max=255"`
	Email string `json:"email" validate:"required,email"`
}

// GitLabMergeRequest represents a GitLab merge request structure.
type GitLabMergeRequest struct {
	ID                        int64                    `json:"id" validate:"required,gt=0"`
	IID                       int                      `json:"iid" validate:"required,gt=0"`
	Title                     string                   `json:"title" validate:"required,min=1,max=255"`
	Description               *string                  `json:"description,omitempty" validate:"omitempty,max=1000000"`
	State                     string                   `json:"state" validate:"required,oneof=opened closed locked merged"`
	CreatedAt                 time.Time                `json:"created_at" validate:"required"`
	UpdatedAt                 time.Time                `json:"updated_at" validate:"required,gtefield=CreatedAt"`
	MergeStatus               string                   `json:"merge_status" validate:"omitempty,oneof=unchecked can_be_merged cannot_be_merged"`
	TargetBranch              string                   `json:"target_branch" validate:"required,min=1"`
	SourceBranch              string                   `json:"source_branch" validate:"required,min=1"`
	SourceProjectID           int64                    `json:"source_project_id" validate:"required,gt=0"`
	TargetProjectID           int64                    `json:"target_project_id" validate:"required,gt=0"`
	AuthorID                  int64                    `json:"author_id" validate:"required,gt=0"`
	AssigneeID                *int64                   `json:"assignee_id,omitempty" validate:"omitempty,gt=0"`
	URL                       string                   `json:"url" validate:"required,url"`
	Source                    GitLabMergeRequestSource `json:"source" validate:"required"`
	Target                    GitLabMergeRequestTarget `json:"target" validate:"required"`
	LastCommit                GitLabLastCommit         `json:"last_commit" validate:"required"`
	WorkInProgress            bool                     `json:"work_in_progress"`
	Assignee                  *GitLabUser              `json:"assignee,omitempty" validate:"omitempty"`
	MilestoneID               *int64                   `json:"milestone_id,omitempty" validate:"omitempty,gt=0"`
	MergeCommitSHA            *string                  `json:"merge_commit_sha,omitempty" validate:"omitempty,len=40,hexadecimal"`
	MergeError                *string                  `json:"merge_error,omitempty"`
	MergeParams               *GitLabMergeParams       `json:"merge_params,omitempty" validate:"omitempty"`
	MergeWhenPipelineSucceeds bool                     `json:"merge_when_pipeline_succeeds"`
	MergeUserID               *int64                   `json:"merge_user_id,omitempty" validate:"omitempty,gt=0"`
	DeleteSourceBranch        bool                     `json:"delete_source_branch"`
	TimeEstimate              int                      `json:"time_estimate" validate:"gte=0"`
	TotalTimeSpent            int                      `json:"total_time_spent" validate:"gte=0"`
	Squash                    bool                     `json:"squash"`
	Upvotes                   int                      `json:"upvotes" validate:"gte=0"`
	Downvotes                 int                      `json:"downvotes" validate:"gte=0"`
}

// GitLabMergeRequestSource represents merge request source information.
type GitLabMergeRequestSource struct {
	Name              string  `json:"name" validate:"required,min=1"`
	Description       string  `json:"description" validate:"omitempty"`
	WebURL            string  `json:"web_url" validate:"required,url"`
	AvatarURL         *string `json:"avatar_url,omitempty" validate:"omitempty,url"`
	GitSSHURL         string  `json:"git_ssh_url" validate:"required,min=1"`
	GitHTTPURL        string  `json:"git_http_url" validate:"required,url"`
	Namespace         string  `json:"namespace" validate:"required,min=1"`
	VisibilityLevel   int     `json:"visibility_level" validate:"gte=0,lte=20"`
	PathWithNamespace string  `json:"path_with_namespace" validate:"required,min=1"`
	DefaultBranch     string  `json:"default_branch" validate:"required,min=1"`
	Homepage          string  `json:"homepage" validate:"omitempty,url"`
	URL               string  `json:"url" validate:"required,min=1"`
	SSHURL            string  `json:"ssh_url" validate:"required,min=1"`
	HTTPURL           string  `json:"http_url" validate:"required,url"`
}

// GitLabMergeRequestTarget represents merge request target information.
type GitLabMergeRequestTarget struct {
	Name              string  `json:"name" validate:"required,min=1"`
	Description       string  `json:"description" validate:"omitempty"`
	WebURL            string  `json:"web_url" validate:"required,url"`
	AvatarURL         *string `json:"avatar_url,omitempty" validate:"omitempty,url"`
	GitSSHURL         string  `json:"git_ssh_url" validate:"required,min=1"`
	GitHTTPURL        string  `json:"git_http_url" validate:"required,url"`
	Namespace         string  `json:"namespace" validate:"required,min=1"`
	VisibilityLevel   int     `json:"visibility_level" validate:"gte=0,lte=20"`
	PathWithNamespace string  `json:"path_with_namespace" validate:"required,min=1"`
	DefaultBranch     string  `json:"default_branch" validate:"required,min=1"`
	Homepage          string  `json:"homepage" validate:"omitempty,url"`
	URL               string  `json:"url" validate:"required,min=1"`
	SSHURL            string  `json:"ssh_url" validate:"required,min=1"`
	HTTPURL           string  `json:"http_url" validate:"required,url"`
}

// GitLabLastCommit represents the last commit in a merge request.
type GitLabLastCommit struct {
	ID        string           `json:"id" validate:"required,len=40,hexadecimal"`
	Message   string           `json:"message" validate:"required,min=1"`
	Timestamp time.Time        `json:"timestamp" validate:"required"`
	URL       string           `json:"url" validate:"required,url"`
	Author    GitLabCommitUser `json:"author" validate:"required"`
}

// GitLabMergeParams represents merge parameters.
type GitLabMergeParams struct {
	ForceRemoveSourceBranch bool `json:"force_remove_source_branch"`
}

// GitLabLabel represents a GitLab label.
type GitLabLabel struct {
	ID          int64     `json:"id" validate:"required,gt=0"`
	Title       string    `json:"title" validate:"required,min=1,max=255"`
	Color       string    `json:"color" validate:"required,hexcolor"`
	ProjectID   *int64    `json:"project_id,omitempty" validate:"omitempty,gt=0"`
	CreatedAt   time.Time `json:"created_at" validate:"required"`
	UpdatedAt   time.Time `json:"updated_at" validate:"required,gtefield=CreatedAt"`
	Template    bool      `json:"template"`
	Description string    `json:"description" validate:"omitempty,max=500"`
	Type        string    `json:"type" validate:"omitempty,oneof=ProjectLabel GroupLabel"`
	GroupID     *int64    `json:"group_id,omitempty" validate:"omitempty,gt=0"`
}

// GitLabChanges represents changes in a webhook payload.
type GitLabChanges struct {
	Title        *GitLabChangeField          `json:"title,omitempty" validate:"omitempty"`
	Description  *GitLabChangeField          `json:"description,omitempty" validate:"omitempty"`
	TargetBranch *GitLabChangeField          `json:"target_branch,omitempty" validate:"omitempty"`
	State        *GitLabChangeField          `json:"state,omitempty" validate:"omitempty"`
	MergeStatus  *GitLabChangeField          `json:"merge_status,omitempty" validate:"omitempty"`
	UpdatedAt    *GitLabTimestampChangeField `json:"updated_at,omitempty" validate:"omitempty"`
	Labels       *GitLabLabelsChangeField    `json:"labels,omitempty" validate:"omitempty"`
	Assignees    *GitLabAssigneesChangeField `json:"assignees,omitempty" validate:"omitempty"`
}

// GitLabChangeField represents a changed field with previous and current values.
type GitLabChangeField struct {
	Previous *string `json:"previous,omitempty"`
	Current  *string `json:"current,omitempty"`
}

// GitLabTimestampChangeField represents a changed timestamp field.
type GitLabTimestampChangeField struct {
	Previous *time.Time `json:"previous,omitempty"`
	Current  *time.Time `json:"current,omitempty"`
}

// GitLabLabelsChangeField represents changes to labels.
type GitLabLabelsChangeField struct {
	Previous []GitLabLabel `json:"previous,omitempty" validate:"omitempty,dive"`
	Current  []GitLabLabel `json:"current,omitempty" validate:"omitempty,dive"`
}

// GitLabAssigneesChangeField represents changes to assignees.
type GitLabAssigneesChangeField struct {
	Previous []GitLabUser `json:"previous,omitempty" validate:"omitempty,dive"`
	Current  []GitLabUser `json:"current,omitempty" validate:"omitempty,dive"`
}
