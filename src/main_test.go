package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestValidationServer_Health tests the health endpoint
func TestValidationServer_Health(t *testing.T) {
	server := NewValidationServer()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	var health map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &health)
	if err != nil {
		t.Errorf("Failed to unmarshal health response: %v", err)
	}

	if health["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", health["status"])
	}

	expectedFields := []string{"status", "timestamp", "version", "go_version", "workers", "requests", "uptime"}
	for _, field := range expectedFields {
		if _, exists := health[field]; !exists {
			t.Errorf("Missing field in health response: %s", field)
		}
	}
}

// TestValidationServer_Metrics tests the metrics endpoint
func TestValidationServer_Metrics(t *testing.T) {
	server := NewValidationServer()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	server.handleMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	var metrics map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &metrics)
	if err != nil {
		t.Errorf("Failed to unmarshal metrics response: %v", err)
	}

	expectedFields := []string{"requests_total", "workers_active", "queue_size", "pending_results", "goroutines", "gomaxprocs", "uptime_seconds"}
	for _, field := range expectedFields {
		if _, exists := metrics[field]; !exists {
			t.Errorf("Missing field in metrics response: %s", field)
		}
	}
}

// TestValidationServer_GitHubValidation_ValidPayload tests valid GitHub payload validation
func TestValidationServer_GitHubValidation_ValidPayload(t *testing.T) {
	server := NewValidationServer()

	// Create a minimal valid payload
	payload := GitHubPayload{
		Action: "opened",
		Number: 123,
		PullRequest: PullRequest{
			ID:                987654321,
			NodeID:            "PR_test123",
			Number:            123,
			State:             "open",
			Title:             "Test Pull Request",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			CommitsURL:        "https://api.github.com/repos/test/test/pulls/123/commits",
			ReviewCommentsURL: "https://api.github.com/repos/test/test/pulls/123/comments",
			CommentsURL:       "https://api.github.com/repos/test/test/issues/123/comments",
			StatusesURL:       "https://api.github.com/repos/test/test/statuses/sha123",
			Head: Reference{
				Label: "feature-branch",
				Ref:   "refs/heads/feature-branch",
				SHA:   "1234567890123456789012345678901234567890",
				User:  createTestUser(),
				Repo:  createTestRepository(),
			},
			Base: Reference{
				Label: "main",
				Ref:   "refs/heads/main",
				SHA:   "0987654321098765432109876543210987654321",
				User:  createTestUser(),
				Repo:  createTestRepository(),
			},
			User:         createTestUser(),
			Commits:      1,
			Additions:    10,
			Deletions:    5,
			ChangedFiles: 2,
		},
		Repository: createTestRepository(),
		Sender:     createTestUser(),
	}

	jsonData, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/validate/github", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "test-123")
	w := httptest.NewRecorder()

	server.handleGitHubValidation(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	var result ValidationResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	if err != nil {
		t.Errorf("Failed to unmarshal validation response: %v", err)
	}

	if !result.IsValid {
		t.Errorf("Expected valid payload, got invalid with errors: %v", result.Errors)
	}

	if result.ID != "test-123" {
		t.Errorf("Expected request ID 'test-123', got %s", result.ID)
	}
}

// TestValidationServer_GitHubValidation_InvalidPayload tests invalid GitHub payload validation
func TestValidationServer_GitHubValidation_InvalidPayload(t *testing.T) {
	server := NewValidationServer()

	// Create an invalid payload (missing required fields)
	payload := map[string]interface{}{
		"action": "invalid_action",
		"number": -1,
	}

	jsonData, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/validate/github", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleGitHubValidation(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Expected status code 422, got %d", w.Code)
	}

	var result ValidationResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	if err != nil {
		t.Errorf("Failed to unmarshal validation response: %v", err)
	}

	if result.IsValid {
		t.Errorf("Expected invalid payload, got valid")
	}

	if len(result.Errors) == 0 {
		t.Errorf("Expected validation errors, got none")
	}
}

