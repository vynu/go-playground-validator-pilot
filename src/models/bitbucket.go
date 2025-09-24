// Package models contains Bitbucket webhook payload models with comprehensive validation rules.
// This module defines all Bitbucket-specific data structures for webhook validation.
package models

import (
	"time"
)

// BitbucketPayload represents the top-level Bitbucket webhook payload structure.
type BitbucketPayload struct {
	Repository  BitbucketRepository   `json:"repository" validate:"required"`
	Actor       BitbucketUser         `json:"actor" validate:"required"`
	PullRequest *BitbucketPullRequest `json:"pullrequest,omitempty" validate:"omitempty"`
	Push        *BitbucketPush        `json:"push,omitempty" validate:"omitempty"`
	Comment     *BitbucketComment     `json:"comment,omitempty" validate:"omitempty"`
	Approval    *BitbucketApproval    `json:"approval,omitempty" validate:"omitempty"`
	Changes     []BitbucketChange     `json:"changes,omitempty" validate:"omitempty,dive"`
}

// BitbucketRepository represents a Bitbucket repository structure.
type BitbucketRepository struct {
	UUID       string            `json:"uuid" validate:"required,min=1"`
	Name       string            `json:"name" validate:"required,min=1,max=100"`
	FullName   string            `json:"full_name" validate:"required,min=1,max=200"`
	IsPrivate  bool              `json:"is_private"`
	Type       string            `json:"type" validate:"required,oneof=repository"`
	Owner      BitbucketUser     `json:"owner" validate:"required"`
	Website    *string           `json:"website,omitempty" validate:"omitempty,url"`
	Language   string            `json:"language" validate:"omitempty"`
	HasIssues  bool              `json:"has_issues"`
	HasWiki    bool              `json:"has_wiki"`
	CreatedOn  time.Time         `json:"created_on" validate:"required"`
	UpdatedOn  time.Time         `json:"updated_on" validate:"required,gtefield=CreatedOn"`
	Size       int64             `json:"size" validate:"gte=0"`
	ForkPolicy string            `json:"fork_policy" validate:"omitempty,oneof=allow_forks no_public_forks no_forks"`
	Project    *BitbucketProject `json:"project,omitempty" validate:"omitempty"`
	Links      BitbucketLinks    `json:"links" validate:"required"`
	MainBranch *BitbucketBranch  `json:"mainbranch,omitempty" validate:"omitempty"`
}

// BitbucketUser represents a Bitbucket user or team.
type BitbucketUser struct {
	UUID        string             `json:"uuid" validate:"required,min=1"`
	Username    string             `json:"username" validate:"required,bitbucket_username"`
	DisplayName string             `json:"display_name" validate:"required,min=1,max=255"`
	AccountID   string             `json:"account_id" validate:"required,min=1"`
	Type        string             `json:"type" validate:"required,oneof=user team"`
	Links       BitbucketUserLinks `json:"links" validate:"required"`
	Nickname    *string            `json:"nickname,omitempty" validate:"omitempty,max=50"`
	Website     *string            `json:"website,omitempty" validate:"omitempty,url"`
	Location    *string            `json:"location,omitempty" validate:"omitempty,max=255"`
	CreatedOn   *time.Time         `json:"created_on,omitempty" validate:"omitempty"`
}

// BitbucketProject represents a Bitbucket project.
type BitbucketProject struct {
	UUID        string         `json:"uuid" validate:"required,min=1"`
	Key         string         `json:"key" validate:"required,min=1,max=10"`
	Name        string         `json:"name" validate:"required,min=1,max=255"`
	Description *string        `json:"description,omitempty" validate:"omitempty,max=1000"`
	IsPrivate   bool           `json:"is_private"`
	Type        string         `json:"type" validate:"required,oneof=project"`
	Owner       BitbucketUser  `json:"owner" validate:"required"`
	Links       BitbucketLinks `json:"links" validate:"required"`
	CreatedOn   time.Time      `json:"created_on" validate:"required"`
	UpdatedOn   time.Time      `json:"updated_on" validate:"required,gtefield=CreatedOn"`
}

