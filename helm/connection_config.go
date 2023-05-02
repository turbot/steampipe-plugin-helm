package helm

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/schema"
)

type helmConfig struct {
	Paths []string `cty:"paths"`
}

var ConfigSchema = map[string]*schema.Attribute{
	"paths": {
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
