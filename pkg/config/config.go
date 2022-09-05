package config

import (
	"os"
	"path/filepath"

	"github.com/owenrumney/lazytrivy/pkg/logger"
	"gopkg.in/yaml.v3"
)

type Config struct {
	AWS            AWSConfig
	Vulnerability  VulnerabilityConfig
	CacheDirectory string
	Debug          bool
}

type AWSConfig struct {
	AccountNo      string
	Region         string
	CacheDirectory string
}

type VulnerabilityConfig struct {
	IgnoreUnfixed bool
}

var defaultConfig *Config

var configPath string

func init() {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}
	trivyCacheDir := filepath.Join(homeDir, ".cache", "trivy")
	awsCacheDir := filepath.Join(trivyCacheDir, "cloud", "aws")

	defaultConfig = &Config{
		CacheDirectory: trivyCacheDir,
		Debug:          true,
		AWS: AWSConfig{
			CacheDirectory: awsCacheDir,
		},
		Vulnerability: VulnerabilityConfig{
			IgnoreUnfixed: false,
		},
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return
	}

	lazyTrivyConfigDir := filepath.Join(configDir, "lazytrivy")

	_ = os.MkdirAll(lazyTrivyConfigDir, os.ModePerm)

	configPath = filepath.Join(lazyTrivyConfigDir, "config.yaml")

}

func Load() (*Config, error) {
	logger.Debug("Attempting to load config from %s", configPath)
	if _, err := os.Stat(configPath); err != nil {
		logger.Debug("No config file found, using defaults")
		return defaultConfig, nil
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		logger.Error("Error reading config file: %s", err)
		return defaultConfig, nil
	}

	if err := yaml.Unmarshal(content, &defaultConfig); err != nil {
		logger.Error("Error parsing config file: %s", err)
		return defaultConfig, err
	}

	return defaultConfig, nil
}

func Save(config *Config) error {
	logger.Debug("Saving the config to %s", configPath)
	content, err := yaml.Marshal(config)
	if err != nil {
		logger.Error("Error marshalling config: %s", err)
		return err
	}

	if err := os.WriteFile(configPath, content, 0644); err != nil {
		logger.Error("Error writing config file: %s", err)
		return err
	}

	return nil
}

func (c *Config) Save() error {
	return Save(c)
}
