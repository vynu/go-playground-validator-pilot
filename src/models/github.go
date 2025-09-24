// Package models contains GitHub webhook payload models with comprehensive validation rules.
// This module defines all GitHub-specific data structures for webhook validation.
package models

import (
	"time"
)

// GitHubPayload represents the top-level GitHub webhook payload structure.
// This is the main entry point for validation and contains all required fields
// that GitHub sends in pull request webhook events.
type GitHubPayload struct {
	Action       string      `json:"action" validate:"required,oneof=opened closed reopened synchronize"`
	Number       int         `json:"number" validate:"required,gt=0"`
	PullRequest  PullRequest `json:"pull_request" validate:"required"`
	Repository   Repository  `json:"repository" validate:"required"`
	Sender       User        `json:"sender" validate:"required"`
	Organization *User       `json:"organization,omitempty" validate:"omitempty"`
	Installation *struct {
		ID     int    `json:"id" validate:"required,gt=0"`
		NodeID string `json:"node_id" validate:"required,min=1"`
	} `json:"installation,omitempty" validate:"omitempty"`
}

// PullRequest represents a comprehensive GitHub pull request structure.
// It includes all metadata, state information, and validation rules
// required for thorough data quality assessment.
type PullRequest struct {
	ID                  int64      `json:"id" validate:"required,gt=0"`
	NodeID              string     `json:"node_id" validate:"required,min=1"`
	Number              int        `json:"number" validate:"required,gt=0"`
	State               string     `json:"state" validate:"required,oneof=open closed merged"`
	Locked              bool       `json:"locked"`
	Title               string     `json:"title" validate:"required,min=1,max=256"`
	Body                *string    `json:"body" validate:"omitempty,max=65536"`
	CreatedAt           time.Time  `json:"created_at" validate:"required"`
	UpdatedAt           time.Time  `json:"updated_at" validate:"required,gtefield=CreatedAt"`
	ClosedAt            *time.Time `json:"closed_at,omitempty" validate:"omitempty"`
	MergedAt            *time.Time `json:"merged_at,omitempty" validate:"omitempty"`
	MergeCommitSHA      *string    `json:"merge_commit_sha,omitempty" validate:"omitempty,len=40,hexadecimal"`
	Assignee            *User      `json:"assignee,omitempty" validate:"omitempty"`
	Assignees           []User     `json:"assignees" validate:"omitempty,dive"`
	RequestedReviewers  []User     `json:"requested_reviewers" validate:"omitempty,dive"`
	RequestedTeams      []Team     `json:"requested_teams" validate:"omitempty,dive"`
	Labels              []Label    `json:"labels" validate:"omitempty,dive"`
	Milestone           *Milestone `json:"milestone,omitempty" validate:"omitempty"`
	Draft               bool       `json:"draft"`
	CommitsURL          string     `json:"commits_url" validate:"required,url"`
	ReviewCommentsURL   string     `json:"review_comments_url" validate:"required,url"`
	CommentsURL         string     `json:"comments_url" validate:"required,url"`
	StatusesURL         string     `json:"statuses_url" validate:"required,url"`
	Head                Reference  `json:"head" validate:"required"`
	Base                Reference  `json:"base" validate:"required"`
	User                User       `json:"user" validate:"required"`
	Mergeable           *bool      `json:"mergeable,omitempty"`
	Rebaseable          *bool      `json:"rebaseable,omitempty"`
	MergeableState      string     `json:"mergeable_state" validate:"omitempty,oneof=clean unstable dirty unknown blocked behind draft"`
	MergedBy            *User      `json:"merged_by,omitempty" validate:"omitempty"`
	Comments            int        `json:"comments" validate:"gte=0"`
	ReviewComments      int        `json:"review_comments" validate:"gte=0"`
	MaintainerCanModify bool       `json:"maintainer_can_modify"`
	Commits             int        `json:"commits" validate:"required,gt=0"`
	Additions           int        `json:"additions" validate:"gte=0"`
	Deletions           int        `json:"deletions" validate:"gte=0"`
	ChangedFiles        int        `json:"changed_files" validate:"gte=0"`
}

