package local

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
	DefaultRule      map[string]interface{}   `json:"default_rule,omitempty"`
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

func parseSecrets(rootDir string) (SecretsStructure, error) {
	path := filepath.Join(rootDir, FileSecrets.String())

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return SecretsStructure{}, nil
		}
		return SecretsStructure{}, err
	}

	data, dataErr := readFile(path)
	if dataErr != nil {
		return SecretsStructure{}, dataErr
	}

	var secrets SecretsStructure
	if err := json.Unmarshal(data, &secrets); err != nil {
		return SecretsStructure{}, err
	}
	return secrets, nil
}

func parseServices(rootDir string) ([]ServiceStructure, error) {
	var out []ServiceStructure

	dw := directoryWalker{
		path:     filepath.Join(rootDir, NameServices),
		onlyDirs: true,
	}
	if walkErr := dw.walk(func(file os.FileInfo, path string) error {
		var svc ServiceStructure

		// Config
		config, err := parseJSON(filepath.Join(path, FileConfig.String()))
		if err != nil {
			return err
		}
		svc.Config = config

		// Default Rule
		defaultRule, err := parseJSON(filepath.Join(path, FileDefaultRule.String()))
		if err != nil {
			return err
		}
		svc.DefaultRule = defaultRule

		// Webhooks
		webhooks, err := parseFunctions(filepath.Join(path, NameIncomingWebhooks))
		if err != nil {
			return err
		}
		svc.IncomingWebhooks = webhooks

		// Rules
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

func writeSecrets(rootDir string, secrets SecretsStructure) error {
	if len(secrets.AuthProviders) == 0 && len(secrets.Services) == 0 {
		return nil
	}
	data, err := MarshalJSON(secrets)
	if err != nil {
		return err
	}
	return WriteFile(filepath.Join(rootDir, FileSecrets.String()), 0666, bytes.NewReader(data))
}

func writeEnvironments(rootDir string, environments map[string]map[string]interface{}) error {
	for env, values := range environments {
		data, err := MarshalJSON(values)
		if err != nil {
			return err
		}
		if err := WriteFile(
			filepath.Join(rootDir, NameEnvironments, env),
			0666,
			bytes.NewReader(data),
		); err != nil {
			return err
		}
	}
	return nil
}

func writeValues(rootDir string, values []map[string]interface{}) error {
	dir := filepath.Join(rootDir, NameValues)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	for _, value := range values {
		data, err := MarshalJSON(value)
		if err != nil {
			return err
		}
		if err := WriteFile(
			filepath.Join(dir, fmt.Sprintf("%s%s", value["name"], extJSON)),
			0666,
			bytes.NewReader(data),
		); err != nil {
			return err
		}
	}
	return nil
}

func writeGraphQL(rootDir string, graphql GraphQLStructure) error {
	dir := filepath.Join(rootDir, NameGraphQL)
	if err := os.MkdirAll(filepath.Join(dir, NameCustomResolvers), os.ModePerm); err != nil {
		return err
	}
	if graphql.Config != nil {
		data, err := MarshalJSON(graphql.Config)
		if err != nil {
			return err
		}
		if err := WriteFile(
			filepath.Join(dir, FileConfig.String()),
			0666,
			bytes.NewReader(data),
		); err != nil {
			return err
		}
	}
	for _, customResolver := range graphql.CustomResolvers {
		data, err := MarshalJSON(customResolver)
		if err != nil {
			return err
		}
		nameFile := strings.ToLower(fmt.Sprintf("%s_%s%s", customResolver["on_type"], customResolver["field_name"], extJSON))
		if err := WriteFile(
			filepath.Join(dir, NameCustomResolvers, nameFile),
			0666,
			bytes.NewReader(data),
		); err != nil {
			return err
		}
	}
	return nil
}

func writeServices(rootDir string, services []ServiceStructure) error {
	dir := filepath.Join(rootDir, NameServices)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	for _, svc := range services {
		nameSvc, ok := svc.Config["name"].(string)
		if !ok {
			return errors.New("error writing services")
		}
		dirSvc := filepath.Join(dir, nameSvc)

		// Config
		data, err := MarshalJSON(svc.Config)
		if err != nil {
			return err
		}
		if err := WriteFile(
			filepath.Join(dirSvc, FileConfig.String()),
			0666,
			bytes.NewReader(data),
		); err != nil {
			return err
		}

		// Default Rule
		if svc.DefaultRule != nil {
			defaultRule, err := MarshalJSON(svc.DefaultRule)
			if err != nil {
				return err
			}
			if err := WriteFile(
				filepath.Join(dirSvc, FileDefaultRule.String()),
				0666,
				bytes.NewReader(defaultRule),
			); err != nil {
				return err
			}
		}

		// Webhooks
		for _, webhook := range svc.IncomingWebhooks {
			src, ok := webhook[NameSource].(string)
			if !ok {
				return errors.New("error writing services")
			}
			nameWebhook, ok := webhook["name"].(string)
			if !ok {
				return errors.New("error writing services")
			}
			webhookTemp := map[string]interface{}{}
			for k, v := range webhook {
				webhookTemp[k] = v
			}
			delete(webhookTemp, NameSource)
			data, err := MarshalJSON(webhookTemp)
			if err != nil {
				return err
			}
			if err := WriteFile(
				filepath.Join(dirSvc, NameIncomingWebhooks, nameWebhook, FileConfig.String()),
				0666,
				bytes.NewReader(data),
			); err != nil {
				return err
			}
			if err := WriteFile(
				filepath.Join(dirSvc, NameIncomingWebhooks, nameWebhook, FileSource.String()),
				0666,
				bytes.NewReader([]byte(src)),
			); err != nil {
				return err
			}
		}

		// Rules
		for _, rule := range svc.Rules {
			data, err := MarshalJSON(rule)
			if err != nil {
				return err
			}

			// Mongo service rules (AKA NamespaceRules) do not have an exported "name" field and should construct the rule name
			// using the database and collection values
			ruleName := fmt.Sprintf("%s.%s", rule["database"], rule["collection"])

			if _, isMongoSvcRule := rule["database"]; !isMongoSvcRule {
				// Rules with a type of BuiltinRule have an exported "name" field. These are rules used for non-Mongo services
				if name, ok := rule["name"].(string); ok {
					ruleName = name
				}
			}

			if err := WriteFile(
				filepath.Join(dirSvc, NameRules, fmt.Sprintf("%s%s", ruleName, extJSON)),
				0666,
				bytes.NewReader(data),
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeTriggers(rootDir string, triggers []map[string]interface{}) error {
	for _, trigger := range triggers {
		name, ok := trigger["name"].(string)
		if !ok {
			return errors.New("error writing triggers")
		}
		data, err := MarshalJSON(trigger)
		if err != nil {
			return err
		}
		if err := WriteFile(
			filepath.Join(rootDir, NameTriggers, name+extJSON),
			0666,
			bytes.NewReader(data),
		); err != nil {
			return err
		}
	}
	return nil
}

func writeLogForwarders(rootDir string, logForwarders []map[string]interface{}) error {
	for _, lf := range logForwarders {
		name, ok := lf["name"].(string)
		if !ok {
			return errors.New("error writing log forwarders")
		}
		data, err := MarshalJSON(lf)
		if err != nil {
			return err
		}
		if err := WriteFile(
			filepath.Join(rootDir, NameLogForwarders, name+extJSON),
			0666,
			bytes.NewReader(data),
		); err != nil {
			return err
		}
	}
	return nil
}

func writeEndpoints(rootDir string, endpoints EndpointStructure) error {
	data, err := MarshalJSON(endpoints.Configs)
	if err != nil {
		return err
	}
	if err := WriteFile(
		filepath.Join(rootDir, NameHTTPEndpoints, FileConfig.String()),
		0666,
		bytes.NewReader(data),
	); err != nil {
		return err
	}
	return nil
}
