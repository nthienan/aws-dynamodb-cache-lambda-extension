package plugins

import (
	"os"
	"testing"

	"gopkg.in/yaml.v2"
)

var (
	configFilePath string = "testdata/config.yaml"
)

func loadConfigFile() (*DynamoDbConfiguration, error) {
	wd, _ := os.Getwd()
	absPath := wd + string(os.PathSeparator) + configFilePath
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	config := DynamoDbConfiguration{}
	yaml.Unmarshal(data, &config)
	return &config, nil
}

func TestLoadConfigFile(t *testing.T) {
	config, err := loadConfigFile()
	if err != nil {
		t.Fatalf("Failed to read config file %s. Error: %s", configFilePath, err)
	}

	if config.Table == "" {
		t.Error("table is required")
	}

	if config.HashKey == "" {
		t.Error("hashKey is required")
	}

	if config.HashKeyType == "" {
		t.Error("hashKeyType is required")
	}

	if config.SortKey == "" {
		t.Error("sortKey is required")
	}

	if config.SortKeyType == "" {
		t.Error("sortKeyType is required")
	}
}

func TestGetData(t *testing.T) {
	config, err := loadConfigFile()
	if err != nil {
		t.Fatalf("Failed to read config file %s. Error: %s", configFilePath, err)
	}
	result := LoadData(*config)
    
	if result == false {
		t.Error("Expected data existing. Got empty data")
	}
}
