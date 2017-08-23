// +build mock

package atlas

import "errors"

// GetAllClusters returns all Atlas groups with their list of cluster names.
func GetAllClusters() (clusters map[string][]string, err error) {
	return map[string][]string{
		"group-1": []string{"cluster0", "cluster1"},
		"group-2": []string{"clustera", "clusterb"},
	}, nil
}

// GetClusters returns the Atlas clusters for the given group in arrays of (name, URI)
func GetClusters(group string) (clusters [][2]string, err error) {
	switch group {
	case "group-1":
		return [][2]string{
			{"cluster0", "mongodb://localhost:27017/test?ssl=false"},
			{"cluster1", "mongodb://localhost:27017/test?ssl=false"},
		}, nil
	case "group-2":
		return [][2]string{
			{"clustera", "mongodb://localhost:27017/test?ssl=false"},
			{"clusterb", "mongodb://localhost:27017/test?ssl=false"},
		}, nil
	}
	return nil, errors.New("not in group")
}

// GetClusterURI returns the URI corresponding to the specified Atlas cluster.
func GetClusterURI(group, cluster string) (uri string, err error) {
	switch group {
	case "group-1":
		switch cluster {
		case "cluster0", "cluster1":
			return "mongodb://localhost:27017/test?ssl=false", nil
		}
		return "", errors.New("cluster not in group")
	case "group-2":
		switch cluster {
		case "clustera", "clusterb":
			return "mongodb://localhost:27017/test?ssl=false", nil
		}
		return "", errors.New("cluster not in group")
	case "group-3":
		return "", errors.New("cluster not in group")
	}
	return "", errors.New("not in group")
}
