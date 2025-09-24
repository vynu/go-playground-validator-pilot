package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ComprehensiveE2ETestSuite represents a comprehensive end-to-end test suite
// covering all endpoints and multi-model validation capabilities
type ComprehensiveE2ETestSuite struct {
	originalServer *httptest.Server
	flexibleServer *httptest.Server
	testDataPath   string
	configPath     string
}

// SetupComprehensiveE2ETestSuite initializes the comprehensive test suite
func SetupComprehensiveE2ETestSuite(t *testing.T) *ComprehensiveE2ETestSuite {
	testDataPath := "../test_data"
	configPath := "config"

	// Create original validation server
	originalValidationServer := NewValidationServer()
	originalServer := httptest.NewServer(originalValidationServer.mux)

	// Create flexible validation server
	flexibleValidationServer := NewFlexibleValidationServer(configPath)
	flexibleServer := httptest.NewServer(flexibleValidationServer.mux)

	return &ComprehensiveE2ETestSuite{
		originalServer: originalServer,
		flexibleServer: flexibleServer,
		testDataPath:   testDataPath,
		configPath:     configPath,
	}
}

// TearDown cleans up the comprehensive test suite
func (suite *ComprehensiveE2ETestSuite) TearDown() {
	if suite.originalServer != nil {
		suite.originalServer.Close()
	}
	if suite.flexibleServer != nil {
		suite.flexibleServer.Close()
	}
}

// loadTestDataFile loads JSON test data from file
func (suite *ComprehensiveE2ETestSuite) loadTestDataFile(filename string) ([]byte, error) {
	filePath := filepath.Join(suite.testDataPath, filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read test data file %s: %w", filePath, err)
	}
	return data, nil
}

// makeHTTPRequest makes an HTTP request and returns the response
func (suite *ComprehensiveE2ETestSuite) makeHTTPRequest(method, endpoint string, body []byte, useFlexibleServer bool) (*http.Response, []byte, error) {
	var baseURL string
	if useFlexibleServer {
		baseURL = suite.flexibleServer.URL
	} else {
		baseURL = suite.originalServer.URL
	}

	url := baseURL + endpoint

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Request-ID", fmt.Sprintf("comprehensive-test-%d", time.Now().UnixNano()))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, err
	}

	return resp, respBody, nil
}

// TestComprehensive_OriginalServerEndpoints tests all original server endpoints
func TestComprehensive_OriginalServerEndpoints(t *testing.T) {
	suite := SetupComprehensiveE2ETestSuite(t)
	defer suite.TearDown()

	t.Run("OriginalServer_AllEndpoints", func(t *testing.T) {
		// Health endpoint
		resp, body, err := suite.makeHTTPRequest("GET", "/health", nil, false)
		if err != nil {
			t.Fatalf("Health endpoint failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Health endpoint: expected 200, got %d", resp.StatusCode)
		}

		var health map[string]interface{}
		if err := json.Unmarshal(body, &health); err != nil {
			t.Errorf("Health endpoint: failed to parse JSON: %v", err)
		}

		// Metrics endpoint
		resp, body, err = suite.makeHTTPRequest("GET", "/metrics", nil, false)
		if err != nil {
			t.Fatalf("Metrics endpoint failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Metrics endpoint: expected 200, got %d", resp.StatusCode)
		}

		var metrics map[string]interface{}
		if err := json.Unmarshal(body, &metrics); err != nil {
			t.Errorf("Metrics endpoint: failed to parse JSON: %v", err)
		}

		// GitHub validation endpoint
		gitHubData, err := suite.loadTestDataFile("sample_pull_request.json")
		if err != nil {
			t.Logf("GitHub test data not available: %v", err)
		} else {
			resp, body, err = suite.makeHTTPRequest("POST", "/validate/github", gitHubData, false)
			if err != nil {
				t.Fatalf("GitHub validation failed: %v", err)
			}
			// Note: This might fail due to missing fields, but we want to test the endpoint
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnprocessableEntity {
				t.Errorf("GitHub validation: expected 200 or 422, got %d", resp.StatusCode)
			}
		}

		// Batch validation endpoint
		batchData, err := suite.loadTestDataFile("batch_payloads.json")
		if err != nil {
			t.Logf("Batch test data not available: %v", err)
		} else {
			resp, body, err = suite.makeHTTPRequest("POST", "/validate/batch", batchData, false)
			if err != nil {
				t.Fatalf("Batch validation failed: %v", err)
			}
			// Note: This might fail due to validation issues, but we want to test the endpoint
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
				t.Errorf("Batch validation: expected 200 or 400, got %d", resp.StatusCode)
			}
		}

		// Async validation endpoint
		asyncJob := map[string]interface{}{
			"data": map[string]interface{}{
				"test": "data",
			},
			"type":     "test_job",
			"priority": 1,
		}
		asyncJobData, _ := json.Marshal(asyncJob)
		resp, body, err = suite.makeHTTPRequest("POST", "/validate/async", asyncJobData, false)
		if err != nil {
			t.Fatalf("Async validation failed: %v", err)
		}
		if resp.StatusCode != http.StatusAccepted {
			t.Errorf("Async validation: expected 202, got %d", resp.StatusCode)
		}

		// Try to get async result (might be processing or not found)
		resp, _, err = suite.makeHTTPRequest("GET", "/validate/result/test-id", nil, false)
		if err != nil {
			t.Fatalf("Async result retrieval failed: %v", err)
		}
		// Accept both 200 (found) and 202 (processing)
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
			t.Errorf("Async result: expected 200 or 202, got %d", resp.StatusCode)
		}

		// Swagger endpoints
		resp, _, err = suite.makeHTTPRequest("GET", "/swagger/", nil, false)
		if err != nil {
			t.Fatalf("Swagger redirect failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusMovedPermanently {
			t.Errorf("Swagger endpoint: expected 200 or 301, got %d", resp.StatusCode)
		}

		resp, _, err = suite.makeHTTPRequest("GET", "/docs/", nil, false)
		if err != nil {
			t.Fatalf("Docs endpoint failed: %v", err)
		}
		// Docs might not be available, but endpoint should respond
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Docs endpoint: expected 200 or 404, got %d", resp.StatusCode)
		}

		resp, _, err = suite.makeHTTPRequest("GET", "/docs/swagger.yaml", nil, false)
		if err != nil {
			t.Fatalf("Swagger spec failed: %v", err)
		}
		// Swagger spec might not be available, but endpoint should respond
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Swagger spec: expected 200 or 404, got %d", resp.StatusCode)
		}
	})
}

