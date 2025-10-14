package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGitHubPayload_Validation(t *testing.T) {
	tests := []struct {
		name    string
		payload GitHubPayload
		wantErr bool
	}{
		{
			name:    "valid github payload",
			payload: getValidGitHubPayload(),
			wantErr: false,
		},
		{
			name: "missing action",
			payload: func() GitHubPayload {
				p := getValidGitHubPayload()
				p.Action = ""
				return p
			}(),
			wantErr: true,
		},
		{
			name: "invalid action",
			payload: func() GitHubPayload {
				p := getValidGitHubPayload()
				p.Action = "invalid"
				return p
			}(),
			wantErr: true,
		},
		{
			name: "zero number",
			payload: func() GitHubPayload {
				p := getValidGitHubPayload()
				p.Number = 0
				return p
			}(),
			wantErr: true,
		},
	}

	validator := getTestValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitHubPayload validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPullRequest_Validation(t *testing.T) {
	tests := []struct {
		name string
		pr   PullRequest
		want bool
	}{
		{
			name: "valid pull request",
			pr:   getValidPullRequest(),
			want: true,
		},
		{
			name: "invalid state",
			pr: func() PullRequest {
				pr := getValidPullRequest()
				pr.State = "invalid"
				return pr
			}(),
			want: false,
		},
		{
			name: "empty title",
			pr: func() PullRequest {
				pr := getValidPullRequest()
				pr.Title = ""
				return pr
			}(),
			want: false,
		},
		{
			name: "invalid commit SHA",
			pr: func() PullRequest {
				pr := getValidPullRequest()
				sha := "invalid"
				pr.MergeCommitSHA = &sha
				return pr
			}(),
			want: false,
		},
		{
			name: "negative commits",
			pr: func() PullRequest {
				pr := getValidPullRequest()
				pr.Commits = -1
				return pr
			}(),
			want: false,
		},
	}

	validator := getTestValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(tt.pr)
			valid := err == nil
			if valid != tt.want {
				t.Errorf("PullRequest validation = %v, want %v, error: %v", valid, tt.want, err)
			}
		})
	}
}

func TestRepository_Validation(t *testing.T) {
	tests := []struct {
		name string
		repo Repository
		want bool
	}{
		{
			name: "valid repository",
			repo: getValidRepository(),
			want: true,
		},
		{
			name: "empty name",
			repo: func() Repository {
				repo := getValidRepository()
				repo.Name = ""
				return repo
			}(),
			want: false,
		},
		{
			name: "invalid visibility",
			repo: func() Repository {
				repo := getValidRepository()
				repo.Visibility = "invalid"
				return repo
			}(),
			want: false,
		},
		{
			name: "negative stargazers",
			repo: func() Repository {
				repo := getValidRepository()
				repo.StargazersCount = -1
				return repo
			}(),
			want: false,
		},
	}

	validator := getTestValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(tt.repo)
			valid := err == nil
			if valid != tt.want {
				t.Errorf("Repository validation = %v, want %v, error: %v", valid, tt.want, err)
			}
		})
	}
}

func TestUser_Validation(t *testing.T) {
	tests := []struct {
		name string
		user User
		want bool
	}{
		{
			name: "valid user",
			user: getValidUser(),
			want: true,
		},
		{
			name: "empty login",
			user: func() User {
				user := getValidUser()
				user.Login = ""
				return user
			}(),
			want: false,
		},
		{
			name: "invalid type",
			user: func() User {
				user := getValidUser()
				user.Type = "Invalid"
				return user
			}(),
			want: false,
		},
		{
			name: "invalid email",
			user: func() User {
				user := getValidUser()
				email := "invalid-email"
				user.Email = &email
				return user
			}(),
			want: false,
		},
	}

	validator := getTestValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(tt.user)
			valid := err == nil
			if valid != tt.want {
				t.Errorf("User validation = %v, want %v, error: %v", valid, tt.want, err)
			}
		})
	}
}

func TestReference_Validation(t *testing.T) {
	tests := []struct {
		name string
		ref  Reference
		want bool
	}{
		{
			name: "valid reference",
			ref:  getValidReference(),
			want: true,
		},
		{
			name: "invalid SHA length",
			ref: func() Reference {
				ref := getValidReference()
				ref.SHA = "invalid"
				return ref
			}(),
			want: false,
		},
		{
			name: "non-hex SHA",
			ref: func() Reference {
				ref := getValidReference()
				ref.SHA = "gggggggggggggggggggggggggggggggggggggggg"
				return ref
			}(),
			want: false,
		},
	}

	validator := getTestValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(tt.ref)
			valid := err == nil
			if valid != tt.want {
				t.Errorf("Reference validation = %v, want %v, error: %v", valid, tt.want, err)
			}
		})
	}
}

func TestLabel_Validation(t *testing.T) {
	tests := []struct {
		name  string
		label Label
		want  bool
	}{
		{
			name:  "valid label",
			label: getValidLabel(),
			want:  true,
		},
		{
			name: "invalid color",
			label: func() Label {
				label := getValidLabel()
				label.Color = "invalid"
				return label
			}(),
			want: false,
		},
		{
			name: "empty name",
			label: func() Label {
				label := getValidLabel()
				label.Name = ""
				return label
			}(),
			want: false,
		},
	}

	validator := getTestValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(tt.label)
			valid := err == nil
			if valid != tt.want {
				t.Errorf("Label validation = %v, want %v, error: %v", valid, tt.want, err)
			}
		})
	}
}

