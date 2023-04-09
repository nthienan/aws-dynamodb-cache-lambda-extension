GO           ?= go
GOFMT        ?= $(GO) fmt

VERSION		?= $(shell cat VERSION)
GIT_BRANCH	?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))
GIT_COMMIT	?= $(shell git rev-parse HEAD)
USER		?= $(shell echo $USER)
DATE		?= $(shell echo `date`)

APP_BIN	?= aws-dynamodb-cache-lambda-extension
APP_SRC	?= main.go


.PHONY: build
build: go-vet
	@echo ">> building binary file"
	@go build \
		-ldflags "-X 'github.com/nthienan/aws-dynamodb-cache-lambda-extension/internal/version.Version=$(VERSION)' \
			-X 'github.com/nthienan/aws-dynamodb-cache-lambda-extension/internal/version.Revision=$(GIT_COMMIT)' \
			-X 'github.com/nthienan/aws-dynamodb-cache-lambda-extension/internal/version.Branch=$(GIT_BRANCH)' \
			-X 'github.com/nthienan/aws-dynamodb-cache-lambda-extension/internal/version.BuildUser=$(USER)' \
			-X 'github.com/nthienan/aws-dynamodb-cache-lambda-extension/internal/version.BuildDate=$(DATE)'" \
		-o $(APP_BIN) $(APP_SRC)


.PHONY: go-run
go-run: go-vet
	@go run $(APP_SRC)


.PHONY: go-fmt
go-fmt:
	@echo ">> checking code style"
	@fmtRes=$$($(GOFMT) -d $$(find . -path ./vendor -prune -o -name '*.go' -print)); \
	if [ -n "$${fmtRes}" ]; then \
		echo "go fmt checking failed!"; echo "$${fmtRes}"; echo; \
		exit 1; \
	fi


.PHONY: go-vet
go-vet:
	@echo ">> vetting code"
	$(GO) vet $(GOOPTS) ./...