// Repository represents a GitHub repository with comprehensive validation.
// It includes ownership, metadata, settings, and usage statistics
// with appropriate validation constraints for each field.
type Repository struct {
	ID                  int64     `json:"id" validate:"required,gt=0"`
	NodeID              string    `json:"node_id" validate:"required,min=1"`
	Name                string    `json:"name" validate:"required,min=1,max=100"`
	FullName            string    `json:"full_name" validate:"required,min=1,max=200"`
	Private             bool      `json:"private"`
	Owner               User      `json:"owner" validate:"required"`
	HTMLURL             string    `json:"html_url" validate:"required,url"`
	Description         *string   `json:"description,omitempty" validate:"omitempty,max=350"`
	Fork                bool      `json:"fork"`
	URL                 string    `json:"url" validate:"required,url"`
	CreatedAt           time.Time `json:"created_at" validate:"required"`
	UpdatedAt           time.Time `json:"updated_at" validate:"required"`
	PushedAt            time.Time `json:"pushed_at" validate:"required"`
	GitURL              string    `json:"git_url" validate:"required,url"`
	SSHURL              string    `json:"ssh_url" validate:"required,min=1"`
	CloneURL            string    `json:"clone_url" validate:"required,url"`
	Homepage            *string   `json:"homepage,omitempty" validate:"omitempty,url"`
	Size                int       `json:"size" validate:"gte=0"`
	StargazersCount     int       `json:"stargazers_count" validate:"gte=0"`
	WatchersCount       int       `json:"watchers_count" validate:"gte=0"`
	Language            *string   `json:"language,omitempty" validate:"omitempty,min=1,max=50"`
	HasIssues           bool      `json:"has_issues"`
	HasProjects         bool      `json:"has_projects"`
	HasWiki             bool      `json:"has_wiki"`
	HasPages            bool      `json:"has_pages"`
	ForksCount          int       `json:"forks_count" validate:"gte=0"`
	OpenIssuesCount     int       `json:"open_issues_count" validate:"gte=0"`
	Forks               int       `json:"forks" validate:"gte=0"`
	OpenIssues          int       `json:"open_issues" validate:"gte=0"`
	Watchers            int       `json:"watchers" validate:"gte=0"`
	DefaultBranch       string    `json:"default_branch" validate:"required,min=1,max=255"`
	Topics              []string  `json:"topics" validate:"omitempty,dive,min=1,max=50"`
	Visibility          string    `json:"visibility" validate:"required,oneof=public private internal"`
	AllowSquashMerge    bool      `json:"allow_squash_merge"`
	AllowMergeCommit    bool      `json:"allow_merge_commit"`
	AllowRebaseMerge    bool      `json:"allow_rebase_merge"`
	DeleteBranchOnMerge bool      `json:"delete_branch_on_merge"`
}

