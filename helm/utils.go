package helm

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
	"gopkg.in/yaml.v3"

	helmClient "github.com/mittwald/go-helm-client"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

type parsedHelmChart struct {
	Chart *chart.Chart
	Path  string
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
		chart, err := loader.Load(path)
		if err != nil {
			return nil, err
		}
		charts = append(charts, parsedHelmChart{
			Chart: chart,
			Path:  path,
		})
	}

	return charts, nil
}

func getHelmClient(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (helmClient.Client, error) {
	client, err := helmClient.New(nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

type Rows []Row
type Row struct {
	Path        string
	Key         []string
	Value       interface{}
	Tag         *string
	PreComments []string
	HeadComment string
	LineComment string
	FootComment string
	StartLine   int
	StartColumn int
}

func treeToList(tree *yaml.Node, prefix []string, rows *Rows, preComments []string, headComments []string, footComments []string) {
	switch tree.Kind {
	case yaml.DocumentNode:
		for i, v := range tree.Content {
			localComments := []string{}
			headComments = []string{}
			footComments = []string{}
			if i == 0 {
				localComments = append(localComments, preComments...)
				if tree.HeadComment != "" {
					localComments = append(localComments, tree.HeadComment)
					headComments = append(headComments, tree.HeadComment)
				}
				if tree.FootComment != "" {
					footComments = append(footComments, tree.FootComment)
				}
				if tree.LineComment != "" {
					localComments = append(localComments, tree.LineComment)
				}
			}
			treeToList(v, prefix, rows, localComments, headComments, footComments)
		}
	case yaml.SequenceNode:
		if len(tree.Content) == 0 {
			row := Row{
				Key:         prefix,
				Value:       []string{},
				Tag:         &tree.Tag,
				StartLine:   tree.Line,
				StartColumn: tree.Column,
				PreComments: preComments,
				HeadComment: strings.Join(headComments, ","),
				LineComment: tree.LineComment,
				FootComment: strings.Join(footComments, ","),
			}
			*rows = append(*rows, row)
		}

		for i, v := range tree.Content {
			localComments := []string{}
			headComments = []string{}
			footComments = []string{}
			if i == 0 {
				localComments = append(localComments, preComments...)
				if tree.HeadComment != "" {
					localComments = append(localComments, tree.HeadComment)
					headComments = append(headComments, tree.HeadComment)
				}
				if tree.LineComment != "" {
					localComments = append(localComments, tree.LineComment)
				}
			}
			newKey := make([]string, len(prefix))
			copy(newKey, prefix)
			newKey = append(newKey, strconv.Itoa(i))
			treeToList(v, newKey, rows, localComments, headComments, footComments)
		}
	case yaml.MappingNode:
		localComments := []string{}
		headComments = []string{}
		footComments = []string{}
		localComments = append(localComments, preComments...)
		if tree.HeadComment != "" {
			localComments = append(localComments, tree.HeadComment)
			headComments = append(headComments, tree.HeadComment)
		}
		if tree.FootComment != "" {
			footComments = append(footComments, tree.FootComment)
		}
		if tree.LineComment != "" {
			localComments = append(localComments, tree.LineComment)
		}
		if len(tree.Content) == 0 {
			row := Row{
				Key:         prefix,
				Value:       map[string]interface{}{},
				Tag:         &tree.Tag,
				StartLine:   tree.Line,
				StartColumn: tree.Column,
				PreComments: preComments,
				HeadComment: strings.Join(headComments, ","),
				LineComment: tree.LineComment,
				FootComment: strings.Join(footComments, ","),
			}
			*rows = append(*rows, row)
		}
		i := 0
		for i < len(tree.Content)-1 {
			key := tree.Content[i]
			val := tree.Content[i+1]
			i = i + 2
			if key.HeadComment != "" {
				localComments = append(localComments, key.HeadComment)
				headComments = append(headComments, key.HeadComment)
			}
			if key.FootComment != "" {
				footComments = append(footComments, key.FootComment)
			}
			if key.LineComment != "" {
				localComments = append(localComments, key.LineComment)
			}
			newKey := make([]string, len(prefix))
			copy(newKey, prefix)
			newKey = append(newKey, key.Value)
			treeToList(val, newKey, rows, localComments, headComments, footComments)
			localComments = make([]string, 0)
			headComments = make([]string, 0)
			footComments = make([]string, 0)
		}
	case yaml.ScalarNode:
		row := Row{
			Key:         prefix,
			Value:       tree.Value,
			Tag:         &tree.Tag,
			StartLine:   tree.Line,
			StartColumn: tree.Column,
			PreComments: preComments,
			HeadComment: strings.Join(headComments, ","),
			LineComment: tree.LineComment,
			FootComment: strings.Join(footComments, ","),
		}
		if tree.Tag == "!!null" {
			row.Value = nil
		}
		*rows = append(*rows, row)
	}
}

func keysToSnakeCase(_ context.Context, d *transform.TransformData) (interface{}, error) {
	keys := d.Value.([]string)
	snakes := []string{}
	re := regexp.MustCompile(`[^A-Za-z0-9_]`)
	for _, k := range keys {
		snakes = append(snakes, re.ReplaceAllString(k, "_"))
	}
	return strings.Join(snakes, "."), nil
}
