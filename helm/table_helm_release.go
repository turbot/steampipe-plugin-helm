package helm

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
	// "helm.sh/helm/v3/pkg/chart/loader"
)

//// TABLE DEFINITION

func tableHelmRelease(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "helm_release",
		Description: "",
		List: &plugin.ListConfig{
			Hydrate: listHelmReleases,
		},
		Columns: []*plugin.Column{
			{Name: "name", Type: proto.ColumnType_STRING, Description: "The name of the release."},
			{Name: "namespace", Type: proto.ColumnType_STRING, Description: "The kubernetes namespace of the release."},
			{Name: "version", Type: proto.ColumnType_INT, Description: "The revision of the release."},
			{Name: "status", Type: proto.ColumnType_STRING, Description: "The current state of the release.", Transform: transform.FromField("Info.Status").Transform(transform.ToString)},
			{Name: "description", Type: proto.ColumnType_STRING, Description: "A human-friendly description about the release.", Transform: transform.FromField("Info.Description")},
			// {Name: "first_deployed", Type: proto.ColumnType_STRING, Description: "The time when the release was first deployed.", Transform: transform.FromField("Info.FirstDeployed")},
			// {Name: "last_deployed", Type: proto.ColumnType_STRING, Description: "The time when the release was last deployed.", Transform: transform.FromField("Info.LastDeployed")},
			// {Name: "deleted", Type: proto.ColumnType_STRING, Description: "The time when this object was deleted.", Transform: transform.FromField("Info.Deleted")},
			{Name: "notes", Type: proto.ColumnType_STRING, Description: "Contains the rendered templates/NOTES.txt if available.", Transform: transform.FromField("Info.Notes")},
			{Name: "manifest", Type: proto.ColumnType_STRING, Description: "The string representation of the rendered template."},
			{Name: "config", Type: proto.ColumnType_JSON, Description: "The set of extra Values added to the chart. These values override the default values inside of the chart."},
			{Name: "labels", Type: proto.ColumnType_JSON, Description: "The labels of the release."},
		},
	}
}

//// LIST FUNCTION

func listHelmReleases(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := getHelmClient(ctx, d, nil)
	if err != nil {
		return nil, err
	}

	releases, err := client.ListDeployedReleases()
	if err != nil {
		return nil, err
	}

	for _, release := range releases {
		d.StreamListItem(ctx, release)
	}

	return nil, nil
}
