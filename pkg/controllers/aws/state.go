package aws

import (
	"os"
	"path/filepath"
	"sync"
)

type state struct {
	stateLock       sync.Mutex
	services        []string
	selectedService string
	serviceWidth    int
	cacheDirectory  string
}

func (s *state) accountRegionCache(accountID, region string) string {
	cacheDir := filepath.Join(s.cacheDirectory, accountID, region, "data")
	return cacheDir
}

func (s *state) accountRegionCacheExists(accountID, region string) bool {
	if _, err := os.Stat(s.accountRegionCache(accountID, region)); err == nil {
		return true
	}
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

func (s *state) updateServices(services []string) {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()
	s.services = services

	s.serviceWidth = getLongestName(services)
	s.selectedService = ""
}

func (s *state) setSelected(selectedImage string) {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()
	s.selectedService = selectedImage
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
