ORG := stelligent
PACKAGE := mu
TARGET_OS := linux windows darwin

###
BRANCH := $(or $(TRAVIS_BRANCH), $(shell git rev-parse --abbrev-ref HEAD))
IS_MASTER := $(filter master, $(BRANCH))
VERSION := $(shell cat VERSION)$(if $(IS_MASTER),,-$(BRANCH))
ARCH := $(shell go env GOARCH)
BUILD_FILES = $(foreach os, $(TARGET_OS), .release/$(PACKAGE)-$(os)-$(ARCH))
UPLOAD_FILES = $(foreach os, $(TARGET_OS), $(PACKAGE)-$(os)-$(ARCH))
GOLDFLAGS = "-X main.version=$(VERSION)"
TAG_VERSION = v$(VERSION)

default: build

deps:
	@echo "=== preparing $(VERSION) from $(BRANCH) ==="
	go get "github.com/jteeuwen/go-bindata/..."
	go get -t -d -v ./...
	go generate ./...

lint:
	@echo "=== linting ==="
	go vet ./...
	go get "github.com/golang/lint/golint"
	golint -set_exit_status ./...

test: lint
	@echo "=== testing ==="
	go test -cover ./...

build: $(BUILD_FILES)

$(BUILD_FILES):
	@echo "=== building $(VERSION) - $@ ==="
	mkdir -p .release
	GOOS=$(word 2,$(subst -, ,$@)) GOARCH=$(word 3,$(subst -, ,$@)) go build -ldflags=$(GOLDFLAGS) -o '$@'

release-setup:
ifneq ($(GITHUB_TOKEN),)
	git config credential.helper "store --file=.git/credentials"
	echo "https://${GITHUB_TOKEN}:@github.com" > .git/credentials
endif
	go get "github.com/aktau/github-release"

release-clean:
ifeq ($(IS_MASTER),)
	@echo "=== clearing old release $(VERSION) ==="
	github-release info -u $(ORG) -r $(PACKAGE) -t $(TAG_VERSION) && github-release delete -u $(ORG) -r $(PACKAGE) -t $(TAG_VERSION) || echo "No release to cleanup"
	git push --delete origin $(TAG_VERSION) || echo "No tag to delete"
endif

release-create: release-clean
	@echo "=== creating pre-release $(VERSION) ==="
	git tag -f $(TAG_VERSION)
	git push origin $(TAG_VERSION)
	echo "waiting for dust to settle..." && sleep 5
	github-release release -u $(ORG) -r $(PACKAGE) -t $(TAG_VERSION) -p

$(TARGET_OS): release-create
	@echo "=== uploading $@ ==="
	github-release upload -u $(ORG) -r $(PACKAGE) -t $(TAG_VERSION) -n "$(PACKAGE)-$@-$(ARCH)" -f ".release/$(PACKAGE)-$@-$(ARCH)"

dev-release: $(TARGET_OS)

release: dev-release
ifneq ($(IS_MASTER),)
	@echo "=== releasing $(VERSION) ==="
	github-release edit -u $(ORG) -r $(PACKAGE) -t $(TAG_VERSION)

	github-release info -u $(ORG) -r $(PACKAGE) -t $(TAG_VERSION)-develop && github-release delete -u $(ORG) -r $(PACKAGE) -t $(TAG_VERSION)-develop || echo "No pre-release to cleanup"
	git push --delete origin $(TAG_VERSION)-develop || echo "No pre-release tag to delete"
endif

clean:
	@echo "=== cleaning ==="
	rm -rf .release

.PHONY: default lint test build deps clean release-clean release-create dev-release release $(UPLOAD_FILES) $(TARGET_OS)
