package helm

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"helm.sh/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

type parsedHelmChart struct {
	chart.Chart
	Path string
}

// Get the parsed contents of the given files.
func getParsedHelmChart(ctx context.Context, d *plugin.QueryData) ([]parsedHelmChart, error) {
	conn, err := parsedHelmChartCached(ctx, d, nil)
	if err != nil {
		return nil, err
	}
	return conn.([]parsedHelmChart), nil
}

// Cached form of the parsed file content.
var parsedHelmChartCached = plugin.HydrateFunc(parsedHelmChartUncached).Memoize()

// parsedHelmChartUncached is the actual implementation of getParsedHelmChart, which should
// be run only once per connection. Do not call this directly, use
// getParsedHelmChart instead.
func parsedHelmChartUncached(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (any, error) {
	plugin.Logger(ctx).Debug("parsedHelmChartUncached", "Parsing file content...", "connection", d.Connection.Name)

	// Read the config
	helmConfig := GetConfig(d.Connection)

	var charts []parsedHelmChart
	for _, path := range helmConfig.Paths {
		chart, err := chartutil.Load(path)
		if err != nil {
			return nil, err
		}
		charts = append(charts, parsedHelmChart{*chart, path})
	}

	return charts, nil
}
