package filesystem

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type dir struct {
	dirs      map[string]*dir
	dirNames  []string
	files     map[string]struct{}
	fileIndex int
}

func newDir() *dir {
	return &dir{
		dirs:  make(map[string]*dir),
		files: make(map[string]struct{}),
	}
}

func (d *dir) addFile(target, original string) {
	sep := fmt.Sprintf("%c", os.PathSeparator)
	if strings.Contains(target, ":") {

		parts := strings.SplitN(target, ":", 2)
		root := parts[0]
		remaining := parts[1]

		if _, ok := d.dirs[root]; !ok {
			d.dirs[root] = newDir()
			d.dirNames = append(d.dirNames, root)
		}

		nd := d.dirs[root]
		nd.addFile(remaining, original)

	} else if strings.Contains(target, sep) {
		parts := strings.SplitN(target, sep, 2)
		root := parts[0]
		remaining := parts[1]

		if _, ok := d.dirs[root]; !ok {
			d.dirs[root] = newDir()
			d.dirNames = append(d.dirNames, root)
		}

		nd := d.dirs[root]
		nd.addFile(remaining, original)
	} else {
		fileID := fmt.Sprintf("%s|%s", target, original)

		if _, ok := d.files[fileID]; !ok {
			d.files[fileID] = struct{}{}
		}
	}
}

func (d *dir) generateTree(lines []string, depth int) []string {

	sort.Strings(d.dirNames)

	prefix := ""
	if depth >= 0 {
		prefix = fmt.Sprintf("%s└─ ", strings.Repeat(" ", depth))
		depth += 3
	}
	for _, dirName := range d.dirNames {
		if depth == -1 {
			depth = 0
		}
		children := d.dirs[dirName]
		lines = append(lines, fmt.Sprintf("%s%s", prefix, dirName))
		lines = children.generateTree(lines, depth)
	}

	for file, _ := range d.files {
		lines = append(lines, fmt.Sprintf("%s%s", prefix, file))
	}

	return lines
}

func createRootDir(targets []string) *dir {
	sort.Slice(targets, func(i, j int) bool {
		return strings.ToLower(targets[i]) < strings.ToLower(targets[j])
	})

	root := newDir()
	for _, target := range targets {
		root.addFile(target, target)
	}

	return root
}
