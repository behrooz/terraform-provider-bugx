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
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ClusterPayload represents the JSON body sent to /createcluster.
type ClusterPayload struct {
	Name            string `json:"Name"`
	ClusterID       string `json:"ClusterID"`
	ControlPlane    string `json:"ControlPlane"`
	Status          string `json:"Status"`
	Cpu             string `json:"Cpu"`
	Memory          string `json:"Memory"`
	PlatformVersion string `json:"PlatformVersion"`
	HealthCheck     string `json:"HealthCheck"`
	Alert           string `json:"Alert"`
	EndPoint        string `json:"EndPoint"`
	ClusterType     string `json:"ClusterType"`
	CoreDNSCpu      string `json:"CoreDNSCpu"`
	CoreDNSMemory   string `json:"CoreDNSMemory"`
	ApiServerCpu    string `json:"ApiServerCpu"`
	ApiServerMemory string `json:"ApiServerMemory"`
}

// ClusterInfo represents the JSON structure returned from /clusters.
type ClusterInfo struct {
	Name        string `json:"Name"`
	ClusterID   string `json:"ClusterID"`
	Status      string `json:"Status"`
	Version     string `json:"Version"`
	HealthCheck string `json:"HealthCheck"`
	Alert       string `json:"Alert"`
	EndPoint    string `json:"EndPoint"`
	NameSpace   string `json:"NameSpace"`
}

// resourceCluster defines the vcluster_cluster resource schema and CRUD.
func resourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name":             {Type: schema.TypeString, Required: true},
			"cluster_id":       {Type: schema.TypeString, Required: true},
			"control_plane":    {Type: schema.TypeString, Required: true},
			"status":           {Type: schema.TypeString, Optional: true, Default: "Progressing"},
			"cpu":              {Type: schema.TypeString, Required: true},
			"memory":           {Type: schema.TypeString, Required: true},
			"platform_version": {Type: schema.TypeString, Required: true},
			"health_check":     {Type: schema.TypeString, Optional: true},
			"alert":            {Type: schema.TypeString, Optional: true},
			"endpoint":         {Type: schema.TypeString, Optional: true, Computed: true},
			"namespace":        {Type: schema.TypeString, Optional: true, Computed: true},
			"kubeconfig":       {Type: schema.TypeString, Optional: true, Computed: true, Sensitive: true},
			"cluster_type":     {Type: schema.TypeString, Required: true},
			"coredns_cpu":      {Type: schema.TypeString, Required: true},
			"coredns_memory":   {Type: schema.TypeString, Required: true},
			"apiserver_cpu":    {Type: schema.TypeString, Required: true},
			"apiserver_memory": {Type: schema.TypeString, Required: true},
		},
	}
}

// buildPayload converts Terraform state to API payload.
func buildPayload(d *schema.ResourceData) ClusterPayload {
	return ClusterPayload{
		Name:            d.Get("name").(string),
		ClusterID:       d.Get("cluster_id").(string),
		ControlPlane:    d.Get("control_plane").(string),
		Status:          d.Get("status").(string),
		Cpu:             d.Get("cpu").(string),
		Memory:          d.Get("memory").(string),
		PlatformVersion: d.Get("platform_version").(string),
		HealthCheck:     d.Get("health_check").(string),
		Alert:           d.Get("alert").(string),
		EndPoint:        d.Get("endpoint").(string),
		ClusterType:     d.Get("cluster_type").(string),
		CoreDNSCpu:      d.Get("coredns_cpu").(string),
		CoreDNSMemory:   d.Get("coredns_memory").(string),
		ApiServerCpu:    d.Get("apiserver_cpu").(string),
		ApiServerMemory: d.Get("apiserver_memory").(string),
	}
}

