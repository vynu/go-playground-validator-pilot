// Package models contains API-specific models with comprehensive validation rules.
// This module defines structures for validating API requests, responses, and related data.
package models

import (
	"time"
)

// APIRequest represents a comprehensive API request validation structure.
type APIRequest struct {
	Method        string                 `json:"method" validate:"required,oneof=GET POST PUT PATCH DELETE HEAD OPTIONS TRACE CONNECT"`
	URL           string                 `json:"url" validate:"required,url"`
	Headers       map[string]string      `json:"headers" validate:"omitempty"`
	QueryParams   map[string]interface{} `json:"query_params" validate:"omitempty"`
	PathParams    map[string]string      `json:"path_params" validate:"omitempty"`
	Body          interface{}            `json:"body" validate:"omitempty"`
	Timestamp     time.Time              `json:"timestamp" validate:"required"`
	RequestID     string                 `json:"request_id" validate:"omitempty,min=1,max=255"`
	UserAgent     string                 `json:"user_agent" validate:"omitempty,max=1000"`
	RemoteIP      string                 `json:"remote_ip" validate:"omitempty,ip"`
	ContentType   string                 `json:"content_type" validate:"omitempty,api_content_type"`
	Accept        string                 `json:"accept" validate:"omitempty"`
	Authorization *APIAuthorization      `json:"authorization,omitempty" validate:"omitempty"`
	RateLimit     *APIRateLimit          `json:"rate_limit,omitempty" validate:"omitempty"`
	Timeout       *int                   `json:"timeout,omitempty" validate:"omitempty,gt=0,lte=300"`
	RetryCount    int                    `json:"retry_count" validate:"gte=0,lte=10"`
	Source        string                 `json:"source" validate:"omitempty,oneof=web mobile api cli automation test"`
	Version       string                 `json:"version" validate:"omitempty,api_version"`
	TraceID       string                 `json:"trace_id" validate:"omitempty"`
	SpanID        string                 `json:"span_id" validate:"omitempty"`
}

// APIResponse represents a comprehensive API response validation structure.
type APIResponse struct {
	StatusCode    int                    `json:"status_code" validate:"required,gte=100,lte=599"`
	Headers       map[string]string      `json:"headers" validate:"omitempty"`
	Body          interface{}            `json:"body" validate:"omitempty"`
	ContentType   string                 `json:"content_type" validate:"omitempty,api_content_type"`
	ContentLength *int64                 `json:"content_length,omitempty" validate:"omitempty,gte=0"`
	Timestamp     time.Time              `json:"timestamp" validate:"required"`
	Duration      time.Duration          `json:"duration" validate:"required,gt=0"`
	RequestID     string                 `json:"request_id" validate:"omitempty,min=1,max=255"`
	TraceID       string                 `json:"trace_id" validate:"omitempty"`
	SpanID        string                 `json:"span_id" validate:"omitempty"`
	CacheStatus   string                 `json:"cache_status" validate:"omitempty,oneof=hit miss bypass"`
	ServerID      string                 `json:"server_id" validate:"omitempty"`
	Version       string                 `json:"version" validate:"omitempty,api_version"`
	Errors        []APIError             `json:"errors,omitempty" validate:"omitempty,dive"`
	Warnings      []APIWarning           `json:"warnings,omitempty" validate:"omitempty,dive"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" validate:"omitempty"`
	Links         []APILink              `json:"links,omitempty" validate:"omitempty,dive"`
	Pagination    *APIPagination         `json:"pagination,omitempty" validate:"omitempty"`
}

// APIAuthorization represents API authorization information.
type APIAuthorization struct {
	Type      string                 `json:"type" validate:"required,oneof=Bearer Basic ApiKey OAuth JWT"`
	Token     string                 `json:"token" validate:"required,min=1"`
	Scheme    string                 `json:"scheme" validate:"omitempty"`
	Realm     string                 `json:"realm" validate:"omitempty"`
	Scope     []string               `json:"scope" validate:"omitempty,dive,min=1"`
	ExpiresAt *time.Time             `json:"expires_at,omitempty" validate:"omitempty"`
	Issuer    string                 `json:"issuer" validate:"omitempty,url"`
	Audience  []string               `json:"audience" validate:"omitempty,dive,min=1"`
	Subject   string                 `json:"subject" validate:"omitempty,min=1"`
	Claims    map[string]interface{} `json:"claims,omitempty" validate:"omitempty"`
	Algorithm string                 `json:"algorithm" validate:"omitempty,oneof=HS256 HS384 HS512 RS256 RS384 RS512 ES256 ES384 ES512"`
}

