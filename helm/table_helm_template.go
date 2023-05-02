package helm

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableHelmTemplate(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "helm_template",
		Description: "",
		List: &plugin.ListConfig{
			Hydrate: listHelmTemplates,
		},
		Columns: []*plugin.Column{
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name is the path-like name of the template."},
			{Name: "data", Type: proto.ColumnType_STRING, Description: "Data is the template as byte data.", Transform: transform.FromField("Data").Transform(transform.ToString)},
		},
	}
}

//// LIST FUNCTION

func listHelmTemplates(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	charts, err := getParsedHelmChart(ctx, d)
	if err != nil {
		return nil, err
	}

	for _, chart := range charts {
		for _, template := range chart.Templates {
			d.StreamListItem(ctx, template)
		}
	}

	return nil, nil
}
