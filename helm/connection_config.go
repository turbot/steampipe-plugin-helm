package helm

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/schema"
)

type helmConfig struct {
	ChartDir      *string  `cty:"chart_dir"`
	ValueOverride []string `cty:"value_override"`
}

var ConfigSchema = map[string]*schema.Attribute{
	"chart_dir": {
		Type: schema.TypeString,
	},
	"value_override": {
		Type: schema.TypeList,
		Elem: &schema.Attribute{Type: schema.TypeString},
	},
}

func ConfigInstance() interface{} {
	return &helmConfig{}
}

// GetConfig :: retrieve and cast connection config from query data
func GetConfig(connection *plugin.Connection) helmConfig {
	if connection == nil || connection.Config == nil {
		return helmConfig{}
	}
	config, _ := connection.Config.(helmConfig)
	return config
}
