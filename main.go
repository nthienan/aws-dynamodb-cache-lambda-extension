package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/nthienan/aws-dynamodb-cache-lambda-extension/internal/extension"
	"github.com/nthienan/aws-dynamodb-cache-lambda-extension/internal/ipc"
	"github.com/nthienan/aws-dynamodb-cache-lambda-extension/internal/plugins"
	"github.com/nthienan/aws-dynamodb-cache-lambda-extension/internal/version"
    "github.com/alecthomas/kingpin/v2"
)

var (
	extensionClient = extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
)

func main() {
    kingpin.Version(version.Print("aws-dynamodb-cache-lambda-extension"))

    // parse flags
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		println(plugins.PrintPrefix, "Received", s)
		println(plugins.PrintPrefix, "Exiting")
	}()

	res, err := extensionClient.Register(ctx, plugins.ExtensionName)
	if err != nil {
		panic(err)
	}
	println(plugins.PrintPrefix, "Register response:", plugins.PrettyPrint(res))

	// Initialize all the cache plugins
	extension.InitCacheExtensions()

	// Start HTTP server
	ipc.Start("4000")

	// Will block until shutdown event is received or cancelled via the context.
	processEvents(ctx)
}

// Method to process events
func processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			res, err := extensionClient.NextEvent(ctx)
			if err != nil {
				println(plugins.PrintPrefix, "Error:", err)
				println(plugins.PrintPrefix, "Exiting")
				return
			}

			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				println(plugins.PrintPrefix, "Received SHUTDOWN event")
				println(plugins.PrintPrefix, "Exiting")
				return
			}
		}
	}
}