// resourceClusterCreate calls POST /createcluster.
func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, ok := m.(*apiClient)
	if !ok || client == nil {
		return diag.Errorf("invalid API client configuration")
	}

	payload := buildPayload(d)
	body, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/createcluster", client.BaseURL), bytes.NewReader(body))
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Set Authorization header with raw token as provided by the login API usage.
	req.Header.Set("Authorization", client.Token)
	
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
		return diag.Errorf("createcluster failed: %s: %s", resp.Status, string(b))
	}

	// After creating the cluster, poll /clusters?Name=<name> until the Status becomes Healthy.
	name := payload.Name
	const (
		maxAttempts  = 60
		pollInterval = 10 * time.Second
	)

	var lastStatus string
	for i := 0; i < maxAttempts; i++ {
		info, err := fetchClusterInfo(ctx, client, name)
		if err != nil {
			log.Printf("[WARN] failed to fetch cluster %s status: %v", name, err)
		} else if info != nil {
			lastStatus = info.Status
			log.Printf("[INFO] cluster %s status: %s", name, info.Status)

			// Update a few fields in state from the latest info.
			_ = d.Set("status", info.Status)
			_ = d.Set("endpoint", info.EndPoint)
			_ = d.Set("namespace", info.NameSpace)
			if info.ClusterID != "" {
				_ = d.Set("cluster_id", info.ClusterID)
			}

			if info.Status == "Healthy" {
				// Fetch kubeconfig when cluster is Healthy
				kubeconfig, err := fetchKubeconfig(ctx, client, name)
				if err != nil {
					log.Printf("[WARN] failed to fetch kubeconfig for cluster %s: %v", name, err)
				} else if kubeconfig != "" {
					_ = d.Set("kubeconfig", kubeconfig)
				}

				// Call /clusters (without query) to get the namespace
				allClusters, err := fetchAllClusters(ctx, client)
				if err != nil {
					log.Printf("[WARN] failed to fetch all clusters to get namespace: %v", err)
				} else {
					// Find the cluster by name in the list
					for _, cluster := range allClusters {
						if cluster.Name == name && cluster.NameSpace != "" {
							_ = d.Set("namespace", cluster.NameSpace)
							log.Printf("[INFO] set cluster namespace to %s", cluster.NameSpace)
							break
						}
					}
				}

				// Use ClusterID as Terraform resource ID (from payload or info).
				if info.ClusterID != "" {
					d.SetId(info.ClusterID)
				} else {
					d.SetId(payload.ClusterID)
				}
				return resourceClusterRead(ctx, d, m)
			}
		}

		if i < maxAttempts-1 {
			select {
			case <-ctx.Done():
				return diag.FromErr(ctx.Err())
			case <-time.After(pollInterval):
			}
		}
	}

	return diag.Errorf("cluster %s did not become Healthy within the timeout; last known status: %s", name, lastStatus)
}

// resourceClusterRead reads cluster information from the API
func resourceClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, ok := m.(*apiClient)
	if !ok || client == nil {
		return diag.Errorf("invalid API client configuration")
	}

	// When importing, the ID is the cluster ID, so we need to find the cluster by ID
	// For now, we'll use the name field, but if importing, we might need to search by ID
	name := d.Get("name").(string)
	resourceID := d.Id()
	
	// If we have an ID but no name (e.g., from import), try to find cluster by ID
	if name == "" && resourceID != "" {
		// Try to fetch all clusters and find by ID
		allClusters, err := fetchAllClusters(ctx, client)
		if err == nil {
			for _, cluster := range allClusters {
				if cluster.ClusterID == resourceID {
					name = cluster.Name
					break
				}
			}
		}
	}
	
	if name == "" {
		// If we still don't have a name, mark as gone
		d.SetId("")
		return nil
	}

	info, err := fetchClusterInfo(ctx, client, name)
	if err != nil {
		log.Printf("[WARN] failed to read cluster %s: %v", name, err)
		return diag.FromErr(err)
	}
	if info == nil {
		// Cluster not found; mark resource as gone.
		d.SetId("")
		return nil
	}

	_ = d.Set("status", info.Status)
	_ = d.Set("endpoint", info.EndPoint)
	_ = d.Set("namespace", info.NameSpace)
	if info.ClusterID != "" {
		_ = d.Set("cluster_id", info.ClusterID)
	}

	// Fetch kubeconfig if cluster is Healthy
	if info.Status == "Healthy" {
		kubeconfig, err := fetchKubeconfig(ctx, client, name)
		if err != nil {
			log.Printf("[WARN] failed to fetch kubeconfig for cluster %s: %v", name, err)
		} else if kubeconfig != "" {
			_ = d.Set("kubeconfig", kubeconfig)
		}
	}

	return nil
}

// resourceClusterUpdate is a stub; you can extend it to call an update endpoint.
func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// TODO: Implement update behavior when API supports it.
	return resourceClusterRead(ctx, d, m)
}

