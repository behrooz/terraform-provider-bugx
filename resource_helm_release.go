package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// HelmInstallPayload represents the JSON body sent to /helm_install.
type HelmInstallPayload struct {
	Clustername string `json:"Clustername"`
	Namespace   string `json:"Namespace"`
	Release     string `json:"Release"`
	Chart       string `json:"Chart"`
	Repo        string `json:"Repo"`
	Version     string `json:"Version,omitempty"` // Optional: Chart version
	Values      string `json:"Values,omitempty"`   // Optional: Helm values as YAML string
}

// resourceHelmRelease defines the vcluster_helm_release resource schema and CRUD.
func resourceHelmRelease() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHelmReleaseCreate,
		ReadContext:   resourceHelmReleaseRead,
		UpdateContext: resourceHelmReleaseUpdate,
		DeleteContext: resourceHelmReleaseDelete,

		Schema: map[string]*schema.Schema{
			"cluster_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the vcluster where to deploy the Helm release",
			},
			"namespace": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Kubernetes namespace where to deploy the release",
			},
			"release": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Helm release",
			},
			"chart": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Chart name (e.g., 'bitnami/mysql' or 'mysql')",
			},
			"repo": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Helm repository URL (e.g., 'https://charts.bitnami.com/bitnami')",
			},
			"values": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Helm values as YAML string. You can use file() or templatefile() to load from a file",
			},
			"values_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Path to a Helm values YAML file. Alternative to 'values' attribute",
			},
			"chart_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Version of the Helm chart to install (e.g., '8.0.0'). If not specified, the latest version is used",
			},
		},
	}
}

// buildHelmPayload converts Terraform state to API payload.
func buildHelmPayload(d *schema.ResourceData) (*HelmInstallPayload, error) {
	payload := &HelmInstallPayload{
		Clustername: d.Get("cluster_name").(string),
		Namespace:   d.Get("namespace").(string),
		Release:     d.Get("release").(string),
		Chart:       d.Get("chart").(string),
		Repo:        d.Get("repo").(string),
	}

	// Handle chart version if provided
	if chartVersion, ok := d.Get("chart_version").(string); ok && chartVersion != "" {
		payload.Version = chartVersion
	}

	// Handle values - prefer values_file if both are provided
	valuesFile := d.Get("values_file").(string)
	values := d.Get("values").(string)

	if valuesFile != "" {
		// Read values from file
		fileContent, err := os.ReadFile(valuesFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read values file %s: %w", valuesFile, err)
		}
		payload.Values = string(fileContent)
	} else if values != "" {
		// Use inline values
		payload.Values = values
	}

	return payload, nil
}

// resourceHelmReleaseCreate calls POST /helm_install.
func resourceHelmReleaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, ok := m.(*apiClient)
	if !ok || client == nil {
		return diag.Errorf("invalid API client configuration")
	}

	payload, err := buildHelmPayload(d)
	if err != nil {
		return diag.FromErr(err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/helm_install", client.BaseURL), bytes.NewReader(body))
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Check if token already includes "Bearer " prefix, if not add it
	authHeader := client.Token
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] != "Bearer " {
		authHeader = "Bearer " + authHeader
	}
	req.Header.Set("Authorization", authHeader)
	
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

	// Always read the response body
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("[WARN] failed to read helm_install response body: %v", readErr)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyStr := string(bodyBytes)
		if bodyStr == "" {
			bodyStr = "(no response body)"
		}
		return diag.Errorf("helm_install failed: %s: %s", resp.Status, bodyStr)
	}

	// Use a composite ID: cluster_name:namespace:release
	resourceID := fmt.Sprintf("%s:%s:%s", payload.Clustername, payload.Namespace, payload.Release)
	d.SetId(resourceID)

	log.Printf("[INFO] successfully installed Helm release %s in cluster %s", payload.Release, payload.Clustername)
	return resourceHelmReleaseRead(ctx, d, m)
}

// resourceHelmReleaseRead is a stub - you can extend this if your API supports reading Helm releases.
func resourceHelmReleaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// TODO: Implement read if your API supports GET /helm_releases or similar
	// For now, we assume the release exists if the resource is in state
	return nil
}

