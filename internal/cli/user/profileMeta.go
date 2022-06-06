package user

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ProfileMeta struct {
	Name     string // no file extension
	Filepath string // full filepath
}

func Profiles() ([]ProfileMeta, error) {
	dir, dirErr := HomeDir()
	if dirErr != nil {
		return nil, fmt.Errorf("failed to get CLI profiles: %w", dirErr)
	}

	dirEntryList, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get CLI profiles: %w", dirErr)
	}

	profileMetas := make([]ProfileMeta, 0, len(dirEntryList))
	for _, v := range dirEntryList {
		if strings.Contains(v.Name(), profileType) {
			name := strings.TrimSuffix(v.Name(), filepath.Ext(v.Name()))
			filepath := filepath.Join(dir, v.Name())
			profileMetas = append(profileMetas, ProfileMeta{Name: name, Filepath: filepath})
		}
	}

	return profileMetas, nil
}
