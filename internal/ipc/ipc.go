package ipc

import (
	"github.com/nthienan/aws-dynamodb-cache-lambda-extension/internal/extension"
	"github.com/nthienan/aws-dynamodb-cache-lambda-extension/internal/plugins"
	"github.com/gorilla/mux"
	"net/http"
)

// Start begins running the sidecar
func Start(port string) {
	go startHTTPServer(port)
}

// Method that responds back with the cached values
func startHTTPServer(port string) {
	router := mux.NewRouter()
	router.Path("/{cacheType}").Queries("name", "{name}").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			value := extension.RouteCache(vars["cacheType"], vars["name"])

			if len(value) != 0 {
				_, _ = w.Write([]byte(value))
			} else {
				_, _ = w.Write([]byte("No data found"))
			}
		})

	println(plugins.PrintPrefix, "Starting Httpserver on port ", port)
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		panic(err)
	}
}