// User represents a GitHub user, organization, or bot account.
// It includes all profile information and validation for GitHub-specific
// constraints like username format and URL structures.
type User struct {
	Login             string     `json:"login" validate:"required,github_username"`
	ID                int64      `json:"id" validate:"required,gt=0"`
	NodeID            string     `json:"node_id" validate:"required,min=1"`
	AvatarURL         string     `json:"avatar_url" validate:"required,url"`
	GravatarID        *string    `json:"gravatar_id,omitempty"`
	URL               string     `json:"url" validate:"required,url"`
	HTMLURL           string     `json:"html_url" validate:"required,url"`
	FollowersURL      string     `json:"followers_url" validate:"required,url"`
	FollowingURL      string     `json:"following_url" validate:"required,url"`
	GistsURL          string     `json:"gists_url" validate:"required,url"`
	StarredURL        string     `json:"starred_url" validate:"required,url"`
	SubscriptionsURL  string     `json:"subscriptions_url" validate:"required,url"`
	OrganizationsURL  string     `json:"organizations_url" validate:"required,url"`
	ReposURL          string     `json:"repos_url" validate:"required,url"`
	EventsURL         string     `json:"events_url" validate:"required,url"`
	ReceivedEventsURL string     `json:"received_events_url" validate:"required,url"`
	Type              string     `json:"type" validate:"required,oneof=User Organization Bot"`
	SiteAdmin         bool       `json:"site_admin"`
	Name              *string    `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Company           *string    `json:"company,omitempty" validate:"omitempty,max=255"`
	Blog              *string    `json:"blog,omitempty" validate:"omitempty,url"`
	Location          *string    `json:"location,omitempty" validate:"omitempty,max=255"`
	Email             *string    `json:"email,omitempty" validate:"omitempty,email"`
	Bio               *string    `json:"bio,omitempty" validate:"omitempty,max=160"`
	TwitterUsername   *string    `json:"twitter_username,omitempty" validate:"omitempty,max=15"`
	PublicRepos       *int       `json:"public_repos,omitempty" validate:"omitempty,gte=0"`
	PublicGists       *int       `json:"public_gists,omitempty" validate:"omitempty,gte=0"`
	Followers         *int       `json:"followers,omitempty" validate:"omitempty,gte=0"`
	Following         *int       `json:"following,omitempty" validate:"omitempty,gte=0"`
	CreatedAt         *time.Time `json:"created_at,omitempty" validate:"omitempty"`
	UpdatedAt         *time.Time `json:"updated_at,omitempty" validate:"omitempty"`
}

// Reference represents a Git reference (branch, tag, or commit).
// It includes the commit SHA, branch information, and associated
// repository details with full validation.
type Reference struct {
	Label string     `json:"label" validate:"required,min=1"`
	Ref   string     `json:"ref" validate:"required,min=1"`
	SHA   string     `json:"sha" validate:"required,len=40,hexadecimal"`
	User  User       `json:"user" validate:"required"`
	Repo  Repository `json:"repo" validate:"required"`
}

// Label represents a GitHub issue/PR label with color and metadata.
// It validates the hexadecimal color format and text constraints.
type Label struct {
	ID          int64   `json:"id" validate:"required,gt=0"`
	NodeID      string  `json:"node_id" validate:"required,min=1"`
	URL         string  `json:"url" validate:"required,url"`
	Name        string  `json:"name" validate:"required,min=1,max=50"`
	Color       string  `json:"color" validate:"required,hexcolor"`
	Default     bool    `json:"default"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=100"`
}

// Team represents a GitHub team
type Team struct {
	ID              int64   `json:"id" validate:"required,gt=0"`
	NodeID          string  `json:"node_id" validate:"required,min=1"`
	URL             string  `json:"url" validate:"required,url"`
	HTMLURL         string  `json:"html_url" validate:"required,url"`
	Name            string  `json:"name" validate:"required,min=1,max=255"`
	Slug            string  `json:"slug" validate:"required,min=1,max=255"`
	Description     *string `json:"description,omitempty" validate:"omitempty,max=255"`
	Privacy         string  `json:"privacy" validate:"required,oneof=closed secret"`
	Permission      string  `json:"permission" validate:"required,oneof=pull push admin maintain triage"`
	MembersURL      string  `json:"members_url" validate:"required,url"`
	RepositoriesURL string  `json:"repositories_url" validate:"required,url"`
	Parent          *Team   `json:"parent,omitempty" validate:"omitempty"`
}

// Milestone represents a GitHub milestone
type Milestone struct {
	URL          string     `json:"url" validate:"required,url"`
	HTMLURL      string     `json:"html_url" validate:"required,url"`
	LabelsURL    string     `json:"labels_url" validate:"required,url"`
	ID           int64      `json:"id" validate:"required,gt=0"`
	NodeID       string     `json:"node_id" validate:"required,min=1"`
	Number       int        `json:"number" validate:"required,gt=0"`
	State        string     `json:"state" validate:"required,oneof=open closed"`
	Title        string     `json:"title" validate:"required,min=1,max=255"`
	Description  *string    `json:"description,omitempty" validate:"omitempty,max=1000"`
	Creator      User       `json:"creator" validate:"required"`
	OpenIssues   int        `json:"open_issues" validate:"gte=0"`
	ClosedIssues int        `json:"closed_issues" validate:"gte=0"`
	CreatedAt    time.Time  `json:"created_at" validate:"required"`
	UpdatedAt    time.Time  `json:"updated_at" validate:"required,gtefield=CreatedAt"`
	ClosedAt     *time.Time `json:"closed_at,omitempty" validate:"omitempty"`
	DueOn        *time.Time `json:"due_on,omitempty" validate:"omitempty"`
}

