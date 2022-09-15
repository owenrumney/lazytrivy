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
	Filesystem     FileSystemConfig
	CacheDirectory string `json:"-"`
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

type FileSystemConfig struct {
	ScanSecrets          bool
	ScanMisconfiguration bool
	ScanVulnerabilities  bool
	WorkingDirectory     string `json:"-"`
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
		Debug:          false,
		AWS: AWSConfig{
			CacheDirectory: awsCacheDir,
		},
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
		return
	}

	lazyTrivyConfigDir := filepath.Join(configDir, "lazytrivy")

	_ = os.MkdirAll(lazyTrivyConfigDir, os.ModePerm)

	configPath = filepath.Join(lazyTrivyConfigDir, "config.yaml")
}

func Load() (*Config, error) {
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

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	defaultConfig.Filesystem.WorkingDirectory = cwd

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

	return nil
}

func (c *Config) Save() error {
	return Save(c)
}
