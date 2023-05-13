package plugins

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

// Struct to store Dynamodb cache confirmation
type DynamoDbConfiguration struct {
	Table        string `yaml:"table"`
	HashKey      string `yaml:"hashKey"`
	HashKeyType  string `yaml:"hashKeyType"`
	HashKeyValue string `yaml:"hashKeyValue"`
	SortKey      string `yaml:"sortKey"`
	SortKeyType  string `yaml:"sortKeyType"`
	SortKeyValue string `yaml:"sortKeyValue"`
	Fields       string `yaml:"fields"`
}

// Struct for caching the information
type DynamoDbCache struct {
	Data   CacheData
	Config DynamoDbConfiguration
}

var (
	dynamoDbCache  = make(map[string]DynamoDbCache)
	dynamoDbClient = GetDynamoDbClient()
)
var initializedConfig map[string]DynamoDbConfiguration

// Initialize map and cache data (only if requested)
func InitDynamodb(configs []DynamoDbConfiguration, initializeCache bool) {
	initializedConfig = make(map[string]DynamoDbConfiguration, len(configs))
	for _, config := range configs {
		initializedConfig[config.Table] = config
		if initializeCache {
			// Load data from Dynamodb
			LoadData(config)
		}
	}
}

func buildProjectionExpression(fieldExpr string) (*string, map[string]*string) {
	if len(fieldExpr) != 0 {
		fields := strings.Split(fieldExpr, ",")
		println(PrintPrefix, "Projection fields: ", fields)

		proj := expression.NamesList(expression.Name(fields[0]))
		for i := 1; i < len(fields); i++ {
			proj = expression.AddNames(proj, expression.Name(fields[i]))
		}

		expr, err := expression.NewBuilder().
			WithProjection(proj).
			Build()
		if err != nil {
			println(PrintPrefix, "Caught an unexpected error: %s", err)
		}
		return expr.Projection(), expr.Names()
	}
	return nil, nil
}

// Load data from Dynamodb
func LoadData(config DynamoDbConfiguration) bool {
	if config.HashKey != "" {

		// Set up the input parameters for the Scan operation
		projection, attributeNames := buildProjectionExpression(config.Fields)
		params := &dynamodb.ScanInput{
			TableName:                aws.String(config.Table),
			ProjectionExpression:     projection,
			ExpressionAttributeNames: attributeNames,
		}

		// Execute the Scan operation to read every item in the table.
		items := make([]map[string]interface{}, 0)
		err := dynamoDbClient.ScanPages(params, func(page *dynamodb.ScanOutput, lastPage bool) bool {
			// Unmarshal the page of items into a slice of structs.
			pageItems := make([]map[string]interface{}, len(page.Items))
			err := dynamodbattribute.UnmarshalListOfMaps(page.Items, &pageItems)
			if err != nil {
				fmt.Println("Error unmarshaling page items:", err)
				return false
			}

			// Add the page of items to the overall slice.
			items = append(items, pageItems...)

			// If there are more pages, continue scanning.
			return !lastPage
		})
		if err != nil {
			fmt.Println("Error scanning table:", err)
			return false
		}

		// Add it to the cache
		for _, item := range items {
			key := GenerateCacheKey(config, item)
			jsonData, err := json.Marshal(item)
			if err != nil {
				print(err.Error())
			}

			// create a new config object with store hash key value and sort key value to retrieve item when cache exipre
			new_config := new(DynamoDbConfiguration)

			// Copy the values from config to new_config
			*new_config = config

			// Update HashKeyValue and SortKeyValue
			new_config.HashKeyValue, _ = GetHashKeyValue(item, config)
			if config.SortKey != "" {
				new_config.SortKeyValue, _ = GetSortKeyValue(item, config)
			}

			dynamoDbCache[key] = DynamoDbCache{
				Data: CacheData{
					Data:        string(jsonData),
					CacheExpiry: GetCacheExpiry(),
				},
				Config: *new_config,
			}
		}

		return true
	} else {
		println(PrintPrefix, "HashKey not available so caching will not be enabled for %s", config.HashKey)
		return false
	}
}

// Get hash key value from an item in the table based on given configuration
func GetHashKeyValue(data map[string]interface{}, config DynamoDbConfiguration) (string, error) {
	rawHasKeyValue := data[config.HashKey]
	hasKeyValue, ok := rawHasKeyValue.(string)
	if ok {
		return hasKeyValue, nil
	} else {
		return "", fmt.Errorf("the value of hash key %s not a string. Expected a string value", config.HashKey)
	}
}