// APIRateLimit represents rate limiting information.
type APIRateLimit struct {
	Limit     int       `json:"limit" validate:"required,gt=0"`
	Remaining int       `json:"remaining" validate:"gte=0"`
	Reset     time.Time `json:"reset" validate:"required"`
	Window    string    `json:"window" validate:"required,oneof=second minute hour day month"`
	Policy    string    `json:"policy" validate:"omitempty,oneof=sliding_window fixed_window token_bucket"`
	Retry     *int      `json:"retry_after,omitempty" validate:"omitempty,gt=0"`
}

// APIError represents an API error with detailed information.
type APIError struct {
	Code        string                 `json:"code" validate:"required,min=1,max=100"`
	Message     string                 `json:"message" validate:"required,min=1,max=1000"`
	Detail      string                 `json:"detail" validate:"omitempty,max=5000"`
	Field       string                 `json:"field" validate:"omitempty,min=1,max=255"`
	Value       interface{}            `json:"value,omitempty"`
	Constraint  string                 `json:"constraint" validate:"omitempty,max=500"`
	Path        string                 `json:"path" validate:"omitempty,max=1000"`
	Timestamp   time.Time              `json:"timestamp" validate:"required"`
	RequestID   string                 `json:"request_id" validate:"omitempty,min=1,max=255"`
	TraceID     string                 `json:"trace_id" validate:"omitempty"`
	Context     map[string]interface{} `json:"context,omitempty" validate:"omitempty"`
	Suggestions []string               `json:"suggestions,omitempty" validate:"omitempty,dive,min=1,max=500"`
	DocsURL     string                 `json:"docs_url" validate:"omitempty,url"`
}

// APIWarning represents an API warning.
type APIWarning struct {
	Code      string                 `json:"code" validate:"required,min=1,max=100"`
	Message   string                 `json:"message" validate:"required,min=1,max=1000"`
	Detail    string                 `json:"detail" validate:"omitempty,max=5000"`
	Field     string                 `json:"field" validate:"omitempty,min=1,max=255"`
	Value     interface{}            `json:"value,omitempty"`
	Timestamp time.Time              `json:"timestamp" validate:"required"`
	Context   map[string]interface{} `json:"context,omitempty" validate:"omitempty"`
	DocsURL   string                 `json:"docs_url" validate:"omitempty,url"`
}

// APILink represents a HATEOAS link in API responses.
type APILink struct {
	Rel         string `json:"rel" validate:"required,min=1,max=100"`
	Href        string `json:"href" validate:"required,url"`
	Method      string `json:"method" validate:"omitempty,oneof=GET POST PUT PATCH DELETE HEAD OPTIONS"`
	Type        string `json:"type" validate:"omitempty,api_content_type"`
	Title       string `json:"title" validate:"omitempty,max=255"`
	Description string `json:"description" validate:"omitempty,max=1000"`
	Templated   bool   `json:"templated"`
}

// APIPagination represents pagination information in API responses.
type APIPagination struct {
	Page         int                 `json:"page" validate:"required,gte=1"`
	PerPage      int                 `json:"per_page" validate:"required,gte=1,lte=1000"`
	Total        int64               `json:"total" validate:"gte=0"`
	TotalPages   int                 `json:"total_pages" validate:"gte=0"`
	HasNext      bool                `json:"has_next"`
	HasPrevious  bool                `json:"has_previous"`
	NextPage     *int                `json:"next_page,omitempty" validate:"omitempty,gt=0"`
	PreviousPage *int                `json:"previous_page,omitempty" validate:"omitempty,gt=0"`
	FirstPage    int                 `json:"first_page" validate:"required,eq=1"`
	LastPage     int                 `json:"last_page" validate:"required,gte=1"`
	Links        *APIPaginationLinks `json:"links,omitempty" validate:"omitempty"`
}

