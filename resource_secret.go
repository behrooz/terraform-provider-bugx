package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SecretPayload represents the JSON body sent to create/update secrets.
type SecretPayload struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Data        map[string]string `json:"data"`
}

// SecretInfo represents the JSON structure returned from the API.
type SecretInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Data        map[string]string `json:"data"`
	CreatedAt   string            `json:"createdAt,omitempty"`
	UpdatedAt   string            `json:"updatedAt,omitempty"`
}

// SecretsListResponse represents the response from GET /secrets/api/v1/secrets.
type SecretsListResponse struct {
	Secrets []SecretInfo `json:"secrets"`
}

// resourceSecret defines the bugx_secret resource schema and CRUD.
func resourceSecret() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretCreate,
		ReadContext:   resourceSecretRead,
		UpdateContext: resourceSecretUpdate,
		DeleteContext: resourceSecretDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the secret (must be unique)",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional description of the secret",
			},
			"data": {
				Type:        schema.TypeMap,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Key-value pairs of secret data",
				Sensitive:   true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the secret was created",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the secret was last updated",
			},
		},
	}
}

// buildSecretPayload converts Terraform state to API payload.
func buildSecretPayload(d *schema.ResourceData) SecretPayload {
	payload := SecretPayload{
		Name: d.Get("name").(string),
		Data: make(map[string]string),
	}

	if desc, ok := d.Get("description").(string); ok && desc != "" {
		payload.Description = desc
	}

	// Convert the map[string]interface{} to map[string]string
	if dataMap, ok := d.Get("data").(map[string]interface{}); ok {
		for k, v := range dataMap {
			if strVal, ok := v.(string); ok {
				payload.Data[k] = strVal
			}
		}
	}

	return payload
}

// resourceSecretCreate calls POST /secrets/api/v1/secrets.
func resourceSecretCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, ok := m.(*apiClient)
	if !ok || client == nil {
		return diag.Errorf("invalid API client configuration")
	}

	payload := buildSecretPayload(d)
	body, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// Use /secrets/api/v1/secrets endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/secrets/api/v1/secrets", client.BaseURL), bytes.NewReader(body))
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Set Authorization header
	authHeader := client.Token
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] != "Bearer " {
		authHeader = "Bearer " + authHeader
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// Set GetBody for retry support
	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}
	}

	resp, diags := doRequestWithRetryDiag(ctx, client, req, client.RetryConfig)
	if diags != nil && diags.HasError() {
		return diags
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return diag.Errorf("create secret failed: %s: %s", resp.Status, string(b))
	}

	// Read the created secret from response
	var secret SecretInfo
	if err := json.NewDecoder(resp.Body).Decode(&secret); err != nil {
		// If response doesn't contain the secret, try to fetch it by name
		log.Printf("[WARN] failed to decode create response, will fetch by name: %v", err)
		return resourceSecretRead(ctx, d, m)
	}

	// Set the ID from the response
	if secret.ID != "" {
		d.SetId(secret.ID)
	} else {
		// If no ID in response, use name as ID (fallback)
		d.SetId(payload.Name)
	}

	return resourceSecretRead(ctx, d, m)
}

// resourceSecretRead calls GET /secrets/api/v1/secrets/:id or GET /secrets/api/v1/secrets to find by name.
func resourceSecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, ok := m.(*apiClient)
	if !ok || client == nil {
		return diag.Errorf("invalid API client configuration")
	}

	resourceID := d.Id()
	name := d.Get("name").(string)

	// Try to fetch by ID first
	var secret *SecretInfo
	var err error

	if resourceID != "" && resourceID != name {
		// Try GET /secrets/api/v1/secrets/:id
		secret, err = fetchSecretByID(ctx, client, resourceID)
		if err != nil {
			log.Printf("[WARN] failed to fetch secret by ID %s: %v", resourceID, err)
		}
	}

	// If not found by ID, try to find by name
	if secret == nil {
		secret, err = fetchSecretByName(ctx, client, name)
		if err != nil {
			log.Printf("[WARN] failed to fetch secret by name %s: %v", name, err)
		}
	}

	if secret == nil {
		// Secret not found; mark resource as gone.
		d.SetId("")
		return nil
	}

	// Update state with the secret data
	_ = d.Set("name", secret.Name)
	_ = d.Set("description", secret.Description)
	_ = d.Set("data", secret.Data)
	_ = d.Set("created_at", secret.CreatedAt)
	_ = d.Set("updated_at", secret.UpdatedAt)

	// Ensure ID is set
	if secret.ID != "" {
		d.SetId(secret.ID)
	} else if d.Id() == "" {
		// Fallback to name if no ID
		d.SetId(secret.Name)
	}

	return nil
}

