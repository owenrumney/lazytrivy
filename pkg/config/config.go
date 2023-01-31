package config

import (
	"os"
	"path/filepath"

	"github.com/owenrumney/lazytrivy/pkg/logger"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Vulnerability  VulnerabilityConfig
	Filesystem     FileSystemConfig
	CacheDirectory string `json:"-"`
	Debug          bool
	Trace          bool
	DockerEndpoint string
}

type VulnerabilityConfig struct {
	IgnoreUnfixed bool
}

type FileSystemConfig struct {
	ScanSecrets          bool
	ScanMisconfiguration bool
	ScanVulnerabilities  bool
	WorkingDirectory     string `json:"-"`
}

var defaultConfig *Config

var configPath string

func createDefaultConfig() error {
	logger.Debugf("Creating default config")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Errorf("Error getting user home directory: %s", err)
		homeDir = os.TempDir()
	}
	trivyCacheDir := filepath.Join(homeDir, ".cache", "trivy")

	defaultConfig = &Config{
		CacheDirectory: trivyCacheDir,
		Debug:          false,
		Vulnerability: VulnerabilityConfig{
			IgnoreUnfixed: false,
		},
		Filesystem: FileSystemConfig{
			ScanSecrets:          false,
			ScanMisconfiguration: true,
			ScanVulnerabilities:  true,
		},
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	lazyTrivyConfigDir := filepath.Join(configDir, "lazytrivy")

	if err := os.MkdirAll(lazyTrivyConfigDir, os.ModePerm); err != nil {
		return err
	}

	configPath = filepath.Join(lazyTrivyConfigDir, "config.yaml")

	return nil
}

func Load() (*Config, error) {

	if err := createDefaultConfig(); err != nil {
		return nil, err
	}

	logger.Debugf("Attempting to load config from %s", configPath)
	if _, err := os.Stat(configPath); err != nil {
		logger.Debugf("No config file found, using defaults")
		return defaultConfig, nil
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		logger.Errorf("Error reading config file: %s", err)
		return defaultConfig, nil
	}

	if err := yaml.Unmarshal(content, &defaultConfig); err != nil {
		logger.Errorf("Error parsing config file: %s", err)
		return defaultConfig, err
	}
	logger.Infof("Loaded config from %s", configPath)

	return defaultConfig, nil
}

func Save(config *Config) error {
	logger.Debugf("Saving the config to %s", configPath)
	content, err := yaml.Marshal(config)
	if err != nil {
		logger.Errorf("Error marshalling config: %s", err)
		return err
	}

	if err := os.WriteFile(configPath, content, 0600); err != nil {
		logger.Errorf("Error writing config file: %s", err)
		return err
	}

	logger.Infof("Saved config to %s", configPath)
	return nil
}

func (c *Config) Save() error {
	return Save(c)
}