// APIPaginationLinks represents pagination navigation links.
type APIPaginationLinks struct {
	First    string  `json:"first" validate:"required,url"`
	Last     string  `json:"last" validate:"required,url"`
	Next     *string `json:"next,omitempty" validate:"omitempty,url"`
	Previous *string `json:"previous,omitempty" validate:"omitempty,url"`
	Self     string  `json:"self" validate:"required,url"`
}

// APIWebhook represents webhook configuration and validation.
type APIWebhook struct {
	ID           string                 `json:"id" validate:"required,min=1,max=255"`
	URL          string                 `json:"url" validate:"required,url"`
	Events       []string               `json:"events" validate:"required,min=1,dive,min=1,max=100"`
	Secret       string                 `json:"secret" validate:"omitempty,min=8,max=255"`
	ContentType  string                 `json:"content_type" validate:"required,oneof=application/json application/x-www-form-urlencoded"`
	Active       bool                   `json:"active"`
	SSL          bool                   `json:"ssl"`
	CreatedAt    time.Time              `json:"created_at" validate:"required"`
	UpdatedAt    time.Time              `json:"updated_at" validate:"required,gtefield=CreatedAt"`
	LastDelivery *time.Time             `json:"last_delivery,omitempty" validate:"omitempty"`
	LastResponse *APIWebhookResponse    `json:"last_response,omitempty" validate:"omitempty"`
	Config       map[string]interface{} `json:"config" validate:"omitempty"`
	Headers      map[string]string      `json:"headers" validate:"omitempty"`
	Timeout      int                    `json:"timeout" validate:"omitempty,gt=0,lte=30"`
	Retries      int                    `json:"retries" validate:"omitempty,gte=0,lte=5"`
}

// APIWebhookResponse represents a webhook delivery response.
type APIWebhookResponse struct {
	StatusCode int               `json:"status_code" validate:"required,gte=100,lte=599"`
	Headers    map[string]string `json:"headers" validate:"omitempty"`
	Body       string            `json:"body" validate:"omitempty"`
	Duration   time.Duration     `json:"duration" validate:"required,gt=0"`
	Timestamp  time.Time         `json:"timestamp" validate:"required"`
	Success    bool              `json:"success"`
	Error      string            `json:"error" validate:"omitempty"`
	Attempt    int               `json:"attempt" validate:"required,gte=1,lte=10"`
}

// APIEndpoint represents API endpoint configuration.
type APIEndpoint struct {
	Path        string                     `json:"path" validate:"required,min=1,max=1000"`
	Method      string                     `json:"method" validate:"required,oneof=GET POST PUT PATCH DELETE HEAD OPTIONS"`
	Version     string                     `json:"version" validate:"omitempty,api_version"`
	Description string                     `json:"description" validate:"omitempty,max=1000"`
	Tags        []string                   `json:"tags" validate:"omitempty,dive,min=1,max=50"`
	Deprecated  bool                       `json:"deprecated"`
	Public      bool                       `json:"public"`
	RateLimit   *APIRateLimit              `json:"rate_limit,omitempty" validate:"omitempty"`
	Auth        *APIEndpointAuth           `json:"auth,omitempty" validate:"omitempty"`
	Parameters  []APIParameter             `json:"parameters" validate:"omitempty,dive"`
	RequestBody *APIRequestBody            `json:"request_body,omitempty" validate:"omitempty"`
	Responses   map[string]APIResponseSpec `json:"responses" validate:"omitempty"`
	Examples    []APIExample               `json:"examples" validate:"omitempty,dive"`
	CreatedAt   time.Time                  `json:"created_at" validate:"required"`
	UpdatedAt   time.Time                  `json:"updated_at" validate:"required,gtefield=CreatedAt"`
}

// APIEndpointAuth represents endpoint authentication requirements.
type APIEndpointAuth struct {
	Required bool     `json:"required"`
	Types    []string `json:"types" validate:"omitempty,dive,oneof=Bearer Basic ApiKey OAuth JWT"`
	Scopes   []string `json:"scopes" validate:"omitempty,dive,min=1"`
}