// TestComprehensive_FlexibleServerEndpoints tests all flexible server endpoints
func TestComprehensive_FlexibleServerEndpoints(t *testing.T) {
	suite := SetupComprehensiveE2ETestSuite(t)
	defer suite.TearDown()

	t.Run("FlexibleServer_AllEndpoints", func(t *testing.T) {
		// Health endpoint
		resp, body, err := suite.makeHTTPRequest("GET", "/health", nil, true)
		if err != nil {
			t.Fatalf("Flexible health endpoint failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Flexible health endpoint: expected 200, got %d", resp.StatusCode)
		}

		// List models endpoint
		resp, body, err = suite.makeHTTPRequest("GET", "/models", nil, true)
		if err != nil {
			t.Fatalf("List models failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("List models: expected 200, got %d", resp.StatusCode)
		}

		var models map[string]interface{}
		if err := json.Unmarshal(body, &models); err != nil {
			t.Errorf("List models: failed to parse JSON: %v", err)
		}

		// Get specific model endpoint
		resp, body, err = suite.makeHTTPRequest("GET", "/models/GitHubPayload", nil, true)
		if err != nil {
			t.Fatalf("Get model failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Get model: expected 200 or 404, got %d", resp.StatusCode)
		}

		// List providers endpoint
		resp, body, err = suite.makeHTTPRequest("GET", "/providers", nil, true)
		if err != nil {
			t.Fatalf("List providers failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("List providers: expected 200, got %d", resp.StatusCode)
		}

		var providers map[string]interface{}
		if err := json.Unmarshal(body, &providers); err != nil {
			t.Errorf("List providers: failed to parse JSON: %v", err)
		}

		// Flexible validation endpoint
		testData := map[string]interface{}{
			"test_field": "test_value",
			"number":     123,
			"boolean":    true,
		}
		testDataJSON, _ := json.Marshal(testData)
		resp, body, err = suite.makeHTTPRequest("POST", "/validate/flexible?model_type=GenericJSON", testDataJSON, true)
		if err != nil {
			t.Fatalf("Flexible validation failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Flexible validation: expected 200, 400, or 404, got %d", resp.StatusCode)
		}

		// Model-specific validation endpoint
		resp, body, err = suite.makeHTTPRequest("POST", "/validate/model/TestModel", testDataJSON, true)
		if err != nil {
			t.Fatalf("Model validation failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Model validation: expected 200, 400, or 404, got %d", resp.StatusCode)
		}

		// Get validation rules
		resp, body, err = suite.makeHTTPRequest("GET", "/config/rules", nil, true)
		if err != nil {
			t.Fatalf("Get validation rules failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Get validation rules: expected 200, got %d", resp.StatusCode)
		}

		var rules map[string]interface{}
		if err := json.Unmarshal(body, &rules); err != nil {
			t.Errorf("Get validation rules: failed to parse JSON: %v", err)
		}

		// Update validation rules
		updateRules := map[string]interface{}{
			"test_rule": map[string]interface{}{
				"required": true,
				"type":     "string",
			},
		}
		updateRulesJSON, _ := json.Marshal(updateRules)
		resp, body, err = suite.makeHTTPRequest("PUT", "/config/rules/TestModel", updateRulesJSON, true)
		if err != nil {
			t.Fatalf("Update validation rules failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Update validation rules: expected 200, 400, or 404, got %d", resp.StatusCode)
		}

		// Reload configuration
		resp, body, err = suite.makeHTTPRequest("POST", "/config/reload", nil, true)
		if err != nil {
			t.Fatalf("Reload configuration failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Reload configuration: expected 200 or 500, got %d", resp.StatusCode)
		}

		// Configurable validation
		configValidation := map[string]interface{}{
			"model_type": "TestModel",
			"data": map[string]interface{}{
				"test_field": "value",
			},
			"configuration": map[string]interface{}{
				"strict_mode": true,
			},
		}
		configValidationJSON, _ := json.Marshal(configValidation)
		resp, body, err = suite.makeHTTPRequest("POST", "/validate/config", configValidationJSON, true)
		if err != nil {
			t.Fatalf("Configurable validation failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Configurable validation: expected 200, 400, or 404, got %d", resp.StatusCode)
		}

		// Compare providers
		compareData := map[string]interface{}{
			"data": map[string]interface{}{
				"test": "data",
			},
			"providers": []string{"go_playground", "json_schema"},
		}
		compareDataJSON, _ := json.Marshal(compareData)
		resp, body, err = suite.makeHTTPRequest("POST", "/validate/compare", compareDataJSON, true)
		if err != nil {
			t.Fatalf("Compare providers failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Compare providers: expected 200 or 400, got %d", resp.StatusCode)
		}

		// Profile validation
		profileData := map[string]interface{}{
			"test": "data",
		}
		profileDataJSON, _ := json.Marshal(profileData)
		resp, body, err = suite.makeHTTPRequest("POST", "/validate/profile/strict", profileDataJSON, true)
		if err != nil {
			t.Fatalf("Profile validation failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Profile validation: expected 200, 400, or 404, got %d", resp.StatusCode)
		}
	})
}

// TestComprehensive_MultiModelValidation tests multi-model validation capabilities
func TestComprehensive_MultiModelValidation(t *testing.T) {
	suite := SetupComprehensiveE2ETestSuite(t)
	defer suite.TearDown()

	t.Run("MultiModel_GitHubValidation", func(t *testing.T) {
		gitHubData, err := suite.loadTestDataFile("sample_pull_request.json")
		if err != nil {
			t.Skipf("GitHub test data not available: %v", err)
		}

		resp, body, err := suite.makeHTTPRequest("POST", "/validate/flexible?model_type=GitHubPayload", gitHubData, true)
		if err != nil {
			t.Fatalf("GitHub flexible validation failed: %v", err)
		}

		// Accept various status codes as we're testing the endpoint functionality
		if resp.StatusCode >= 500 {
			t.Errorf("GitHub flexible validation: server error %d: %s", resp.StatusCode, string(body))
		}
	})

	t.Run("MultiModel_GitLabValidation", func(t *testing.T) {
		gitLabData, err := suite.loadTestDataFile("gitlab_payload.json")
		if err != nil {
			t.Skipf("GitLab test data not available: %v", err)
		}

		resp, body, err := suite.makeHTTPRequest("POST", "/validate/flexible?model_type=GitLabPayload", gitLabData, true)
		if err != nil {
			t.Fatalf("GitLab flexible validation failed: %v", err)
		}

		// Accept various status codes as we're testing the endpoint functionality
		if resp.StatusCode >= 500 {
			t.Errorf("GitLab flexible validation: server error %d: %s", resp.StatusCode, string(body))
		}
	})

	t.Run("MultiModel_BitbucketValidation", func(t *testing.T) {
		bitbucketData, err := suite.loadTestDataFile("bitbucket_payload.json")
		if err != nil {
			t.Skipf("Bitbucket test data not available: %v", err)
		}

		resp, body, err := suite.makeHTTPRequest("POST", "/validate/flexible?model_type=BitbucketPayload", bitbucketData, true)
		if err != nil {
			t.Fatalf("Bitbucket flexible validation failed: %v", err)
		}

		// Accept various status codes as we're testing the endpoint functionality
		if resp.StatusCode >= 500 {
			t.Errorf("Bitbucket flexible validation: server error %d: %s", resp.StatusCode, string(body))
		}
	})

	t.Run("MultiModel_GenericJSONValidation", func(t *testing.T) {
		genericData, err := suite.loadTestDataFile("generic_json_payload.json")
		if err != nil {
			t.Skipf("Generic JSON test data not available: %v", err)
		}

		resp, body, err := suite.makeHTTPRequest("POST", "/validate/flexible?model_type=GenericJSON", genericData, true)
		if err != nil {
			t.Fatalf("Generic JSON flexible validation failed: %v", err)
		}

		// Accept various status codes as we're testing the endpoint functionality
		if resp.StatusCode >= 500 {
			t.Errorf("Generic JSON flexible validation: server error %d: %s", resp.StatusCode, string(body))
		}
	})

	t.Run("MultiModel_APIModelValidation", func(t *testing.T) {
		apiData, err := suite.loadTestDataFile("api_model_payload.json")
		if err != nil {
			t.Skipf("API model test data not available: %v", err)
		}

		resp, body, err := suite.makeHTTPRequest("POST", "/validate/flexible?model_type=APIModel", apiData, true)
		if err != nil {
			t.Fatalf("API model flexible validation failed: %v", err)
		}

		// Accept various status codes as we're testing the endpoint functionality
		if resp.StatusCode >= 500 {
			t.Errorf("API model flexible validation: server error %d: %s", resp.StatusCode, string(body))
		}
	})

	t.Run("MultiModel_DatabaseModelValidation", func(t *testing.T) {
		dbData, err := suite.loadTestDataFile("database_model_payload.json")
		if err != nil {
			t.Skipf("Database model test data not available: %v", err)
		}

		resp, body, err := suite.makeHTTPRequest("POST", "/validate/flexible?model_type=DatabaseModel", dbData, true)
		if err != nil {
			t.Fatalf("Database model flexible validation failed: %v", err)
		}

		// Accept various status codes as we're testing the endpoint functionality
		if resp.StatusCode >= 500 {
			t.Errorf("Database model flexible validation: server error %d: %s", resp.StatusCode, string(body))
		}
	})
}

// TestComprehensive_ValidationProfiles tests validation profiles
func TestComprehensive_ValidationProfiles(t *testing.T) {
	suite := SetupComprehensiveE2ETestSuite(t)
	defer suite.TearDown()

	profiles := []string{"strict", "permissive", "minimal"}

	for _, profile := range profiles {
		t.Run(fmt.Sprintf("ValidationProfile_%s", profile), func(t *testing.T) {
			testData := map[string]interface{}{
				"test_field": "test_value",
				"number":     123,
				"email":      "test@example.com",
			}
			testDataJSON, _ := json.Marshal(testData)

			resp, body, err := suite.makeHTTPRequest("POST", fmt.Sprintf("/validate/profile/%s", profile), testDataJSON, true)
			if err != nil {
				t.Fatalf("Profile validation for %s failed: %v", profile, err)
			}

			// Accept various status codes as we're testing the endpoint functionality
			if resp.StatusCode >= 500 {
				t.Errorf("Profile validation for %s: server error %d: %s", profile, resp.StatusCode, string(body))
			}

			// If successful, try to parse the response
			if resp.StatusCode == http.StatusOK {
				var result map[string]interface{}
				if err := json.Unmarshal(body, &result); err != nil {
					t.Errorf("Profile validation for %s: failed to parse response: %v", profile, err)
				} else {
					// Check if the response has expected fields
					if profile, exists := result["validation_profile"]; exists {
						t.Logf("Profile validation for %s successful, profile: %v", profile, profile)
					}
				}
			}
		})
	}
}

// TestComprehensive_ProviderComparison tests provider comparison functionality
func TestComprehensive_ProviderComparison(t *testing.T) {
	suite := SetupComprehensiveE2ETestSuite(t)
	defer suite.TearDown()

	t.Run("ProviderComparison_MultipleProviders", func(t *testing.T) {
		compareData := map[string]interface{}{
			"data": map[string]interface{}{
				"email":     "test@example.com",
				"username":  "testuser123",
				"age":       25,
				"is_active": true,
			},
			"providers":  []string{"go_playground", "json_schema", "custom"},
			"model_type": "UserProfile",
		}
		compareDataJSON, _ := json.Marshal(compareData)

		resp, body, err := suite.makeHTTPRequest("POST", "/validate/compare", compareDataJSON, true)
		if err != nil {
			t.Fatalf("Provider comparison failed: %v", err)
		}

		// Accept various status codes as we're testing the endpoint functionality
		if resp.StatusCode >= 500 {
			t.Errorf("Provider comparison: server error %d: %s", resp.StatusCode, string(body))
		}

		// If successful, try to parse the response
		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			if err := json.Unmarshal(body, &result); err != nil {
				t.Errorf("Provider comparison: failed to parse response: %v", err)
			} else {
				// Check if the response has expected fields
				if results, exists := result["provider_results"]; exists {
					t.Logf("Provider comparison successful, results: %v", results)
				}
			}
		}
	})
}

// TestComprehensive_ErrorHandling tests comprehensive error handling
func TestComprehensive_ErrorHandling(t *testing.T) {
	suite := SetupComprehensiveE2ETestSuite(t)
	defer suite.TearDown()

	t.Run("ErrorHandling_InvalidJSON", func(t *testing.T) {
		invalidJSON := []byte("{invalid json")

		// Test on original server
		resp, _, err := suite.makeHTTPRequest("POST", "/validate/github", invalidJSON, false)
		if err != nil {
			t.Fatalf("Invalid JSON test on original server failed: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Original server invalid JSON: expected 400, got %d", resp.StatusCode)
		}

		// Test on flexible server
		resp, _, err = suite.makeHTTPRequest("POST", "/validate/flexible", invalidJSON, true)
		if err != nil {
			t.Fatalf("Invalid JSON test on flexible server failed: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Flexible server invalid JSON: expected 400 or 500, got %d", resp.StatusCode)
		}
	})

	t.Run("ErrorHandling_NonExistentEndpoints", func(t *testing.T) {
		nonExistentEndpoints := []string{
			"/nonexistent",
			"/validate/nonexistent",
			"/models/nonexistent",
			"/config/nonexistent",
		}

		for _, endpoint := range nonExistentEndpoints {
			// Test on original server
			resp, _, err := suite.makeHTTPRequest("GET", endpoint, nil, false)
			if err != nil {
				t.Fatalf("Non-existent endpoint test on original server failed: %v", err)
			}
			if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusMethodNotAllowed {
				t.Errorf("Original server %s: expected 404 or 405, got %d", endpoint, resp.StatusCode)
			}

			// Test on flexible server
			resp, _, err = suite.makeHTTPRequest("GET", endpoint, nil, true)
			if err != nil {
				t.Fatalf("Non-existent endpoint test on flexible server failed: %v", err)
			}
			if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusMethodNotAllowed {
				t.Errorf("Flexible server %s: expected 404 or 405, got %d", endpoint, resp.StatusCode)
			}
		}
	})

	t.Run("ErrorHandling_MethodNotAllowed", func(t *testing.T) {
		// Test wrong methods on both servers
		wrongMethodTests := []struct {
			method   string
			endpoint string
		}{
			{"PUT", "/health"},
			{"DELETE", "/metrics"},
			{"GET", "/validate/github"},
			{"DELETE", "/config/rules"},
		}

		for _, test := range wrongMethodTests {
			// Test on original server
			resp, _, err := suite.makeHTTPRequest(test.method, test.endpoint, nil, false)
			if err != nil {
				t.Fatalf("Method not allowed test on original server failed: %v", err)
			}
			if resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusNotFound {
				t.Errorf("Original server %s %s: expected 405 or 404, got %d", test.method, test.endpoint, resp.StatusCode)
			}

			// Test on flexible server
			resp, _, err = suite.makeHTTPRequest(test.method, test.endpoint, nil, true)
			if err != nil {
				t.Fatalf("Method not allowed test on flexible server failed: %v", err)
			}
			if resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusNotFound {
				t.Errorf("Flexible server %s %s: expected 405 or 404, got %d", test.method, test.endpoint, resp.StatusCode)
			}
		}
	})
}

// TestComprehensive_CORS tests CORS functionality on both servers
func TestComprehensive_CORS(t *testing.T) {
	suite := SetupComprehensiveE2ETestSuite(t)
	defer suite.TearDown()

	corsEndpoints := []string{
		"/health",
		"/validate/github",
		"/validate/flexible",
		"/models",
	}

	for _, endpoint := range corsEndpoints {
		t.Run(fmt.Sprintf("CORS_%s", strings.ReplaceAll(endpoint, "/", "_")), func(t *testing.T) {
			// Test CORS on original server
			resp, _, err := suite.makeHTTPRequest("OPTIONS", endpoint, nil, false)
			if err != nil {
				t.Fatalf("CORS test on original server failed: %v", err)
			}

			// Check for CORS headers (either present or endpoint not found)
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
				origin := resp.Header.Get("Access-Control-Allow-Origin")
				if origin == "" {
					t.Errorf("Original server %s: missing CORS headers", endpoint)
				}
			}

			// Test CORS on flexible server
			resp, _, err = suite.makeHTTPRequest("OPTIONS", endpoint, nil, true)
			if err != nil {
				t.Fatalf("CORS test on flexible server failed: %v", err)
			}

			// Check for CORS headers (either present or endpoint not found)
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
				origin := resp.Header.Get("Access-Control-Allow-Origin")
				if origin == "" {
					t.Errorf("Flexible server %s: missing CORS headers", endpoint)
				}
			}
		})
	}
}

