package main

import (
	"github.com/turbot/steampipe-plugin-helm/helm"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		PluginFunc: helm.Plugin})
}
