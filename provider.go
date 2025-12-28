package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// apiClient holds configuration and auth token for talking to the backend API.
type apiClient struct {
	BaseURL     string
	Token       string
	HTTPClient  *http.Client
	RetryConfig RetryConfig
}

// loginRequest represents the request body for /login.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// loginResponse represents the response body from /login.
type loginResponse struct {
	Token string `json:"token"`
}

// Provider defines the vcluster Terraform provider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"base_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Base URL of vcluster API, e.g. http://192.168.1.4",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Username for login to vcluster API",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Password for login to vcluster API",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     300,
				Description: "HTTP client timeout in seconds (default: 300)",
			},
			"max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3,
				Description: "Maximum number of retries for failed requests (default: 3)",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"vcluster_cluster":        resourceCluster(),
			"vcluster_helm_release":   resourceHelmRelease(),
			"vcluster_orphan_cleanup": resourceOrphanCleanup(),
			"vcluster_secret":         resourceSecret(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"vcluster_cluster": dataSourceCluster(),
		},
		ConfigureContextFunc: func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
			baseURL := d.Get("base_url").(string)
			username := d.Get("username").(string)
			password := d.Get("password").(string)

			// Get optional configuration
			timeoutSeconds := d.Get("timeout").(int)
			maxRetries := d.Get("max_retries").(int)

			if timeoutSeconds <= 0 {
				timeoutSeconds = 300 // Default 5 minutes
			}
			if maxRetries < 0 {
				maxRetries = 3 // Default 3 retries
			}

			// Create HTTP client with proper timeouts
			httpClient := &http.Client{
				Timeout: time.Duration(timeoutSeconds) * time.Second,
				Transport: &http.Transport{
					IdleConnTimeout:       90 * time.Second,
					TLSHandshakeTimeout:   10 * time.Second,
					ExpectContinueTimeout: 1 * time.Second,
				},
			}

			// Configure retry settings
			retryConfig := RetryConfig{
				MaxRetries:        maxRetries,
				InitialDelay:      1 * time.Second,
				MaxDelay:          30 * time.Second,
				BackoffMultiplier: 2.0,
			}

			client := &apiClient{
				BaseURL:     baseURL,
				HTTPClient:  httpClient,
				RetryConfig: retryConfig,
			}

			// Perform login to obtain token.
			reqBody, err := json.Marshal(loginRequest{
				Username: username,
				Password: password,
			})
			if err != nil {
				return nil, diag.FromErr(err)
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/login", baseURL), bytes.NewReader(reqBody))
			if err != nil {
				return nil, diag.FromErr(err)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := httpClient.Do(req)
			if err != nil {
				return nil, diag.FromErr(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				b, _ := io.ReadAll(resp.Body)
				return nil, diag.Errorf("login failed: %s: %s", resp.Status, string(b))
			}

			var lr loginResponse
			if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
				return nil, diag.FromErr(err)
			}
			if lr.Token == "" {
				return nil, diag.Errorf("login succeeded but no token returned")
			}

			client.Token = lr.Token
			return client, nil
		},
	}
}