// Get sort key value from an item in the table based on given configuration
func GetSortKeyValue(data map[string]interface{}, config DynamoDbConfiguration) (string, error) {
	rawSortKeyValue := data[config.SortKey]
	sortKeyValue, ok := rawSortKeyValue.(string)
	if ok {
		return sortKeyValue, nil
	} else {
		return "", fmt.Errorf("the value of sort key %s not a string. Expected a string value", config.SortKey)
	}
}

// Generate key to store in map based with a format "tableName + "@@" + hashKeyValue + "@@" + sortKeyValue"
func GenerateCacheKey(config DynamoDbConfiguration, data map[string]interface{}) string {
	hasKeyValue, err := GetHashKeyValue(data, config)
	if err != nil {
		fmt.Println("Error while gettting has key value.", err)
	}

	var key = config.Table + "@@" + hasKeyValue

	if config.SortKey != "" {
		sortKeyValue, err := GetSortKeyValue(data, config)
		if err != nil {
			fmt.Println("Error while gettting sort key value.", err)
		}
		key += "@@" + sortKeyValue
	}

	return key
}

// Read specific data from Dynamodb
func GetData(config DynamoDbConfiguration) string {
	println(PrintPrefix, "Fetch data to cache for '"+config.HashKeyValue+"'")
	if config.HashKey != "" {
		// Create attributeValue map based on hash and sort key
		var attributeMap = map[string]*dynamodb.AttributeValue{}
		UpdateAttributeMap(attributeMap, config)
		projection, attributeNames := buildProjectionExpression(config.Fields)

		result, err := dynamoDbClient.GetItem(&dynamodb.GetItemInput{
			TableName:                aws.String(config.Table),
			Key:                      attributeMap,
			ProjectionExpression:     projection,
			ExpressionAttributeNames: attributeNames,
		})

		if err != nil {
			println(PrintPrefix, PrettyPrint(err.Error()))
			return ""
		}

		if result.Item == nil {
			println(PrintPrefix, "Could not find '"+config.HashKeyValue+"'")
			return ""
		}

		// Convert data from Map to JSON string
		var data = make(map[string]interface{})
		_ = dynamodbattribute.UnmarshalMap(result.Item, &data)

		// Convert map to JSON string
		jsonData, err := json.Marshal(data)
		if err != nil {
			println(err.Error())
		}

		// Add it to the cache
		var value = string(jsonData)
		dynamoDbCache[GenerateCacheKey(config, data)] = DynamoDbCache{
			Data: CacheData{
				Data:        value,
				CacheExpiry: GetCacheExpiry(),
			},
			Config: config,
		}

		return value
	} else {
		println(PrintPrefix, "Hash key not available so caching will not be enabled for %s", config.HashKey)
		return ""
	}
}

// Create attributeValue based on key type and presence of sortKey definition
func UpdateAttributeMap(attributeMap map[string]*dynamodb.AttributeValue, dynamodbConfig DynamoDbConfiguration) {
	GetAttributeValue(attributeMap, dynamodbConfig.HashKey, dynamodbConfig.HashKeyValue, dynamodbConfig.HashKeyType)
	if dynamodbConfig.SortKey != "" {
		GetAttributeValue(attributeMap, dynamodbConfig.SortKey, dynamodbConfig.SortKeyValue, dynamodbConfig.SortKeyType)
	}
}

// Supports attributeValue with data types "S" and "N"
func GetAttributeValue(attributeMap map[string]*dynamodb.AttributeValue, key string, value string, keyType string) {
	switch keyType {
	case "S":
		attributeMap[key] = &dynamodb.AttributeValue{S: aws.String(value)}
	case "N":
		attributeMap[key] = &dynamodb.AttributeValue{N: aws.String(value)}
	}
}

// Get Dynamodb to read data
func GetDynamoDbClient() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create Dynamodb client
	return dynamodb.New(sess)
}

// Fetch data from cache
func FetchDynamoDbCache(name string) string {
	var dbCache = dynamoDbCache[name]

	// If expired or not available in cache then read it from Dynamodb, else return from cache
	if dbCache.Data.Data == "" || IsExpired(dbCache.Data.CacheExpiry) {
		config := dynamoDbCache[name].Config
		if dbCache.Config.HashKeyValue == "" {
			cacheKeyInfo := strings.Split(name, "@@")
			tableName := cacheKeyInfo[0]
			config = initializedConfig[tableName]
			config.HashKeyValue = cacheKeyInfo[1]
			config.SortKeyValue = cacheKeyInfo[2]
		}
		return GetData(config)
	} else {
		return dynamoDbCache[name].Data.Data
	}
}