// resourceClusterDelete calls DELETE /deletecluster?Name=<name>&Namespace=<namespace>.
func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, ok := m.(*apiClient)
	if !ok || client == nil {
		return diag.Errorf("invalid API client configuration")
	}

	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)

	if name == "" {
		// If we don't have a name, just clear the state
		d.SetId("")
		return nil
	}

	if namespace == "" {
		// Try to fetch the namespace from the API if we don't have it stored
		info, err := fetchClusterInfo(ctx, client, name)
		if err != nil {
			log.Printf("[WARN] failed to fetch cluster %s info for delete: %v", name, err)
		} else if info != nil && info.NameSpace != "" {
			namespace = info.NameSpace
		}
	}

	if namespace == "" {
		// If we still don't have namespace, proceed with delete anyway (API might handle it)
		log.Printf("[WARN] deleting cluster %s without namespace", name)
	}

	// Build the delete URL with query parameters
	u := fmt.Sprintf("%s/deletecluster?Name=%s", client.BaseURL, url.QueryEscape(name))
	if namespace != "" {
		u += fmt.Sprintf("&Namespace=%s", url.QueryEscape(namespace))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Set("Accept", "application/json")
	if client.Token != "" {
		req.Header.Set("Authorization", client.Token)
	}

	resp, diags := doRequestWithRetryDiag(ctx, client, req, client.RetryConfig)
	if diags != nil && diags.HasError() {
		// If we get EOF or connection error, verify the cluster is actually deleted
		// Some APIs close the connection immediately after processing the delete
		log.Printf("[WARN] delete request returned error, verifying cluster deletion...")

		// Wait a moment for the deletion to complete
		time.Sleep(2 * time.Second)

		// Check if cluster still exists
		info, checkErr := fetchClusterInfo(ctx, client, name)
		if checkErr != nil {
			log.Printf("[WARN] failed to verify cluster deletion: %v", checkErr)
		}

		if info == nil {
			// Cluster is gone, deletion was successful despite the connection error
			log.Printf("[INFO] cluster %s successfully deleted (verified)", name)
			d.SetId("")
			return nil
		}

		// Cluster still exists, return the original error
		return diags
	}
	defer resp.Body.Close()

	// Always read the response body to allow connection reuse
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("[WARN] failed to read delete response body: %v", readErr)
	}

	// Accept 200-299 and 404 (already deleted) as success
	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[INFO] cluster %s not found (already deleted)", name)
		d.SetId("")
		return nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyStr := string(bodyBytes)
		if bodyStr == "" {
			bodyStr = "(no response body)"
		}
		// Even if status code indicates error, verify the cluster is actually gone
		log.Printf("[WARN] delete returned status %s, verifying cluster deletion...", resp.Status)
		time.Sleep(2 * time.Second)
		info, checkErr := fetchClusterInfo(ctx, client, name)
		if checkErr == nil && info == nil {
			// Cluster is gone, deletion was successful
			log.Printf("[INFO] cluster %s successfully deleted (verified despite error status)", name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("deletecluster failed: %s: %s", resp.Status, bodyStr)
	}

	log.Printf("[INFO] successfully deleted cluster %s (namespace: %s)", name, namespace)
	d.SetId("")
	return nil
}

// fetchAllClusters queries /clusters (without query parameter) and returns all clusters.
func fetchAllClusters(ctx context.Context, client *apiClient) ([]ClusterInfo, error) {
	u := fmt.Sprintf("%s/clusters", client.BaseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "*/*")
	// Check if token already includes "Bearer " prefix, if not add it
	authHeader := client.Token
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] != "Bearer " {
		authHeader = "Bearer " + authHeader
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("clusters fetch failed: %s: %s", resp.Status, string(b))
	}

	var list []ClusterInfo
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}
	return list, nil
}

// fetchClusterInfo queries /clusters?Name=<name> and returns the first matching cluster info.
func fetchClusterInfo(ctx context.Context, client *apiClient, name string) (*ClusterInfo, error) {
	u := fmt.Sprintf("%s/clusters?Name=%s", client.BaseURL, url.QueryEscape(name))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	// Check if token already includes "Bearer " prefix, if not add it
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
		return nil, fmt.Errorf("clusters fetch failed: %s: %s", resp.Status, string(b))
	}

	var list []ClusterInfo
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	return &list[0], nil
}

// fetchKubeconfig queries /connect?Name=<name> and returns the kubeconfig content.
func fetchKubeconfig(ctx context.Context, client *apiClient, name string) (string, error) {
	u := fmt.Sprintf("%s/connect?Name=%s", client.BaseURL, url.QueryEscape(name))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "*/*")
	if client.Token != "" {
		req.Header.Set("Authorization", client.Token)
	}

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("kubeconfig fetch failed: %s: %s", resp.Status, string(b))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read kubeconfig response: %w", err)
	}

	return string(body), nil
}
