package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/owenrumney/lazytrivy/pkg/logger"
	"gopkg.in/yaml.v3"
)

var defaultConfig *Config

var configPath string

type LegacyConfig struct {
	Vulnerability  VulnerabilityConfig
	Filesystem     FileSystemConfig
	CacheDirectory string `json:"-"`
	Debug          bool
	Trace          bool
	Insecure       bool
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

type Scanner struct {
	ScanSecrets          bool `yaml:"scan_secrets" json:"scan_secrets"`
	ScanMisconfiguration bool `yaml:"scan_misconfiguration" json:"scan_misconfiguration"`
	ScanVulnerabilities  bool `yaml:"scan_vulnerabilities" json:"scan_vulnerabilities"`
	IgnoreUnfixed        bool
}

type Config struct {
	Scanner          Scanner `yaml:"scanner" json:"scanner"`
	CacheDirectory   string  `yaml:"cache_directory" json:"cache_directory"`
	Debug            bool    `yaml:"debug" json:"debug"`
	Trace            bool    `yaml:"trace" json:"trace"`
	Insecure         bool    `yaml:"insecure" json:"insecure"`
	DockerEndpoint   string  `yaml:"docker_endpoint" json:"docker_endpoint"`
	WorkingDirectory string  `yaml:"working_directory" json:"working_directory"`
}

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
		Insecure:       false,
		Scanner: Scanner{
			ScanSecrets:          false,
			ScanMisconfiguration: true,
			ScanVulnerabilities:  true,
			IgnoreUnfixed:        false,
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

func migrateConfig(configPath string) error {
	var legacy LegacyConfig
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	if strings.Contains(string(content), "vulnerability") {
		if err := yaml.Unmarshal(content, &legacy); err != nil {
			return err
		}

		newConfig := &Config{
			CacheDirectory: legacy.CacheDirectory,
			Debug:          legacy.Debug,
			Trace:          legacy.Trace,
			Insecure:       legacy.Insecure,
			Scanner: Scanner{
				ScanSecrets:          legacy.Filesystem.ScanSecrets,
				ScanMisconfiguration: legacy.Filesystem.ScanMisconfiguration,
				ScanVulnerabilities:  legacy.Filesystem.ScanVulnerabilities,
				IgnoreUnfixed:        legacy.Vulnerability.IgnoreUnfixed,
			},
			DockerEndpoint: legacy.DockerEndpoint,
		}

		backupConfigPath := configPath + ".bak"
		if err := os.Rename(configPath, backupConfigPath); err != nil {
			return err
		}

		if err := newConfig.Save(); err != nil {
			return err
		}
	}
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

	if err := migrateConfig(configPath); err != nil {
		return nil, err
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