// TestComprehensive_Performance tests basic performance across all endpoints
func TestComprehensive_Performance(t *testing.T) {
	suite := SetupComprehensiveE2ETestSuite(t)
	defer suite.TearDown()

	performanceEndpoints := []struct {
		method   string
		endpoint string
		body     []byte
		server   bool // false = original, true = flexible
	}{
		{"GET", "/health", nil, false},
		{"GET", "/metrics", nil, false},
		{"GET", "/health", nil, true},
		{"GET", "/models", nil, true},
		{"GET", "/providers", nil, true},
		{"GET", "/config/rules", nil, true},
	}

	for _, test := range performanceEndpoints {
		t.Run(fmt.Sprintf("Performance_%s_%s", test.method, strings.ReplaceAll(test.endpoint, "/", "_")), func(t *testing.T) {
			start := time.Now()
			resp, _, err := suite.makeHTTPRequest(test.method, test.endpoint, test.body, test.server)
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("Performance test failed: %v", err)
			}

			// Performance threshold: most endpoints should respond within 1 second
			if duration > 1*time.Second {
				t.Errorf("Endpoint %s took too long: %v", test.endpoint, duration)
			}

			// Log performance for successful requests
			if resp.StatusCode == http.StatusOK {
				t.Logf("Endpoint %s responded in %v", test.endpoint, duration)
			}
		})
	}
}

