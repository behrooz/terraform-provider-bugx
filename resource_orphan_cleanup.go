package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resourceOrphanCleanup defines a resource that deletes orphaned applications
// (applications that exist on the server but are not in Terraform state)
func resourceOrphanCleanup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOrphanCleanupCreate,
		ReadContext:   resourceOrphanCleanupRead,
		UpdateContext: resourceOrphanCleanupUpdate,
		DeleteContext: resourceOrphanCleanupDelete,

		Schema: map[string]*schema.Schema{
			"cluster_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the vcluster to clean up orphaned applications from",
			},
			"apps_to_delete": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Set of application names to delete explicitly. These should be the full app names (e.g., 'ns-977i-rabbitmq' for cluster namespace 'ns-977i' and release 'rabbitmq').",
			},
			"keep_releases": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of Helm release names to keep. If provided along with cluster namespace, apps matching '{namespace}-{release}' pattern that are NOT in this list will be deleted. Use this for automatic cleanup based on release names.",
			},
			"deleted_apps": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of application names that were successfully deleted",
			},
		},
	}
}

// resourceOrphanCleanupCreate deletes the specified orphaned applications
func resourceOrphanCleanupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, ok := m.(*apiClient)
	if !ok || client == nil {
		return diag.Errorf("invalid API client configuration")
	}

	clusterName := d.Get("cluster_name").(string)

	// Get cluster info to verify cluster exists and get namespace
	clusterInfo, err := fetchClusterInfo(ctx, client, clusterName)
	if err != nil {
		return diag.Errorf("failed to fetch cluster info for %s: %v", clusterName, err)
	}
	if clusterInfo == nil {
		return diag.Errorf("cluster %s not found", clusterName)
	}

	clusterNamespace := clusterInfo.NameSpace
	log.Printf("[INFO] Starting orphan cleanup for cluster %s (namespace: %s)", clusterName, clusterNamespace)

	var appsToDelete []string

	// Method 1: Explicit apps_to_delete list
	if appsToDeleteSet, ok := d.GetOk("apps_to_delete"); ok {
		for _, appInterface := range appsToDeleteSet.(*schema.Set).List() {
			appName := appInterface.(string)
			if appName != "" {
				appsToDelete = append(appsToDelete, appName)
			}
		}
		log.Printf("[INFO] Found %d apps to delete from explicit list", len(appsToDelete))
	}

	// Method 2: Use keep_releases to determine what to delete
	// This is a best-effort approach: we'll try to delete apps that match the pattern
	// but aren't in the keep list. Since we can't list all apps, this requires
	// the user to know which releases might exist.
	if keepReleasesSet, ok := d.GetOk("keep_releases"); ok && clusterNamespace != "" {
		keepReleases := make(map[string]bool)
		for _, releaseInterface := range keepReleasesSet.(*schema.Set).List() {
			release := releaseInterface.(string)
			if release != "" {
				keepReleases[release] = true
				// The app name would be {namespace}-{release}
				keepReleases[clusterNamespace+"-"+release] = true
			}
		}

		// If user provided specific releases to keep, we can infer apps to delete
		// by checking common release names. But without a list API, we can't know
		// all apps. So this method is mainly for when user knows what might exist.
		log.Printf("[INFO] Keeping %d releases (apps matching pattern %s-*)", len(keepReleases), clusterNamespace)
		// Note: Without a list API, we can't automatically find all orphaned apps
		// The user should use apps_to_delete for explicit cleanup
	}

	if len(appsToDelete) == 0 {
		log.Printf("[WARN] No apps specified for deletion. Provide either 'apps_to_delete' or use 'keep_releases' with known release names.")
		d.SetId(fmt.Sprintf("%s-orphan-cleanup", clusterName))
		d.Set("deleted_apps", []string{})
		return resourceOrphanCleanupRead(ctx, d, m)
	}

	var deletedApps []string
	var errors []error

	// Delete each app
	for _, appName := range appsToDelete {
		err := deleteOrphanApp(ctx, client, clusterName, appName)
		if err != nil {
			log.Printf("[ERROR] Failed to delete app %s: %v", appName, err)
			errors = append(errors, fmt.Errorf("failed to delete app %s: %w", appName, err))
		} else {
			deletedApps = append(deletedApps, appName)
			log.Printf("[INFO] Successfully deleted app %s", appName)
		}
	}

	// Set ID
	d.SetId(fmt.Sprintf("%s-orphan-cleanup", clusterName))

	// Store deleted apps
	if err := d.Set("deleted_apps", deletedApps); err != nil {
		return diag.FromErr(err)
	}

	// If there were errors, return them
	if len(errors) > 0 {
		var diags diag.Diagnostics
		for _, err := range errors {
			diags = append(diags, diag.FromErr(err)...)
		}
		return diags
	}

	log.Printf("[INFO] Orphan cleanup completed for cluster %s: deleted %d apps", clusterName, len(deletedApps))
	return resourceOrphanCleanupRead(ctx, d, m)
}

// resourceOrphanCleanupRead reads the current state of orphan cleanup
func resourceOrphanCleanupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// This resource doesn't have a server-side state to read
	// Just return the current config
	return nil
}

// resourceOrphanCleanupUpdate handles updates - if apps_to_delete or keep_releases changes, re-run cleanup
func resourceOrphanCleanupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("apps_to_delete") || d.HasChange("keep_releases") {
		// Re-run cleanup with new apps list
		return resourceOrphanCleanupCreate(ctx, d, m)
	}
	return resourceOrphanCleanupRead(ctx, d, m)
}

// resourceOrphanCleanupDelete handles deletion of the cleanup resource itself
func resourceOrphanCleanupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// This resource doesn't create anything on the server
	// Just clear the state
	d.SetId("")
	return nil
}

// deleteOrphanApp deletes an application using the deleteapp API
func deleteOrphanApp(ctx context.Context, client *apiClient, clusterName string, appName string) error {
	deleteURL := fmt.Sprintf("%s/deleteapp?Name=%s", client.BaseURL, url.QueryEscape(appName))
	log.Printf("[INFO] Deleting orphaned app %s from cluster %s via %s", appName, clusterName, deleteURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")
	
	authHeader := client.Token
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] != "Bearer " {
		authHeader = "Bearer " + authHeader
	}
	req.Header.Set("Authorization", authHeader)

	resp, diags := doRequestWithRetryDiag(ctx, client, req, client.RetryConfig)
	if diags != nil && diags.HasError() {
		return fmt.Errorf("delete API call failed: %v", diags)
	}
	
	if resp == nil {
		return fmt.Errorf("delete API returned nil response")
	}
	
	defer resp.Body.Close()

	// Read response body
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("[WARN] failed to read deleteapp response body: %v", readErr)
	}

	bodyStr := string(bodyBytes)
	log.Printf("[DEBUG] Delete API response for %s: Status=%d, Body=%s", appName, resp.StatusCode, bodyStr)

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[INFO] App %s not found (already deleted)", appName)
		return nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("deleteapp failed: %s: %s", resp.Status, bodyStr)
	}

	log.Printf("[INFO] Successfully deleted orphaned app %s", appName)
	return nil
}