// APIParameter represents an API parameter.
type APIParameter struct {
	Name        string        `json:"name" validate:"required,min=1,max=255"`
	In          string        `json:"in" validate:"required,oneof=query header path cookie"`
	Description string        `json:"description" validate:"omitempty,max=1000"`
	Required    bool          `json:"required"`
	Type        string        `json:"type" validate:"required,oneof=string number integer boolean array object"`
	Format      string        `json:"format" validate:"omitempty"`
	Pattern     string        `json:"pattern" validate:"omitempty"`
	MinLength   *int          `json:"min_length,omitempty" validate:"omitempty,gte=0"`
	MaxLength   *int          `json:"max_length,omitempty" validate:"omitempty,gte=0"`
	Minimum     *float64      `json:"minimum,omitempty"`
	Maximum     *float64      `json:"maximum,omitempty"`
	Enum        []interface{} `json:"enum,omitempty" validate:"omitempty"`
	Default     interface{}   `json:"default,omitempty"`
	Example     interface{}   `json:"example,omitempty"`
}

// APIRequestBody represents request body specification.
type APIRequestBody struct {
	Description string                  `json:"description" validate:"omitempty,max=1000"`
	Required    bool                    `json:"required"`
	Content     map[string]APIMediaType `json:"content" validate:"required,min=1"`
}

// APIMediaType represents media type specification.
type APIMediaType struct {
	Schema   *APISchema             `json:"schema,omitempty" validate:"omitempty"`
	Example  interface{}            `json:"example,omitempty"`
	Examples map[string]APIExample  `json:"examples,omitempty" validate:"omitempty"`
	Encoding map[string]APIEncoding `json:"encoding,omitempty" validate:"omitempty"`
}

// APISchema represents JSON schema specification.
type APISchema struct {
	Type        string                `json:"type" validate:"omitempty,oneof=string number integer boolean array object null"`
	Format      string                `json:"format" validate:"omitempty"`
	Pattern     string                `json:"pattern" validate:"omitempty"`
	MinLength   *int                  `json:"min_length,omitempty" validate:"omitempty,gte=0"`
	MaxLength   *int                  `json:"max_length,omitempty" validate:"omitempty,gte=0"`
	Minimum     *float64              `json:"minimum,omitempty"`
	Maximum     *float64              `json:"maximum,omitempty"`
	Enum        []interface{}         `json:"enum,omitempty" validate:"omitempty"`
	Properties  map[string]*APISchema `json:"properties,omitempty" validate:"omitempty"`
	Items       *APISchema            `json:"items,omitempty" validate:"omitempty"`
	Required    []string              `json:"required,omitempty" validate:"omitempty,dive,min=1"`
	Description string                `json:"description" validate:"omitempty,max=1000"`
	Example     interface{}           `json:"example,omitempty"`
	Default     interface{}           `json:"default,omitempty"`
	ReadOnly    bool                  `json:"read_only"`
	WriteOnly   bool                  `json:"write_only"`
	Deprecated  bool                  `json:"deprecated"`
}

// APIResponseSpec represents response specification.
type APIResponseSpec struct {
	Description string                  `json:"description" validate:"required,min=1,max=1000"`
	Headers     map[string]APIParameter `json:"headers,omitempty" validate:"omitempty"`
	Content     map[string]APIMediaType `json:"content,omitempty" validate:"omitempty"`
	Links       map[string]APILink      `json:"links,omitempty" validate:"omitempty"`
}

// APIExample represents an API example.
type APIExample struct {
	Summary       string      `json:"summary" validate:"omitempty,max=255"`
	Description   string      `json:"description" validate:"omitempty,max=1000"`
	Value         interface{} `json:"value" validate:"required"`
	ExternalValue string      `json:"external_value" validate:"omitempty,url"`
}

// APIEncoding represents encoding specification.
type APIEncoding struct {
	ContentType   string                  `json:"content_type" validate:"omitempty,api_content_type"`
	Headers       map[string]APIParameter `json:"headers,omitempty" validate:"omitempty"`
	Style         string                  `json:"style" validate:"omitempty,oneof=form simple spaceDelimited pipeDelimited deepObject"`
	Explode       bool                    `json:"explode"`
	AllowReserved bool                    `json:"allow_reserved"`
}