// BitbucketPullRequest represents a Bitbucket pull request structure.
type BitbucketPullRequest struct {
	ID                int64                  `json:"id" validate:"required,gt=0"`
	Title             string                 `json:"title" validate:"required,min=1,max=255"`
	Description       string                 `json:"description" validate:"omitempty,max=100000"`
	State             string                 `json:"state" validate:"required,oneof=OPEN MERGED DECLINED SUPERSEDED"`
	Author            BitbucketUser          `json:"author" validate:"required"`
	Source            BitbucketBranch        `json:"source" validate:"required"`
	Destination       BitbucketBranch        `json:"destination" validate:"required"`
	MergeCommit       *BitbucketCommit       `json:"merge_commit,omitempty" validate:"omitempty"`
	CommentCount      int                    `json:"comment_count" validate:"gte=0"`
	TaskCount         int                    `json:"task_count" validate:"gte=0"`
	Type              string                 `json:"type" validate:"required,oneof=pullrequest"`
	Reason            string                 `json:"reason" validate:"omitempty"`
	CreatedOn         time.Time              `json:"created_on" validate:"required"`
	UpdatedOn         time.Time              `json:"updated_on" validate:"required,gtefield=CreatedOn"`
	Reviewers         []BitbucketReviewer    `json:"reviewers" validate:"omitempty,dive"`
	Participants      []BitbucketParticipant `json:"participants" validate:"omitempty,dive"`
	Links             BitbucketPRLinks       `json:"links" validate:"required"`
	CloseSourceBranch bool                   `json:"close_source_branch"`
	ClosedBy          *BitbucketUser         `json:"closed_by,omitempty" validate:"omitempty"`
}

// BitbucketBranch represents a Git branch in Bitbucket.
type BitbucketBranch struct {
	Name       string              `json:"name" validate:"required,min=1"`
	Commit     BitbucketCommit     `json:"commit" validate:"required"`
	Repository BitbucketRepository `json:"repository" validate:"required"`
}

// BitbucketCommit represents a Git commit in Bitbucket.
type BitbucketCommit struct {
	Hash    string               `json:"hash" validate:"required,len=40,hexadecimal"`
	Type    string               `json:"type" validate:"required,oneof=commit"`
	Message string               `json:"message" validate:"required,min=1"`
	Date    time.Time            `json:"date" validate:"required"`
	Author  BitbucketCommitUser  `json:"author" validate:"required"`
	Parents []BitbucketCommitRef `json:"parents" validate:"omitempty,dive"`
	Links   BitbucketCommitLinks `json:"links" validate:"required"`
}

// BitbucketCommitUser represents a commit author.
type BitbucketCommitUser struct {
	Raw  string         `json:"raw" validate:"required,min=1"`
	Type string         `json:"type" validate:"required,oneof=author"`
	User *BitbucketUser `json:"user,omitempty" validate:"omitempty"`
}

// BitbucketCommitRef represents a commit reference.
type BitbucketCommitRef struct {
	Hash  string               `json:"hash" validate:"required,len=40,hexadecimal"`
	Type  string               `json:"type" validate:"required,oneof=commit"`
	Links BitbucketCommitLinks `json:"links" validate:"required"`
}

// BitbucketReviewer represents a pull request reviewer.
type BitbucketReviewer struct {
	User     BitbucketUser `json:"user" validate:"required"`
	Role     string        `json:"role" validate:"required,oneof=REVIEWER PARTICIPANT"`
	Approved bool          `json:"approved"`
	Type     string        `json:"type" validate:"required,oneof=participant"`
}

// BitbucketParticipant represents a pull request participant.
type BitbucketParticipant struct {
	User     BitbucketUser `json:"user" validate:"required"`
	Role     string        `json:"role" validate:"required,oneof=REVIEWER PARTICIPANT"`
	Approved bool          `json:"approved"`
	State    *string       `json:"state,omitempty" validate:"omitempty,oneof=approved changes_requested"`
	Type     string        `json:"type" validate:"required,oneof=participant"`
}

// BitbucketPush represents push information.
type BitbucketPush struct {
	Changes []BitbucketChange `json:"changes" validate:"required,min=1,dive"`
}

// BitbucketChange represents a change in a push.
type BitbucketChange struct {
	New       *BitbucketChangeRef  `json:"new,omitempty" validate:"omitempty"`
	Old       *BitbucketChangeRef  `json:"old,omitempty" validate:"omitempty"`
	Created   bool                 `json:"created"`
	Forced    bool                 `json:"forced"`
	Closed    bool                 `json:"closed"`
	Commits   []BitbucketCommit    `json:"commits" validate:"omitempty,dive"`
	Truncated bool                 `json:"truncated"`
	Links     BitbucketChangeLinks `json:"links" validate:"required"`
}

// BitbucketChangeRef represents a reference in a change.
type BitbucketChangeRef struct {
	Type   string            `json:"type" validate:"required,oneof=branch tag"`
	Name   string            `json:"name" validate:"required,min=1"`
	Target BitbucketCommit   `json:"target" validate:"required"`
	Links  BitbucketRefLinks `json:"links" validate:"required"`
}

