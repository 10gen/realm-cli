package local

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// GraphQLStructure represents the Realm app graphql structure
type GraphQLStructure struct {
	Config          map[string]interface{}   `json:"config,omitempty"`
	CustomResolvers []map[string]interface{} `json:"custom_resolvers,omitempty"`
}

// SecretsStructure represents the Realm app secrets
type SecretsStructure struct {
	AuthProviders map[string]map[string]string `json:"auth_providers,omitempty"`
	Services      map[string]map[string]string `json:"services,omitempty"`
}

// ServiceStructure represents the Realm app service structure
type ServiceStructure struct {
	Config           map[string]interface{}   `json:"config,omitempty"`
	IncomingWebhooks []map[string]interface{} `json:"incoming_webhooks"`
	Rules            []map[string]interface{} `json:"rules"`
}

func parseEnvironments(rootDir string) (map[string]map[string]interface{}, error) {
	out := map[string]map[string]interface{}{}

	dw := directoryWalker{
		path:      filepath.Join(rootDir, NameEnvironments),
		onlyFiles: true,
	}
	if err := dw.walk(func(file os.FileInfo, path string) error {
		o, err := parseJSON(path)
		if err != nil {
			return err
		}
		out[file.Name()] = o
		return nil
	}); err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

func parseFunctions(rootDir string) ([]map[string]interface{}, error) {
	if _, err := os.Stat(rootDir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var out []map[string]interface{}

	dw := directoryWalker{path: rootDir, onlyDirs: true}
	if walkErr := dw.walk(func(file os.FileInfo, path string) error {
		if strings.Contains(path, nameNodeModules) {
			return nil // skip node_modules
		}

		config, configErr := parseJSON(filepath.Join(path, FileConfig.String()))
		if configErr != nil {
			return configErr
		}

		src, srcErr := parseJavascript(path, FileSource)
		if srcErr != nil {
			return srcErr
		}

		out = append(out, map[string]interface{}{
			NameConfig: config,
			NameSource: src,
		})
		return nil
	}); walkErr != nil {
		return nil, walkErr
	}

	return out, nil
}

func parseGraphQL(rootDir string) (GraphQLStructure, bool, error) {
	dir := filepath.Join(rootDir, NameGraphQL)

	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return GraphQLStructure{}, false, nil
		}
		return GraphQLStructure{}, false, err
	}

	config, configErr := parseJSON(filepath.Join(dir, FileConfig.String()))
	if configErr != nil {
		return GraphQLStructure{}, false, configErr
	}

	customResolvers, customResolversErr := parseJSONFiles(filepath.Join(dir, NameCustomResolvers))
	if customResolversErr != nil {
		return GraphQLStructure{}, false, customResolversErr
	}

	return GraphQLStructure{config, customResolvers}, true, nil
}

func parseJavascript(rootDir string, file File) (string, error) {
	src, err := ioutil.ReadFile(filepath.Join(rootDir, file.String()))
	if err != nil {
		return "", err
	}
	return string(src), nil
}

func parseJSON(path string) (map[string]interface{}, error) {
	data, dataErr := readFile(path)
	if dataErr != nil {
		return nil, dataErr
	}

	var out map[string]interface{}
	if err := unmarshalJSON(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func parseJSONArray(path string) ([]map[string]interface{}, error) {
	data, dataErr := readFile(path)
	if dataErr != nil {
		return nil, dataErr
	}

	var out []map[string]interface{}
	if err := unmarshalJSON(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func parseJSONFiles(rootDir string) ([]map[string]interface{}, error) {
	if _, err := os.Stat(rootDir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	out := make([]map[string]interface{}, 0)

	dw := directoryWalker{path: rootDir, onlyFiles: true}
	if walkErr := dw.walk(func(file os.FileInfo, path string) error {
		o, err := parseJSON(path)
		if err != nil {
			return err
		}

		out = append(out, o)
		return nil
	}); walkErr != nil {
		return nil, walkErr
	}

	return out, nil
}

func parseSecrets(rootDir string) (*SecretsStructure, error) {
	path := filepath.Join(rootDir, FileSecrets.String())

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	data, dataErr := readFile(path)
	if dataErr != nil {
		return nil, dataErr
	}

	var secrets SecretsStructure
	if err := json.Unmarshal(data, &secrets); err != nil {
		return nil, err
	}
	return &secrets, nil
}

func parseServices(rootDir string) ([]ServiceStructure, error) {
	var out []ServiceStructure

	dw := directoryWalker{
		path:     filepath.Join(rootDir, NameServices),
		onlyDirs: true,
	}
	if walkErr := dw.walk(func(file os.FileInfo, path string) error {
		var svc ServiceStructure

		config, err := parseJSON(filepath.Join(path, FileConfig.String()))
		if err != nil {
			return err
		}
		svc.Config = config

		webhooks, err := parseFunctions(filepath.Join(path, NameIncomingWebhooks))
		if err != nil {
			return err
		}
		svc.IncomingWebhooks = webhooks

		rules, err := parseJSONFiles(filepath.Join(path, NameRules))
		if err != nil {
			return err
		}
		svc.Rules = rules

		out = append(out, svc)
		return nil
	}); walkErr != nil {
		return nil, walkErr
	}

	return out, nil
}
