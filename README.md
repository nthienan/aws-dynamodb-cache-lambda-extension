# AWS Lambda Extension DynamoDB Caching

This extension is inspired by [aws-samples/aws-lambda-extensions](https://github.com/aws-samples/aws-lambda-extensions). This has some modifications and enhancements compared to the original one from AWS.

# Introduction
Having a caching layer inside the Lambda function is a very common use case. It would allow the lambda function to process requests quicker and avoid the additional cost incurred by calling the AWS services over and over again. They are two types of cache:
- Data cache (caching data from databases like RDS, dynamodb, etc.)
- Configuration cache (caching data from a configuration system like parameter store, app config, secrets, etc.)

This extension demo's the Lambda layer that enables data cache using dynamodb.

Here is how it works:
- Uses `cache.yaml` defined part of the lambda function to determine the DynamoDB table that needs to be cached
- All the data are cached in memory before the request gets handled to the lambda function. So no cold start problems
- Starts a local HTTP server at port `4000` that replies to request for reading items from the cache depending upon path variables
- Uses `"CACHE_EXTENSION_TTL"` Lambda environment variable to let users define cache refresh interval (defined based on Go time format, ex: 30s, 3m, 24h etc)
- Uses `"CACHE_EXTENSION_INIT_STARTUP"` Lambda environment variable used to specify whether to load all items specified in `"cache.yml"` into cache part of extension startup (takes boolean value, ex: true and false)

Here are some advantages of having the cache layer part of Lambda extension instead of having it inside the function
- Reuse the code related to cache in multiple Lambda functions
- Common dependencies like SDK are packaged part of the Lambda layers

Once deployed the extension performs the following steps:
1.	On start-up, the extension reads the `cache.yaml` file which determines which resources to cache. The file is deployed as part of the lambda function.
2.	The boolean `CACHE_EXTENSION_INIT_STARTUP` Lambda environment variable specifies whether to load into cache the items specified in `cache.yaml`. If false, nothing happens.
3.	The extension retrieves the required data from DynamoDB. The data is stored in memory.
4.	The extension starts a local HTTP server using TCP port 4000 which serves the cache items to the function. The Lambda can accessed the local in-memory cache by invoking the following endpoint: `http://localhost:4000/dynamodb?name=<name>`. `name` is `<table_name>@@<hash_key_value>@@<sort_key_value>`
5.	If the data is not available in the cache, or has expired, the extension accesses the corresponding AWS service to retrieve the data. It is cached first, and then returned to the lambda function. The `CACHE_EXTENSION_TTL` Lambda environment variable defines the refresh interval (defined based on Go time format, ex: 30s, 3m, 24h etc.)


# Conclusion

This cache extension provides a secure way of caching data in parameter store, and DynamoDB also provides a way to implement TTL for cache items. By using this framework, we can reuse the caching code among multiple lambda functions and package all the required AWS dependencies part of AWS layers.
