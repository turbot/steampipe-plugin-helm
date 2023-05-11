package helm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
	"k8s.io/helm/pkg/chartutil"

	"helm.sh/helm/v3/pkg/engine"
)

//// TABLE DEFINITION

func tableHelmTemplate(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "helm_template",
		Description: "Templates defines in a specific chart directory",
		List: &plugin.ListConfig{
			Hydrate: listHelmTemplates,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "chart_name", Require: plugin.Required},
				{Name: "vals", Require: plugin.Optional, CacheMatch: "exact"},
				{Name: "val_file", Require: plugin.Optional, CacheMatch: "exact"},
			},
		},
		Columns: []*plugin.Column{
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name is the path-like name of the template."},
			{Name: "rendered", Type: proto.ColumnType_STRING, Description: "Data is the template as byte data."},
			{Name: "raw", Type: proto.ColumnType_STRING, Description: "Data is the template as byte data."},
			{Name: "vals", Type: proto.ColumnType_JSON, Description: "Values to override the default values.", Transform: transform.FromQual("vals")},
			{Name: "val_file", Type: proto.ColumnType_STRING, Description: "Values to override the default values.", Transform: transform.FromQual("val_file")},
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
	charts, err := getParsedHelmChart(ctx, d)
	if err != nil {
		return nil, err
	}
	chartName := d.EqualsQualString("chart_name")

	// Check for values from quals to override the default value
	additionalValues := map[string]interface{}{}
	if d.EqualsQuals["vals"] != nil {
		inputVals := d.EqualsQuals["vals"].GetJsonbValue()
		err = json.Unmarshal([]byte(inputVals), &additionalValues)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal input: %w", err)
		}
	}

	// Check for values from file to override the default value
	if d.EqualsQuals["val_file"] != nil {
		valFilePath := d.EqualsQuals["val_file"].GetStringValue()

		valuesBytes, err := os.ReadFile(valFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %s: %v", valFilePath, err)
		}

		values, err := chartutil.ReadValues(valuesBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to read values from file: %s: %v", valFilePath, err)
		}
		additionalValues = mergeMaps(additionalValues, values.AsMap())
	}

	for _, c := range charts {
		if c.Chart.Metadata.Name == chartName {

			values := c.Chart.Values
			values = mergeMaps(values, additionalValues)
			values = map[string]interface{}{
				"Values": values,
				"Release": map[string]interface{}{
					"Service": "Helm",
					"Name":    chartName, // Keeping it as same as the chart name for now. In CLI, either the value can be passed in the arg, or can be auto-generated.
				},
				"Chart": map[string]interface{}{
					"Name":    chartName,
					"Version": c.Chart.Metadata.Version,
				},
			}

			renderedChart, err := engine.Render(c.Chart, values)
			if err != nil {
				return nil, err
			}

			for _, template := range c.Chart.Templates {
				for k, v := range renderedChart {
					if strings.HasSuffix(k, template.Name) {
						d.StreamListItem(ctx, helmTemplate{
							ChartName: chartName,
							Name:      k,
							Rendered:  v,
							Raw:       string(template.Data),
						})
					}
				}
			}
		}
	}

	return nil, nil
}
