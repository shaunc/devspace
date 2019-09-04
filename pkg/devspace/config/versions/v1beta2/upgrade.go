package v1beta2

import (
	"github.com/devspace-cloud/devspace/pkg/devspace/config/versions/config"
	next "github.com/devspace-cloud/devspace/pkg/devspace/config/versions/latest"
	"github.com/devspace-cloud/devspace/pkg/devspace/config/versions/util"
	"github.com/devspace-cloud/devspace/pkg/util/log"
	"github.com/devspace-cloud/devspace/pkg/util/ptr"
)

// Upgrade upgrades the config
func (c *Config) Upgrade() (config.Config, error) {
	nextConfig := &next.Config{}
	err := util.Convert(c, nextConfig)
	if err != nil {
		return nil, err
	}

	// Check if old cluster exists
	if c.Cluster != nil && (c.Cluster.KubeContext != nil || c.Cluster.Namespace != nil) {
		log.Warnf("cluster config option is not supported anymore in v1beta2 and devspace v3")
	}

	if nextConfig.Dev == nil {
		nextConfig.Dev = &next.DevConfig{}
	}
	if nextConfig.Dev.Terminal == nil {
		nextConfig.Dev.Terminal = &next.Terminal{}
	}

	if c.Dev != nil && c.Dev.Terminal != nil && c.Dev.Terminal.Disabled != nil {
		nextConfig.Dev.Terminal.Enabled = ptr.Bool(!*c.Dev.Terminal.Disabled)
	} else {
		nextConfig.Dev.Terminal.Enabled = ptr.Bool(true)
	}

	return nextConfig, nil
}
