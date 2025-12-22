package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffMultiplier float64
}

// DefaultRetryConfig returns sensible defaults for retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        3,
		InitialDelay:      1 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	
	// Retry on network errors, EOF, connection resets, and timeouts
	retryableErrors := []string{
		"EOF",
		"connection reset",
		"connection refused",
		"timeout",
		"temporary failure",
		"no such host",
		"network is unreachable",
	}
	
	for _, retryable := range retryableErrors {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(retryable)) {
			return true
		}
	}
	
	return false
}

// isRetryableStatusCode checks if an HTTP status code is retryable
func isRetryableStatusCode(statusCode int) bool {
	// Retry on 5xx errors and 429 (Too Many Requests)
	return statusCode >= 500 && statusCode < 600 || statusCode == 429
}

// doRequestWithRetry performs an HTTP request with retry logic
func doRequestWithRetry(ctx context.Context, client *apiClient, req *http.Request, retryConfig RetryConfig) (*http.Response, error) {
	var lastErr error
	delay := retryConfig.InitialDelay
	
	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			// Log retry attempt
			log.Printf("[WARN] Retrying request to %s (attempt %d/%d) after %v", req.URL.String(), attempt, retryConfig.MaxRetries, delay)
			
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			
			// Exponential backoff
			delay = time.Duration(float64(delay) * retryConfig.BackoffMultiplier)
			if delay > retryConfig.MaxDelay {
				delay = retryConfig.MaxDelay
			}
		}
		
		// Create a new request for each retry (request body can only be read once)
		var newReq *http.Request
		var err error
		
		// Try to get the body for retry
		var body io.Reader
		if req.GetBody != nil {
			bodyReader, bodyErr := req.GetBody()
			if bodyErr == nil {
				body = bodyReader
			}
		} else if req.Body != nil {
			// If GetBody is not available, read the body into a buffer
			// This is a fallback for requests that don't support GetBody
			bodyBytes, readErr := io.ReadAll(req.Body)
			if readErr == nil {
				body = bytes.NewReader(bodyBytes)
				// Restore original body for potential future reads
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}
		
		newReq, err = http.NewRequestWithContext(ctx, req.Method, req.URL.String(), body)
		if err != nil {
			return nil, fmt.Errorf("failed to create retry request: %w", err)
		}
		
		// Copy headers
		for k, v := range req.Header {
			newReq.Header[k] = v
		}
		
		// Perform the request
		resp, err := client.HTTPClient.Do(newReq)
		
		// Check for retryable errors
		if err != nil {
			lastErr = err
			if isRetryableError(err) && attempt < retryConfig.MaxRetries {
				continue
			}
			return nil, err
		}
		
		// Check for retryable status codes
		if isRetryableStatusCode(resp.StatusCode) && attempt < retryConfig.MaxRetries {
			// Read and close the response body before retrying
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("received retryable status code: %d", resp.StatusCode)
			continue
		}
		
		// Success or non-retryable error
		return resp, nil
	}
	
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// doRequestWithRetryDiag is a wrapper that returns diag.Diagnostics for Terraform
func doRequestWithRetryDiag(ctx context.Context, client *apiClient, req *http.Request, retryConfig RetryConfig) (*http.Response, diag.Diagnostics) {
	resp, err := doRequestWithRetry(ctx, client, req, retryConfig)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return resp, nil
}

