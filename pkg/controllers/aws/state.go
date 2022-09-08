package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

type state struct {
	services        []string
	selectedService string
	serviceWidth    int
	cacheDirectory  string
}

func (s *state) accountRegionCache(accountID, region string) string {
	cacheDir := filepath.Join(s.cacheDirectory, accountID, region, "data")
	return cacheDir
}

func (s *state) listAccountNumbers() ([]string, error) {
	logger.Debugf("listing account numbers")
	var accountNumbers []string
	fileInfos, err := os.ReadDir(s.cacheDirectory)
	if err != nil {
		return nil, err
	}
	for i, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			accountNumbers = append(accountNumbers, fmt.Sprintf("**%d***  %s  ", i, fileInfo.Name()))
		}
	}
	return accountNumbers, nil
}

func (s *state) listRegions(accountNumber string) ([]string, error) {
	logger.Debugf("listing regions")
	var regions []string
	accountPath := filepath.Join(s.cacheDirectory, accountNumber)
	fileInfos, err := os.ReadDir(accountPath)
	if err != nil {
		return nil, err
	}
	for i, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			regions = append(regions, fmt.Sprintf("**%d***  %s  ", i, fileInfo.Name()))
		}
	}
	return regions, nil
}

func (s *state) accountRegionCacheExists(accountID, region string) bool {
	if _, err := os.Stat(s.accountRegionCache(accountID, region)); err == nil {
		return true
	}
	logger.Debugf("cache does not exist for %s (%s)", accountID, region)
	return false
}

func (s *state) accountRegionCacheServices(accountID, region string) ([]string, error) {
	var services []string
	if !s.accountRegionCacheExists(accountID, region) {
		return []string{}, nil
	}

	cachePath := s.accountRegionCache(accountID, region)

	if err := filepath.WalkDir(cachePath, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == cachePath {
			return nil
		}

		if info.IsDir() {
			services = append(services, info.Name())
		}
		return nil
	}); err != nil {
		return []string{}, err
	}

	s.services = services
	s.serviceWidth = getLongestName(services)

	return services, nil
}

func (s *state) getServiceReport(accountID, region, serviceName string) (*output.Report, error) {
	cachePath := s.accountRegionCache(accountID, region)

	serviceFile := filepath.Join(cachePath, serviceName, "cache.json")

	content, err := os.ReadFile(serviceFile)
	if err != nil {
		return nil, err
	}

	var report output.Report
	if err := json.Unmarshal(content, &report); err != nil {
		logger.Errorf("failed to unmarshal report: %s", err)
		return nil, err
	}
	report.Process()

	return &report, nil
}

func getLongestName(names []string) int {
	width := 0
	for _, name := range names {
		if len(name) > width {
			width = len(name)
		}
	}
	return width
}