// TestValidationServer_AsyncValidation tests async validation endpoint
func TestValidationServer_AsyncValidation(t *testing.T) {
	server := NewValidationServer()

	job := ValidationJob{
		Type:     "github_webhook",
		Data:     map[string]interface{}{"action": "opened", "number": 123},
		Priority: 1,
	}

	jsonData, _ := json.Marshal(job)
	req := httptest.NewRequest("POST", "/validate/async", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAsyncValidation(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status code 202, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal async response: %v", err)
	}

	if response["status"] != "queued" {
		t.Errorf("Expected status 'queued', got %v", response["status"])
	}

	if response["id"] == nil {
		t.Errorf("Expected job ID, got nil")
	}
}

// TestValidationServer_BatchValidation tests batch validation endpoint
func TestValidationServer_BatchValidation(t *testing.T) {
	server := NewValidationServer()

	batch := struct {
		Payloads []GitHubPayload `json:"payloads"`
	}{
		Payloads: []GitHubPayload{
			createTestGitHubPayload(),
		},
	}

	jsonData, _ := json.Marshal(batch)
	req := httptest.NewRequest("POST", "/validate/batch", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleBatchValidation(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal batch response: %v", err)
	}

	if response["total"] == nil {
		t.Errorf("Expected total field, got nil")
	}

	if response["results"] == nil {
		t.Errorf("Expected results field, got nil")
	}
}

// TestValidationServer_SwaggerEndpoints tests Swagger documentation endpoints
func TestValidationServer_SwaggerEndpoints(t *testing.T) {
	server := NewValidationServer()

	tests := []struct {
		endpoint   string
		method     string
		statusCode int
	}{
		{"/swagger/", "GET", http.StatusPermanentRedirect},
		{"/docs/swagger.yaml", "GET", http.StatusOK},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.endpoint, nil)
		w := httptest.NewRecorder()

		switch tt.endpoint {
		case "/swagger/":
			server.handleSwaggerRedirect(w, req)
		case "/docs/swagger.yaml":
			server.handleSwaggerSpec(w, req)
		}

		if w.Code != tt.statusCode {
			t.Errorf("Endpoint %s: expected status code %d, got %d", tt.endpoint, tt.statusCode, w.Code)
		}
	}
}

// TestValidationServer_Middleware tests the middleware functionality
func TestValidationServer_Middleware(t *testing.T) {
	server := NewValidationServer()

	// Test CORS headers
	req := httptest.NewRequest("OPTIONS", "/validate/github", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler := server.withMiddleware(server.handleGitHubValidation)
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200 for OPTIONS, got %d", w.Code)
	}

	corsHeader := w.Header().Get("Access-Control-Allow-Origin")
	if corsHeader != "*" {
		t.Errorf("Expected CORS header '*', got %s", corsHeader)
	}

	// Test security headers
	req = httptest.NewRequest("GET", "/health", nil)
	w = httptest.NewRecorder()

	handler = server.withMiddleware(server.handleHealth)
	handler(w, req)

	securityHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
	}

	for header, expectedValue := range securityHeaders {
		actualValue := w.Header().Get(header)
		if actualValue != expectedValue {
			t.Errorf("Expected security header %s: %s, got %s", header, expectedValue, actualValue)
		}
	}
}