// CommitData represents detailed commit information
type CommitData struct {
	SHA         string            `json:"sha" validate:"required,len=40,hexadecimal"`
	NodeID      string            `json:"node_id" validate:"required,min=1"`
	Commit      Commit            `json:"commit" validate:"required"`
	URL         string            `json:"url" validate:"required,url"`
	HTMLURL     string            `json:"html_url" validate:"required,url"`
	CommentsURL string            `json:"comments_url" validate:"required,url"`
	Author      *User             `json:"author,omitempty" validate:"omitempty"`
	Committer   *User             `json:"committer,omitempty" validate:"omitempty"`
	Parents     []CommitReference `json:"parents" validate:"omitempty,dive"`
	Stats       *CommitStats      `json:"stats,omitempty" validate:"omitempty"`
	Files       []CommitFile      `json:"files" validate:"omitempty,dive"`
}

// Commit represents the actual commit data
type Commit struct {
	Author       CommitAuthor    `json:"author" validate:"required"`
	Committer    CommitAuthor    `json:"committer" validate:"required"`
	Message      string          `json:"message" validate:"required,min=1,max=50000"`
	Tree         CommitReference `json:"tree" validate:"required"`
	URL          string          `json:"url" validate:"required,url"`
	CommentCount int             `json:"comment_count" validate:"gte=0"`
	Verification *Verification   `json:"verification,omitempty" validate:"omitempty"`
}

// CommitAuthor represents commit author/committer information
type CommitAuthor struct {
	Name  string    `json:"name" validate:"required,min=1,max=255"`
	Email string    `json:"email" validate:"required,email"`
	Date  time.Time `json:"date" validate:"required"`
}

// CommitReference represents a commit reference
type CommitReference struct {
	SHA string `json:"sha" validate:"required,len=40,hexadecimal"`
	URL string `json:"url" validate:"required,url"`
}

// CommitStats represents commit statistics
type CommitStats struct {
	Additions int `json:"additions" validate:"gte=0"`
	Deletions int `json:"deletions" validate:"gte=0"`
	Total     int `json:"total" validate:"gte=0"`
}

// CommitFile represents a file changed in a commit
type CommitFile struct {
	Filename    string  `json:"filename" validate:"required,min=1"`
	Additions   int     `json:"additions" validate:"gte=0"`
	Deletions   int     `json:"deletions" validate:"gte=0"`
	Changes     int     `json:"changes" validate:"gte=0"`
	Status      string  `json:"status" validate:"required,oneof=added removed modified renamed copied changed unchanged"`
	RawURL      string  `json:"raw_url" validate:"required,url"`
	BlobURL     string  `json:"blob_url" validate:"required,url"`
	Patch       *string `json:"patch,omitempty" validate:"omitempty"`
	SHA         string  `json:"sha" validate:"required,len=40,hexadecimal"`
	ContentsURL string  `json:"contents_url" validate:"required,url"`
}

// Verification represents commit verification information
type Verification struct {
	Verified  bool    `json:"verified"`
	Reason    string  `json:"reason" validate:"required,oneof=expired_key ocsp_pending ocsp_failure disabled_key unknown_signature_type unsigned unverified_email bad_email malformed_signature invalid gpg_error not_signing_key revoked_key no_user unknown_key bad_cert ocsp_revoked"`
	Signature *string `json:"signature,omitempty"`
	Payload   *string `json:"payload,omitempty"`
}
