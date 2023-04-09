package extension

import (
	"log"
	"os"
	"strconv"

	"github.com/nthienan/aws-dynamodb-cache-lambda-extension/internal/plugins"
	"gopkg.in/yaml.v2"
)

// Constants definition
const (
	Parameters = "parameters"
	Dynamodb   = "dynamodb"
	FileName                 = "/var/task/cache.yaml"
	InitializeCacheOnStartup = "CACHE_EXTENSION_INIT_STARTUP"
)

// Struct for storing CacheConfiguration
type CacheConfig struct {
	DynamoDb []plugins.DynamoDbConfiguration
}

var cacheConfig = CacheConfig{}

// Initialize cache and start the background process to refresh cache
func InitCacheExtensions() {
	// Read the cache config file
	data := LoadConfigFile()

	// Unmarshal the configuration to struct
	err := yaml.Unmarshal([]byte(data), &cacheConfig)
	if err != nil {
		log.Fatalf(plugins.PrintPrefix, "error: %v", err)
	}

	// Initialize Cache
	println(plugins.PrintPrefix, "Initializing cache ...")
	InitCache()
	println(plugins.PrintPrefix, "Cache successfully loaded")
}

// Initialize individual cache
func InitCache() {

	// Read Lambda env variable
	var initCache = os.Getenv(InitializeCacheOnStartup)
	var initCacheInBool = false
	if initCache != "" {
		cacheInBool, err := strconv.ParseBool(initCache)
		if err != nil {
			panic(plugins.PrintPrefix + "Error while converting CACHE_EXTENSION_INIT_STARTUP env variable " +
				initCache)
		} else {
			initCacheInBool = cacheInBool
		}
	}

	// Initialize map and load data from individual services if "CACHE_EXTENSION_INIT_STARTUP" = true
	plugins.InitDynamodb(cacheConfig.DynamoDb, initCacheInBool)
}

// Route request to corresponding cache handlers
func RouteCache(cacheType string, name string) string {
	switch cacheType {
	case Dynamodb:
		return plugins.FetchDynamoDbCache(name)
	default:
		return ""
	}
}

// Load the config file
func LoadConfigFile() string {
	data, err := os.ReadFile(FileName)
	if err != nil {
		panic(err)
	}

	return string(data)
}
