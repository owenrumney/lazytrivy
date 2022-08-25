package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AWS            AWSConfig
	CacheDirectory string
}

type AWSConfig struct {
	AccountNo      string
	Region         string
	CacheDirectory string
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
		AWS: AWSConfig{
			CacheDirectory: awsCacheDir,
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
	if _, err := os.Stat(configPath); err != nil {
		return defaultConfig, nil
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return defaultConfig, nil
	}

	if err := yaml.Unmarshal(content, &defaultConfig); err != nil {
		return defaultConfig, err
	}

	return defaultConfig, nil
}

func Save(config *Config) error {
	content, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, content, 0644); err != nil {
		return err
	}

	return nil
}