func TestGitHubPayload_JSONMarshaling(t *testing.T) {
	payload := getValidGitHubPayload()

	// Test JSON marshaling
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal GitHub payload: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled GitHubPayload
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal GitHub payload: %v", err)
	}

	// Compare key fields
	if payload.Action != unmarshaled.Action {
		t.Errorf("Action mismatch: got %s, want %s", unmarshaled.Action, payload.Action)
	}
	if payload.Number != unmarshaled.Number {
		t.Errorf("Number mismatch: got %d, want %d", unmarshaled.Number, payload.Number)
	}
}

func BenchmarkGitHubPayload_Validation(b *testing.B) {
	payload := getValidGitHubPayload()
	validator := getTestValidator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Struct(payload)
	}
}

// Helper functions

func getValidGitHubPayload() GitHubPayload {
	return GitHubPayload{
		Action:      "opened",
		Number:      123,
		PullRequest: getValidPullRequest(),
		Repository:  getValidRepository(),
		Sender:      getValidUser(),
	}
}

func getValidPullRequest() PullRequest {
	now := time.Now()
	validSHA := "a1b2c3d4e5f6789012345678901234567890abcd"
	body := "This is a valid pull request body"

	return PullRequest{
		ID:                12345,
		NodeID:            "PR_kwDOABCDEFGHIJKLMNOP",
		Number:            123,
		State:             "open",
		Title:             "Add new feature",
		Body:              &body,
		CreatedAt:         now,
		UpdatedAt:         now,
		MergeCommitSHA:    &validSHA,
		CommitsURL:        "https://api.github.com/repos/owner/repo/pulls/123/commits",
		ReviewCommentsURL: "https://api.github.com/repos/owner/repo/pulls/123/comments",
		CommentsURL:       "https://api.github.com/repos/owner/repo/issues/123/comments",
		StatusesURL:       "https://api.github.com/repos/owner/repo/statuses/" + validSHA,
		Head:              getValidReference(),
		Base:              getValidReference(),
		User:              getValidUser(),
		Comments:          0,
		ReviewComments:    0,
		Commits:           1,
		Additions:         10,
		Deletions:         0,
		ChangedFiles:      1,
	}
}

func getValidRepository() Repository {
	now := time.Now()
	desc := "A test repository"

	return Repository{
		ID:              123456,
		NodeID:          "R_kgDOABCDEFGHIJKLMNOP",
		Name:            "test-repo",
		FullName:        "owner/test-repo",
		Private:         false,
		Owner:           getValidUser(),
		HTMLURL:         "https://github.com/owner/test-repo",
		Description:     &desc,
		URL:             "https://api.github.com/repos/owner/test-repo",
		CreatedAt:       now,
		UpdatedAt:       now,
		PushedAt:        now,
		GitURL:          "git://github.com/owner/test-repo.git",
		SSHURL:          "git@github.com:owner/test-repo.git",
		CloneURL:        "https://github.com/owner/test-repo.git",
		Size:            100,
		StargazersCount: 10,
		WatchersCount:   5,
		ForksCount:      2,
		OpenIssuesCount: 1,
		Forks:           2,
		OpenIssues:      1,
		Watchers:        5,
		DefaultBranch:   "main",
		Visibility:      "public",
	}
}

func getValidUser() User {
	name := "Test User"
	email := "test@example.com"

	return User{
		Login:             "testuser",
		ID:                123456,
		NodeID:            "U_kgDOABCDEFGHIJKLMNOP",
		AvatarURL:         "https://github.com/images/error/testuser_happy.gif",
		URL:               "https://api.github.com/users/testuser",
		HTMLURL:           "https://github.com/testuser",
		FollowersURL:      "https://api.github.com/users/testuser/followers",
		FollowingURL:      "https://api.github.com/users/testuser/following{/other_user}",
		GistsURL:          "https://api.github.com/users/testuser/gists{/gist_id}",
		StarredURL:        "https://api.github.com/users/testuser/starred{/owner}{/repo}",
		SubscriptionsURL:  "https://api.github.com/users/testuser/subscriptions",
		OrganizationsURL:  "https://api.github.com/users/testuser/orgs",
		ReposURL:          "https://api.github.com/users/testuser/repos",
		EventsURL:         "https://api.github.com/users/testuser/events{/privacy}",
		ReceivedEventsURL: "https://api.github.com/users/testuser/received_events",
		Type:              "User",
		Name:              &name,
		Email:             &email,
	}
}

func getValidReference() Reference {
	return Reference{
		Label: "owner:feature-branch",
		Ref:   "feature-branch",
		SHA:   "a1b2c3d4e5f6789012345678901234567890abcd",
		User:  getValidUser(),
		Repo:  getValidRepository(),
	}
}

func getValidLabel() Label {
	desc := "Enhancement label"

	return Label{
		ID:          123456,
		NodeID:      "L_kwDOABCDEFGHIJKLMNOP",
		URL:         "https://api.github.com/repos/owner/repo/labels/enhancement",
		Name:        "enhancement",
		Color:       "a2eeef",
		Description: &desc,
	}
}
