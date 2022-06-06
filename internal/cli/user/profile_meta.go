package user

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// ProfileMeta contains the name and full filepath of a profile
type ProfileMeta struct {
	Name     string
	Filepath string
}

// Profiles returns a list of each profile meta containing name and filepath
func Profiles() ([]ProfileMeta, error) {
	dir, dirErr := HomeDir()
	if dirErr != nil {
		return nil, fmt.Errorf("failed to get CLI profiles: %w", dirErr)
	}

	dirEntryList, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get CLI profiles: %w", dirErr)
	}

	profileMetas := make([]ProfileMeta, 0, len(dirEntryList))
	for _, v := range dirEntryList {
		if !strings.Contains(v.Name(), profileType) {
			continue
		}
		profileMetas = append(profileMetas, ProfileMeta{
			Name:     strings.TrimSuffix(v.Name(), filepath.Ext(v.Name())),
			Filepath: filepath.Join(dir, v.Name()),
		})
	}

	return profileMetas, nil
}
