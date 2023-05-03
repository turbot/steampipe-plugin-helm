package helm

import (
	"context"
	"strings"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"k8s.io/helm/pkg/renderutil"
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
			// {Name: "path", Type: proto.ColumnType_STRING, Description: "Name is the path-like name of the template."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name is the path-like name of the template."},
			{Name: "rendered", Type: proto.ColumnType_STRING, Description: "Data is the template as byte data."},
			{Name: "raw", Type: proto.ColumnType_STRING, Description: "Data is the template as byte data."},
		},
	}
}

type helmTemplate struct {
	// Path string
	Name     string
	Rendered string
	Raw      string
}

//// LIST FUNCTION

func listHelmTemplates(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	charts, err := getParsedHelmChart(ctx, d)
	if err != nil {
		return nil, err
	}

	for _, chart := range charts {
		renderedChart, err := renderutil.Render(&chart.Chart, chart.Values, renderutil.Options{})
		if err != nil {
			return nil, err
		}

		for _, template := range chart.Templates {
			for k, v := range renderedChart {
				if strings.HasSuffix(k, template.Name) {
					d.StreamListItem(ctx, helmTemplate{k, v, string(template.Data)})
				}
			}
		}
	}

	return nil, nil
}