// BitbucketComment represents a comment on a pull request.
type BitbucketComment struct {
	ID        int64                 `json:"id" validate:"required,gt=0"`
	Content   BitbucketContent      `json:"content" validate:"required"`
	User      BitbucketUser         `json:"user" validate:"required"`
	CreatedOn time.Time             `json:"created_on" validate:"required"`
	UpdatedOn *time.Time            `json:"updated_on,omitempty" validate:"omitempty,gtefield=CreatedOn"`
	Type      string                `json:"type" validate:"required,oneof=pullrequest_comment"`
	Inline    *BitbucketInline      `json:"inline,omitempty" validate:"omitempty"`
	Parent    *BitbucketComment     `json:"parent,omitempty" validate:"omitempty"`
	Links     BitbucketCommentLinks `json:"links" validate:"required"`
}

// BitbucketContent represents comment content.
type BitbucketContent struct {
	Raw    string `json:"raw" validate:"required"`
	Markup string `json:"markup" validate:"required,oneof=markdown plaintext"`
	HTML   string `json:"html" validate:"required"`
	Type   string `json:"type" validate:"required,oneof=rendered"`
}

// BitbucketInline represents inline comment information.
type BitbucketInline struct {
	From *int   `json:"from,omitempty" validate:"omitempty,gte=1"`
	To   int    `json:"to" validate:"required,gte=1"`
	Path string `json:"path" validate:"required,min=1"`
}

// BitbucketApproval represents a pull request approval.
type BitbucketApproval struct {
	User        BitbucketUser        `json:"user" validate:"required"`
	Date        time.Time            `json:"date" validate:"required"`
	Type        string               `json:"type" validate:"required,oneof=approval"`
	PullRequest BitbucketPullRequest `json:"pullrequest" validate:"required"`
}

// Links structures for various Bitbucket entities

// BitbucketLinks represents general links structure.
type BitbucketLinks struct {
	Self   BitbucketLink        `json:"self" validate:"required"`
	HTML   BitbucketLink        `json:"html" validate:"required"`
	Avatar *BitbucketLink       `json:"avatar,omitempty" validate:"omitempty"`
	Clone  []BitbucketCloneLink `json:"clone,omitempty" validate:"omitempty,dive"`
}

// BitbucketUserLinks represents user-specific links.
type BitbucketUserLinks struct {
	Self         BitbucketLink  `json:"self" validate:"required"`
	HTML         BitbucketLink  `json:"html" validate:"required"`
	Avatar       BitbucketLink  `json:"avatar" validate:"required"`
	Followers    *BitbucketLink `json:"followers,omitempty" validate:"omitempty"`
	Following    *BitbucketLink `json:"following,omitempty" validate:"omitempty"`
	Repositories *BitbucketLink `json:"repositories,omitempty" validate:"omitempty"`
}

// BitbucketPRLinks represents pull request-specific links.
type BitbucketPRLinks struct {
	Self     BitbucketLink  `json:"self" validate:"required"`
	HTML     BitbucketLink  `json:"html" validate:"required"`
	Diff     BitbucketLink  `json:"diff" validate:"required"`
	Diffstat BitbucketLink  `json:"diffstat" validate:"required"`
	Comments BitbucketLink  `json:"comments" validate:"required"`
	Activity BitbucketLink  `json:"activity" validate:"required"`
	Merge    *BitbucketLink `json:"merge,omitempty" validate:"omitempty"`
	Decline  *BitbucketLink `json:"decline,omitempty" validate:"omitempty"`
}

// BitbucketCommitLinks represents commit-specific links.
type BitbucketCommitLinks struct {
	Self BitbucketLink  `json:"self" validate:"required"`
	HTML BitbucketLink  `json:"html" validate:"required"`
	Diff *BitbucketLink `json:"diff,omitempty" validate:"omitempty"`
}

// BitbucketChangeLinks represents change-specific links.
type BitbucketChangeLinks struct {
	Diff    *BitbucketLink `json:"diff,omitempty" validate:"omitempty"`
	Commits BitbucketLink  `json:"commits" validate:"required"`
	HTML    BitbucketLink  `json:"html" validate:"required"`
}

// BitbucketRefLinks represents reference-specific links.
type BitbucketRefLinks struct {
	Self    BitbucketLink `json:"self" validate:"required"`
	HTML    BitbucketLink `json:"html" validate:"required"`
	Commits BitbucketLink `json:"commits" validate:"required"`
}

// BitbucketCommentLinks represents comment-specific links.
type BitbucketCommentLinks struct {
	Self BitbucketLink `json:"self" validate:"required"`
	HTML BitbucketLink `json:"html" validate:"required"`
}

// BitbucketLink represents a single link with href.
type BitbucketLink struct {
	Href string  `json:"href" validate:"required,url"`
	Name *string `json:"name,omitempty" validate:"omitempty"`
}

// BitbucketCloneLink represents a clone link with name and href.
type BitbucketCloneLink struct {
	Name string `json:"name" validate:"required,oneof=https ssh"`
	Href string `json:"href" validate:"required,min=1"`
}