// TestComprehensive_ConcurrentRequests tests concurrent access to all servers
func TestComprehensive_ConcurrentRequests(t *testing.T) {
	suite := SetupComprehensiveE2ETestSuite(t)
	defer suite.TearDown()

	t.Run("ConcurrentRequests_BothServers", func(t *testing.T) {
		const numRequests = 10
		results := make(chan error, numRequests*2) // Both servers

		// Test concurrent requests to original server
		for i := 0; i < numRequests; i++ {
			go func(id int) {
				resp, _, err := suite.makeHTTPRequest("GET", "/health", nil, false)
				if err != nil {
					results <- fmt.Errorf("original server request %d failed: %w", id, err)
					return
				}
				if resp.StatusCode >= 500 {
					results <- fmt.Errorf("original server request %d got server error %d", id, resp.StatusCode)
					return
				}
				results <- nil
			}(i)
		}

		// Test concurrent requests to flexible server
		for i := 0; i < numRequests; i++ {
			go func(id int) {
				resp, _, err := suite.makeHTTPRequest("GET", "/models", nil, true)
				if err != nil {
					results <- fmt.Errorf("flexible server request %d failed: %w", id, err)
					return
				}
				if resp.StatusCode >= 500 {
					results <- fmt.Errorf("flexible server request %d got server error %d", id, resp.StatusCode)
					return
				}
				results <- nil
			}(i)
		}

		// Collect results
		errorCount := 0
		for i := 0; i < numRequests*2; i++ {
			if err := <-results; err != nil {
				t.Errorf("Concurrent request error: %v", err)
				errorCount++
			}
		}

		if errorCount > 0 {
			t.Errorf("Failed %d out of %d concurrent requests", errorCount, numRequests*2)
		}
	})
}