// TestWorkerPool tests the worker pool functionality
func TestWorkerPool(t *testing.T) {
	workerPool := NewWorkerPool(2)
	defer workerPool.cancel()

	// Test job submission
	job := ValidationJob{
		ID:   "test-job-1",
		Type: "test",
		Data: map[string]interface{}{"test": "data"},
	}

	select {
	case workerPool.jobQueue <- job:
		// Job submitted successfully
	case <-time.After(1 * time.Second):
		t.Errorf("Failed to submit job to worker pool")
	}

	// Test result retrieval (with timeout)
	select {
	case result := <-workerPool.resultChan:
		if result.ID != job.ID {
			t.Errorf("Expected result ID %s, got %s", job.ID, result.ID)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("Failed to receive result from worker pool")
	}
}

// TestCustomValidators tests custom validation functions
func TestCustomValidators(t *testing.T) {
	server := NewValidationServer()

	// Test GitHub username validation
	testCases := []struct {
		username string
		valid    bool
	}{
		{"valid-user", true},
		{"user123", true},
		{"", false},
		{"a", true},
		{strings.Repeat("a", 40), false}, // Too long
		{"user with spaces", false},      // Invalid characters
		{"-invalid", false},              // Cannot start with hyphen
	}

	for _, tc := range testCases {
		user := User{Login: tc.username, ID: 123, Type: "User"}
		user.NodeID = "test"
		user.AvatarURL = "https://example.com/avatar"
		user.URL = "https://example.com/user"
		user.HTMLURL = "https://example.com/user"
		user.FollowersURL = "https://example.com/followers"
		user.FollowingURL = "https://example.com/following"
		user.GistsURL = "https://example.com/gists"
		user.StarredURL = "https://example.com/starred"
		user.SubscriptionsURL = "https://example.com/subscriptions"
		user.OrganizationsURL = "https://example.com/orgs"
		user.ReposURL = "https://example.com/repos"
		user.EventsURL = "https://example.com/events"
		user.ReceivedEventsURL = "https://example.com/received"

		err := server.validator.Struct(user)
		isValid := err == nil

		if isValid != tc.valid {
			t.Errorf("Username %s: expected valid=%v, got valid=%v, error=%v", tc.username, tc.valid, isValid, err)
		}
	}
}

// TestValidationResult_BusinessLogic tests business logic validation
func TestValidationResult_BusinessLogic(t *testing.T) {
	server := NewValidationServer()

	// Test WIP detection
	wipPayload := createTestGitHubPayload()
	wipPayload.PullRequest.Title = "WIP: Work in progress"

	warnings := server.performBusinessValidation(&wipPayload)

	hasWIPWarning := false
	for _, warning := range warnings {
		if warning.Code == "WIP_DETECTED" {
			hasWIPWarning = true
			break
		}
	}

	if !hasWIPWarning {
		t.Errorf("Expected WIP warning for title with 'WIP:', got none")
	}

	// Test large changeset warning
	largePayload := createTestGitHubPayload()
	largePayload.PullRequest.Additions = 1500
	largePayload.PullRequest.Deletions = 500

	warnings = server.performBusinessValidation(&largePayload)

	hasLargeChangesetWarning := false
	for _, warning := range warnings {
		if warning.Code == "LARGE_CHANGESET" {
			hasLargeChangesetWarning = true
			break
		}
	}

	if !hasLargeChangesetWarning {
		t.Errorf("Expected large changeset warning for >1000 changes, got none")
	}
}

// Helper function to create a test user
func createTestUser() User {
	return User{
		Login:             "testuser",
		ID:                12345,
		NodeID:            "MDQ6VXNlcjEyMzQ1",
		AvatarURL:         "https://avatars.githubusercontent.com/u/12345?v=4",
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
		SiteAdmin:         false,
	}
}

// Helper function to create a test repository
func createTestRepository() Repository {
	return Repository{
		ID:            555555,
		NodeID:        "MDEwOlJlcG9zaXRvcnk1NTU1NTU=",
		Name:          "test-repo",
		FullName:      "testuser/test-repo",
		Private:       false,
		Owner:         createTestUser(),
		HTMLURL:       "https://github.com/testuser/test-repo",
		Fork:          false,
		URL:           "https://api.github.com/repos/testuser/test-repo",
		CreatedAt:     time.Now().AddDate(-1, 0, 0),
		UpdatedAt:     time.Now(),
		PushedAt:      time.Now(),
		GitURL:        "git://github.com/testuser/test-repo.git",
		SSHURL:        "git@github.com:testuser/test-repo.git",
		CloneURL:      "https://github.com/testuser/test-repo.git",
		DefaultBranch: "main",
		Visibility:    "public",
	}
}

// Helper function to create a test GitHub payload
func createTestGitHubPayload() GitHubPayload {
	now := time.Now()
	return GitHubPayload{
		Action: "opened",
		Number: 123,
		PullRequest: PullRequest{
			ID:                987654321,
			NodeID:            "PR_kwDOABCD1234567890",
			Number:            123,
			State:             "open",
			Title:             "Test Pull Request",
			CreatedAt:         now,
			UpdatedAt:         now,
			CommitsURL:        "https://api.github.com/repos/testuser/test-repo/pulls/123/commits",
			ReviewCommentsURL: "https://api.github.com/repos/testuser/test-repo/pulls/123/comments",
			CommentsURL:       "https://api.github.com/repos/testuser/test-repo/issues/123/comments",
			StatusesURL:       "https://api.github.com/repos/testuser/test-repo/statuses/1234567890123456789012345678901234567890",
			Head: Reference{
				Label: "feature-branch",
				Ref:   "refs/heads/feature-branch",
				SHA:   "1234567890123456789012345678901234567890",
				User:  createTestUser(),
				Repo:  createTestRepository(),
			},
			Base: Reference{
				Label: "main",
				Ref:   "refs/heads/main",
				SHA:   "0987654321098765432109876543210987654321",
				User:  createTestUser(),
				Repo:  createTestRepository(),
			},
			User:         createTestUser(),
			Commits:      1,
			Additions:    10,
			Deletions:    5,
			ChangedFiles: 2,
		},
		Repository: createTestRepository(),
		Sender:     createTestUser(),
	}
}

// Benchmark tests for performance validation
func BenchmarkValidationServer_GitHubValidation(b *testing.B) {
	server := NewValidationServer()
	payload := createTestGitHubPayload()
	jsonData, _ := json.Marshal(payload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/validate/github", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.handleGitHubValidation(w, req)
	}
}

func BenchmarkValidationServer_Health(b *testing.B) {
	server := NewValidationServer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		server.handleHealth(w, req)
	}
}
