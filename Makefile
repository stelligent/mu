ORG := stelligent
PACKAGE := mu
TARGET_OS := linux windows darwin

###
BRANCH := $(or $(TRAVIS_BRANCH), $(shell git rev-parse --abbrev-ref HEAD))
IS_MASTER := $(filter master, $(BRANCH))
VERSION := $(shell cat VERSION)$(if $(IS_MASTER),,-$(BRANCH))
SRC_FILES = $(shell glide nv)
ARCH := $(shell go env GOARCH)
OS := $(shell go env GOOS)
BUILD_DIR = $(if $(CIRCLE_ARTIFACTS),$(CIRCLE_ARTIFACTS),.release)
BUILD_FILES = $(foreach os, $(TARGET_OS), $(BUILD_DIR)/$(PACKAGE)-$(os)-$(ARCH))
UPLOAD_FILES = $(foreach os, $(TARGET_OS), $(PACKAGE)-$(os)-$(ARCH))
GOLDFLAGS = "-X main.version=$(VERSION)"
TAG_VERSION = v$(VERSION)

default: all

deps:
	@echo "=== preparing $(VERSION) from $(BRANCH) ==="
	go get "github.com/jteeuwen/go-bindata/..."
	go get "github.com/golang/lint/golint"
	go get "github.com/jstemmer/go-junit-report"
	go get "github.com/aktau/github-release"
	#go get -t -d -v $(SRC_FILES)
	glide install

gen:
	go generate $(SRC_FILES)

lint: fmt
	@echo "=== linting ==="
	go vet $(SRC_FILES)
	glide novendor | xargs -n1 golint -set_exit_status

test: lint gen
	@echo "=== testing ==="
ifneq ($(CIRCLE_TEST_REPORTS),)
	mkdir -p $(CIRCLE_TEST_REPORTS)/unit
	go test -v -cover $(SRC_FILES) | go-junit-report > $(CIRCLE_TEST_REPORTS)/unit/report.xml
else
	go test -cover $(SRC_FILES)
endif


build: gen $(BUILD_FILES)

$(BUILD_FILES):
	@echo "=== building $(VERSION) - $@ ==="
	mkdir -p $(BUILD_DIR)
	GOOS=$(word 2,$(subst -, ,$(notdir $@))) GOARCH=$(word 3,$(subst -, ,$(notdir $@))) go build -ldflags=$(GOLDFLAGS) -o '$@'

install: build
	@echo "=== building $(VERSION) - $(PACKAGE)-$(OS)-$(ARCH) ==="
	cp $(BUILD_DIR)/$(PACKAGE)-$(OS)-$(ARCH) /usr/local/bin/mu
	chmod 755 /usr/local/bin/mu

release-clean:
ifeq ($(IS_MASTER),)
	@echo "=== clearing old release $(VERSION) ==="
	github-release info -u $(ORG) -r $(PACKAGE) -t $(TAG_VERSION) && github-release delete -u $(ORG) -r $(PACKAGE) -t $(TAG_VERSION) || echo "No release to cleanup"
	git push --delete origin $(TAG_VERSION) || echo "No tag to delete"
endif

release-create: build release-clean
	@echo "=== creating pre-release $(VERSION) ==="
	git tag -f $(TAG_VERSION)
	git push origin $(TAG_VERSION)
	echo "waiting for dust to settle..." && sleep 5
	github-release release -u $(ORG) -r $(PACKAGE) -t $(TAG_VERSION) -p

$(TARGET_OS): release-create
	@echo "=== uploading $@ ==="
	github-release upload -u $(ORG) -r $(PACKAGE) -t $(TAG_VERSION) -n "$(PACKAGE)-$@-$(ARCH)" -f "$(BUILD_DIR)/$(PACKAGE)-$@-$(ARCH)"

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
	rm -rf $(BUILD_DIR)
	rm -rf vendor

all: clean deps test build

fmt:
	@echo "=== formatting ==="
	go fmt $(SRC_FILES)


.PHONY: default all lint test build deps gen clean release-clean release-create dev-release release install $(UPLOAD_FILES) $(TARGET_OS)
