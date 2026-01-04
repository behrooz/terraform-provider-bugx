package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceCluster defines a data source to query existing clusters
func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the bugx cluster to query",
			},
			"cluster_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Cluster ID",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current status of the cluster",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Cluster endpoint URL",
			},
			"namespace": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Kubernetes namespace where the cluster is deployed",
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Platform version of the cluster",
			},
			"kubeconfig": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Kubeconfig content for connecting to the cluster",
			},
		},
	}
}

// dataSourceClusterRead queries the API for cluster information
func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, ok := m.(*apiClient)
	if !ok || client == nil {
		return diag.Errorf("invalid API client configuration")
	}

	name := d.Get("name").(string)
	if name == "" {
		return diag.Errorf("cluster name is required")
	}

	// Fetch cluster info
	info, err := fetchClusterInfo(ctx, client, name)
	if err != nil {
		return diag.FromErr(err)
	}

	if info == nil {
		return diag.Errorf("cluster '%s' not found", name)
	}

	// Set the resource ID
	d.SetId(info.ClusterID)

	// Set computed attributes
	if err := d.Set("cluster_id", info.ClusterID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", info.Status); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("endpoint", info.EndPoint); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("namespace", info.NameSpace); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("version", info.Version); err != nil {
		return diag.FromErr(err)
	}

	// Fetch kubeconfig if cluster is healthy
	if info.Status == "Healthy" {
		kubeconfig, err := fetchKubeconfig(ctx, client, name)
		if err != nil {
			log.Printf("[WARN] failed to fetch kubeconfig for cluster %s: %v", name, err)
		} else if kubeconfig != "" {
			if err := d.Set("kubeconfig", kubeconfig); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return nil
}

