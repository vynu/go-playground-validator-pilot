package models

import (
	"testing"
	"time"
)

func TestAPIRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request APIRequest
		wantErr bool
	}{
		{
			name:    "valid API request",
			request: getValidAPIRequest(),
			wantErr: false,
		},
		{
			name: "invalid method",
			request: func() APIRequest {
				req := getValidAPIRequest()
				req.Method = "INVALID"
				return req
			}(),
			wantErr: true,
		},
		{
			name: "invalid URL",
			request: func() APIRequest {
				req := getValidAPIRequest()
				req.URL = "not-a-url"
				return req
			}(),
			wantErr: true,
		},
		{
			name: "invalid IP",
			request: func() APIRequest {
				req := getValidAPIRequest()
				req.RemoteIP = "invalid.ip"
				return req
			}(),
			wantErr: true,
		},
		{
			name: "invalid source",
			request: func() APIRequest {
				req := getValidAPIRequest()
				req.Source = "invalid"
				return req
			}(),
			wantErr: true,
		},
		{
			name: "high retry count",
			request: func() APIRequest {
				req := getValidAPIRequest()
				req.RetryCount = 15
				return req
			}(),
			wantErr: true,
		},
	}

	validator := getTestValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("APIRequest validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAPIResponse_Validation(t *testing.T) {
	tests := []struct {
		name     string
		response APIResponse
		wantErr  bool
	}{
		{
			name:     "valid API response",
			response: getValidAPIResponse(),
			wantErr:  false,
		},
		{
			name: "invalid status code",
			response: func() APIResponse {
				resp := getValidAPIResponse()
				resp.StatusCode = 99
				return resp
			}(),
			wantErr: true,
		},
		{
			name: "negative content length",
			response: func() APIResponse {
				resp := getValidAPIResponse()
				length := int64(-1)
				resp.ContentLength = &length
				return resp
			}(),
			wantErr: true,
		},
		{
			name: "zero duration",
			response: func() APIResponse {
				resp := getValidAPIResponse()
				resp.Duration = 0
				return resp
			}(),
			wantErr: true,
		},
	}

	validator := getTestValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("APIResponse validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAPIAuthorization_Validation(t *testing.T) {
	tests := []struct {
		name string
		auth APIAuthorization
		want bool
	}{
		{
			name: "valid Bearer auth",
			auth: APIAuthorization{
				Type:  "Bearer",
				Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			},
			want: true,
		},
		{
			name: "invalid type",
			auth: APIAuthorization{
				Type:  "Invalid",
				Token: "token123",
			},
			want: false,
		},
		{
			name: "empty token",
			auth: APIAuthorization{
				Type:  "Bearer",
				Token: "",
			},
			want: false,
		},
		{
			name: "invalid algorithm",
			auth: APIAuthorization{
				Type:      "JWT",
				Token:     "token123",
				Algorithm: "INVALID",
			},
			want: false,
		},
	}

	validator := getTestValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(tt.auth)
			valid := err == nil
			if valid != tt.want {
				t.Errorf("APIAuthorization validation = %v, want %v, error: %v", valid, tt.want, err)
			}
		})
	}
}

func TestAPIRateLimit_Validation(t *testing.T) {
	tests := []struct {
		name string
		rate APIRateLimit
		want bool
	}{
		{
			name: "valid rate limit",
			rate: APIRateLimit{
				Limit:     1000,
				Remaining: 950,
				Reset:     time.Now().Add(time.Hour),
				Window:    "hour",
			},
			want: true,
		},
		{
			name: "zero limit",
			rate: APIRateLimit{
				Limit:     0,
				Remaining: 950,
				Reset:     time.Now().Add(time.Hour),
				Window:    "hour",
			},
			want: false,
		},
		{
			name: "negative remaining",
			rate: APIRateLimit{
				Limit:     1000,
				Remaining: -1,
				Reset:     time.Now().Add(time.Hour),
				Window:    "hour",
			},
			want: false,
		},
		{
			name: "invalid window",
			rate: APIRateLimit{
				Limit:     1000,
				Remaining: 950,
				Reset:     time.Now().Add(time.Hour),
				Window:    "invalid",
			},
			want: false,
		},
	}

	validator := getTestValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(tt.rate)
			valid := err == nil
			if valid != tt.want {
				t.Errorf("APIRateLimit validation = %v, want %v, error: %v", valid, tt.want, err)
			}
		})
	}
}

