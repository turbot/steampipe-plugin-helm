package helm

import (
	"context"
	"os"
	"strings"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"gopkg.in/yaml.v3"

	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

//// TABLE DEFINITION

func tableHelmTemplate(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "helm_template",
		Description: "Templates defines in a specific chart directory",
		List: &plugin.ListConfig{
			Hydrate: listHelmTemplates,
		},
		Columns: []*plugin.Column{
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name is the path-like name of the template."},
			{Name: "rendered", Type: proto.ColumnType_STRING, Description: "Data is the template as byte data."},
			{Name: "raw", Type: proto.ColumnType_STRING, Description: "Data is the template as byte data."},
			{Name: "chart_name", Type: proto.ColumnType_STRING, Description: "The name of the chart."},
		},
	}
}

type helmTemplate struct {
	// Path string
	ChartName string
	Name      string
	Rendered  string
	Raw       string
}

//// LIST FUNCTION

func listHelmTemplates(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	chart, err := getParsedHelmChart(ctx, d)
	if err != nil {
		return nil, err
	}
	helmConfig := GetConfig(d.Connection)

	// Check for values from quals to override the default value
	values := chart.Chart.Values

	var valueFiles []string
	if helmConfig.ValueOverride != nil {
		valueFiles = append(valueFiles, helmConfig.ValueOverride...)
	}

	for _, f := range valueFiles {
		var override map[string]interface{}
		bs, err := os.ReadFile(f)
		if err != nil {
			plugin.Logger(ctx).Error("listHelmTemplates", "read_file_error", "connection_name", d.Connection.Name, "failed to read file %s: %v", f, err)
			return nil, err
		}
		if err := yaml.Unmarshal(bs, &values); err != nil {
			plugin.Logger(ctx).Error("listHelmTemplates", "unmarshal_error", "connection_name", d.Connection.Name, "failed to unmarshal file content: %v", err)
			return nil, err
		}

		for k, v := range override {
			values[k] = v
		}
	}

	// values = mergeMaps(values, additionalValues)
	values = map[string]interface{}{
		"Values": values,
		"Release": map[string]interface{}{
			"Service": "Helm",
			"Name":    chart.Chart.Metadata.Name, // Keeping it as same as the chart name for now. In CLI, either the value can be passed in the arg, or can be auto-generated.
		},
		"Chart":        chart.Chart.Metadata,
		"Capabilities": chartutil.Capabilities{},
		"Template": map[string]interface{}{
			"BasePath": "/path/to/base",
		},
	}

	renderedChart, err := engine.Render(chart.Chart, values)
	if err != nil {
		return nil, err
	}

	for _, template := range chart.Chart.Templates {
		for k, v := range renderedChart {
			if strings.HasSuffix(k, template.Name) {
				d.StreamListItem(ctx, helmTemplate{
					ChartName: chart.Chart.Metadata.Name,
					Name:      k,
					Rendered:  v,
					Raw:       string(template.Data),
				})
			}
		}
	}

	return nil, nil
}
