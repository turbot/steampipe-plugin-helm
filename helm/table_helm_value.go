package helm

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

//// TABLE DEFINITION

func tableHelmValue(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "helm_value",
		Description: "",
		List: &plugin.ListConfig{
			Hydrate: listHelmValues,
		},
		Columns: []*plugin.Column{
			{Name: "path", Type: proto.ColumnType_STRING, Description: "Name is the path-like name of the template."},
			{Name: "raw", Type: proto.ColumnType_STRING, Description: "Name is the path-like name of the template."},
		},
	}
}

type HelmValue struct {
	Raw  string
	Path string
}

//// LIST FUNCTION

func listHelmValues(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	charts, err := getParsedHelmChart(ctx, d)
	if err != nil {
		return nil, err
	}

	for _, chart := range charts {
		d.StreamListItem(ctx, HelmValue{
			Raw:  chart.Values.Raw,
			Path: chart.Path,
		})
	}

	return nil, nil
}