func TestAPIError_Validation(t *testing.T) {
	tests := []struct {
		name string
		err  APIError
		want bool
	}{
		{
			name: "valid API error",
			err: APIError{
				Code:      "VALIDATION_ERROR",
				Message:   "The request contains invalid data",
				Timestamp: time.Now(),
			},
			want: true,
		},
		{
			name: "empty code",
			err: APIError{
				Code:      "",
				Message:   "The request contains invalid data",
				Timestamp: time.Now(),
			},
			want: false,
		},
		{
			name: "empty message",
			err: APIError{
				Code:      "VALIDATION_ERROR",
				Message:   "",
				Timestamp: time.Now(),
			},
			want: false,
		},
		{
			name: "invalid docs URL",
			err: APIError{
				Code:      "VALIDATION_ERROR",
				Message:   "The request contains invalid data",
				Timestamp: time.Now(),
				DocsURL:   "not-a-url",
			},
			want: false,
		},
	}

	validator := getTestValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(tt.err)
			valid := err == nil
			if valid != tt.want {
				t.Errorf("APIError validation = %v, want %v, error: %v", valid, tt.want, err)
			}
		})
	}
}

func TestAPILink_Validation(t *testing.T) {
	tests := []struct {
		name string
		link APILink
		want bool
	}{
		{
			name: "valid API link",
			link: APILink{
				Rel:    "self",
				Href:   "https://api.example.com/users/123",
				Method: "GET",
			},
			want: true,
		},
		{
			name: "empty rel",
			link: APILink{
				Rel:  "",
				Href: "https://api.example.com/users/123",
			},
			want: false,
		},
		{
			name: "invalid URL",
			link: APILink{
				Rel:  "self",
				Href: "not-a-url",
			},
			want: false,
		},
		{
			name: "invalid method",
			link: APILink{
				Rel:    "self",
				Href:   "https://api.example.com/users/123",
				Method: "INVALID",
			},
			want: false,
		},
	}

	validator := getTestValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(tt.link)
			valid := err == nil
			if valid != tt.want {
				t.Errorf("APILink validation = %v, want %v, error: %v", valid, tt.want, err)
			}
		})
	}
}

func TestAPIPagination_Validation(t *testing.T) {
	tests := []struct {
		name       string
		pagination APIPagination
		want       bool
	}{
		{
			name: "valid pagination",
			pagination: APIPagination{
				Page:        1,
				PerPage:     20,
				Total:       100,
				TotalPages:  5,
				FirstPage:   1,
				LastPage:    5,
				HasNext:     true,
				HasPrevious: false,
			},
			want: true,
		},
		{
			name: "zero page",
			pagination: APIPagination{
				Page:      0,
				PerPage:   20,
				FirstPage: 1,
				LastPage:  5,
			},
			want: false,
		},
		{
			name: "per page too high",
			pagination: APIPagination{
				Page:      1,
				PerPage:   2000,
				FirstPage: 1,
				LastPage:  5,
			},
			want: false,
		},
		{
			name: "invalid first page",
			pagination: APIPagination{
				Page:      1,
				PerPage:   20,
				FirstPage: 0,
				LastPage:  5,
			},
			want: false,
		},
	}

	validator := getTestValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(tt.pagination)
			valid := err == nil
			if valid != tt.want {
				t.Errorf("APIPagination validation = %v, want %v, error: %v", valid, tt.want, err)
			}
		})
	}
}

func BenchmarkAPIRequest_Validation(b *testing.B) {
	request := getValidAPIRequest()
	validator := getTestValidator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Struct(request)
	}
}

// Helper functions

func getValidAPIResponse() APIResponse {
	now := time.Now()
	length := int64(256)

	return APIResponse{
		StatusCode:    200,
		Headers:       map[string]string{"Content-Type": "application/json"},
		ContentType:   "application/json",
		ContentLength: &length,
		Timestamp:     now,
		Duration:      time.Millisecond * 100,
		RequestID:     "req-123456",
		Version:       "v1",
	}
}
