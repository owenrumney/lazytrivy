package image

import "sync"

type state struct {
	stateLock     sync.Mutex
	images        []string
	selectedImage string
	imageWidth    int
}

func (s *state) updateImages(images []string, maxWidth int) {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()
	s.images = images

	s.imageWidth = getLongestImageName(images, maxWidth)
	s.selectedImage = ""
}

func (s *state) setSelected(selectedImage string) {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()
	s.selectedImage = selectedImage
}

func getLongestImageName(images []string, maxWidth int) int {
	imageWidth := 0
	for _, image := range images {
		if len(image) > imageWidth && len(image) < maxWidth {
			imageWidth = len(image)
		}
	}
	return imageWidth
}