// resourceHelmReleaseUpdate handles updates by reinstalling with new values.
func resourceHelmReleaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// If any of the immutable fields changed, we need to recreate
	if d.HasChanges("cluster_name", "namespace", "release", "chart", "repo", "chart_version") {
		// These changes require recreation
		return diag.Errorf("cannot change cluster_name, namespace, release, chart, repo, or chart_version. These require recreation")
	}

	// If only values changed, reinstall with new values
	if d.HasChanges("values", "values_file") {
		return resourceHelmReleaseCreate(ctx, d, m)
	}

	return resourceHelmReleaseRead(ctx, d, m)
}

// resourceHelmReleaseDelete calls DELETE /deleteapp?Name=<namespace><release> to delete the app.
func resourceHelmReleaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, ok := m.(*apiClient)
	if !ok || client == nil {
		return diag.Errorf("invalid API client configuration")
	}

	// Parse the resource ID to get cluster, namespace, and release
	parts := splitResourceID(d.Id())
	if len(parts) != 3 {
		log.Printf("[WARN] invalid resource ID format, clearing state: %s", d.Id())
		d.SetId("")
		return nil
	}

	clustername := parts[0]
	release := parts[2] // parts[1] is kubernetes namespace, not cluster namespace

	// Get cluster namespace by fetching cluster info
	var appName string
	clusterInfo, err := fetchClusterInfo(ctx, client, clustername)
	if err != nil {
		log.Printf("[WARN] failed to fetch cluster %s info to get namespace: %v", clustername, err)
		// Try to use release name directly if we can't get cluster namespace
		appName = release
		log.Printf("[WARN] falling back to using release name %s directly", appName)
	} else if clusterInfo == nil || clusterInfo.NameSpace == "" {
		log.Printf("[WARN] cluster %s not found or namespace is empty, using release name directly", clustername)
		appName = release
	} else {
		// Use {cluster_namespace}-{release} as the app name
		appName = clusterInfo.NameSpace + "-" + release
		log.Printf("[DEBUG] Using app name %s (namespace: %s + release: %s)", appName, clusterInfo.NameSpace, release)
	}

	// Build the delete URL with query parameter Name=<appName>
	// API endpoint: DELETE /deleteapp?Name=<namespace><release>
	deleteURL := fmt.Sprintf("%s/deleteapp?Name=%s", client.BaseURL, url.QueryEscape(appName))
	log.Printf("[DEBUG] Deleting app %s from cluster %s via %s", appName, clustername, deleteURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL, nil)
	if err != nil {
		return diag.Errorf("failed to create delete request: %v", err)
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")
	// Check if token already includes "Bearer " prefix, if not add it
	authHeader := client.Token
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] != "Bearer " {
		authHeader = "Bearer " + authHeader
	}
	req.Header.Set("Authorization", authHeader)

	resp, diags := doRequestWithRetryDiag(ctx, client, req, client.RetryConfig)
	if diags != nil && diags.HasError() {
		return diags
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("[WARN] failed to read deleteapp response body: %v", readErr)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[INFO] App %s not found (already deleted)", release)
		d.SetId("")
		return nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyStr := string(bodyBytes)
		if bodyStr == "" {
			bodyStr = "(no response body)"
		}
		return diag.Errorf("deleteapp failed for %s: %s: %s", release, resp.Status, bodyStr)
	}

	log.Printf("[INFO] successfully deleted app %s from cluster %s", release, clustername)
	d.SetId("")
	return nil
}

// splitResourceID splits the composite ID into its components.
func splitResourceID(id string) []string {
	// ID format: cluster_name:namespace:release
	parts := make([]string, 0, 3)
	lastIndex := 0
	for i, r := range id {
		if r == ':' {
			if i > lastIndex {
				parts = append(parts, id[lastIndex:i])
			}
			lastIndex = i + 1
		}
	}
	if lastIndex < len(id) {
		parts = append(parts, id[lastIndex:])
	}
	return parts
}