// resourceSecretUpdate calls PUT /secrets/api/v1/secrets/:id.
func resourceSecretUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, ok := m.(*apiClient)
	if !ok || client == nil {
		return diag.Errorf("invalid API client configuration")
	}

	resourceID := d.Id()
	if resourceID == "" {
		return diag.Errorf("secret ID is required for update")
	}

	payload := buildSecretPayload(d)
	body, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// Use PUT /secrets/api/v1/secrets/:id endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/secrets/api/v1/secrets/%s", client.BaseURL, resourceID), bytes.NewReader(body))
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Set Authorization header
	authHeader := client.Token
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] != "Bearer " {
		authHeader = "Bearer " + authHeader
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// Set GetBody for retry support
	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}
	}

	resp, diags := doRequestWithRetryDiag(ctx, client, req, client.RetryConfig)
	if diags != nil && diags.HasError() {
		return diags
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return diag.Errorf("update secret failed: %s: %s", resp.Status, string(b))
	}

	return resourceSecretRead(ctx, d, m)
}

// resourceSecretDelete calls DELETE /secrets/api/v1/secrets/:id.
func resourceSecretDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, ok := m.(*apiClient)
	if !ok || client == nil {
		return diag.Errorf("invalid API client configuration")
	}

	resourceID := d.Id()
	name := d.Get("name").(string)

	// If no ID, try to find the secret by name using the list API
	if resourceID == "" || resourceID == name {
		if name != "" {
			log.Printf("[INFO] No ID found, looking up secret by name: %s", name)
			secret, err := fetchSecretByName(ctx, client, name)
			if err != nil {
				log.Printf("[WARN] failed to find secret by name %s: %v", name, err)
			} else if secret != nil && secret.ID != "" {
				resourceID = secret.ID
				log.Printf("[INFO] Found secret ID: %s for name: %s", resourceID, name)
			}
		}
	}

	if resourceID == "" {
		log.Printf("[WARN] Cannot delete secret: no ID available and name lookup failed")
		d.SetId("")
		return nil
	}

	// Use DELETE /secrets/api/v1/secrets/:id endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/secrets/api/v1/secrets/%s", client.BaseURL, resourceID), nil)
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Set("Accept", "application/json")

	// Set Authorization header
	authHeader := client.Token
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] != "Bearer " {
		authHeader = "Bearer " + authHeader
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, diags := doRequestWithRetryDiag(ctx, client, req, client.RetryConfig)
	if diags != nil && diags.HasError() {
		// Verify deletion by trying to read the secret
		log.Printf("[WARN] delete request returned error, verifying secret deletion...")
		time.Sleep(2 * time.Second)

		secret, checkErr := fetchSecretByID(ctx, client, resourceID)
		if checkErr != nil {
			log.Printf("[WARN] failed to verify secret deletion: %v", checkErr)
		}

		if secret == nil {
			log.Printf("[INFO] secret %s successfully deleted (verified)", resourceID)
			d.SetId("")
			return nil
		}

		return diags
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("[WARN] failed to read delete response body: %v", readErr)
	}

	// Accept 200-299 and 404 (already deleted) as success
	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[INFO] secret %s not found (already deleted)", resourceID)
		d.SetId("")
		return nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyStr := string(bodyBytes)
		if bodyStr == "" {
			bodyStr = "(no response body)"
		}
		// Verify deletion
		log.Printf("[WARN] delete returned status %s, verifying secret deletion...", resp.Status)
		time.Sleep(2 * time.Second)
		secret, checkErr := fetchSecretByID(ctx, client, resourceID)
		if checkErr == nil && secret == nil {
			log.Printf("[INFO] secret %s successfully deleted (verified despite error status)", resourceID)
			d.SetId("")
			return nil
		}
		return diag.Errorf("delete secret failed: %s: %s", resp.Status, bodyStr)
	}

	log.Printf("[INFO] successfully deleted secret %s", resourceID)
	d.SetId("")
	return nil
}

// fetchSecretByID queries GET /secrets/api/v1/secrets/:id and returns the secret.
func fetchSecretByID(ctx context.Context, client *apiClient, id string) (*SecretInfo, error) {
	u := fmt.Sprintf("%s/secrets/api/v1/secrets/%s", client.BaseURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	// Set Authorization header
	authHeader := client.Token
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] != "Bearer " {
		authHeader = "Bearer " + authHeader
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("secret fetch failed: %s: %s", resp.Status, string(b))
	}

	var secret SecretInfo
	if err := json.NewDecoder(resp.Body).Decode(&secret); err != nil {
		return nil, err
	}
	return &secret, nil
}

// fetchSecretByName queries GET /secrets/api/v1/secrets and finds the secret by name.
func fetchSecretByName(ctx context.Context, client *apiClient, name string) (*SecretInfo, error) {
	u := fmt.Sprintf("%s/secrets/api/v1/secrets", client.BaseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	// Set Authorization header
	authHeader := client.Token
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] != "Bearer " {
		authHeader = "Bearer " + authHeader
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("secrets list fetch failed: %s: %s", resp.Status, string(b))
	}

	var listResp SecretsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}

	// Find secret by name
	for _, secret := range listResp.Secrets {
		if secret.Name == name {
			return &secret, nil
		}
	}

	return nil, nil // Not found
}
