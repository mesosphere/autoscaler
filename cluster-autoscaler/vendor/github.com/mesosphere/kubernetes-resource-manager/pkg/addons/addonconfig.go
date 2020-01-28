package addons

// AddonConfig describes an addon to be installed on the cluster
type AddonConfig struct {
	Name    string `json:"name" yaml:"name"`
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Values  string `json:"values,omitempty" yaml:"values,omitempty"`
}

// AddonConfigs is the list of AddonConfig entries
type AddonConfigs []AddonConfig

// GetDefaultAddonConfigs gets the initial AddonConfig list from a joining of the TemplateRepos
func GetDefaultAddonConfigs(provider string, repos TemplateRepos) (AddonConfigs, error) {
	addons, err := AddonsAvailable(provider, repos)
	if err != nil {
		return nil, err
	}

	configs := AddonConfigs{}
	for _, addon := range addons {
		enabled := true
		values := ""
		for _, cp := range addon.GetAddonSpec().CloudProvider {
			if cp.Name == provider {
				enabled = cp.Enabled
				values = cp.Values
			}
		}
		config := AddonConfig{
			Name:    addon.GetName(),
			Enabled: enabled,
			Values:  values,
		}
		configs = append(configs, config)
	}
	return configs, nil
}
