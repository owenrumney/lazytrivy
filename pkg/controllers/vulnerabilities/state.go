package vulnerabilities

import "sync"

type state struct {
	stateLock     sync.Mutex
	images        []string
	selectedImage string
	imageWidth    int
}

func (s *state) updateImages(images []string) {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()
	s.images = images

	s.imageWidth = getLongestImageName(images)
	s.selectedImage = ""
}

func (s *state) setSelected(selectedImage string) {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()
	s.selectedImage = selectedImage
}

func getLongestImageName(images []string) int {
	imageWidth := 0
	for _, image := range images {
		if len(image) > imageWidth {
			imageWidth = len(image)
		}
	}
	return imageWidth
}
