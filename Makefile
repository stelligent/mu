ORG := stelligent
PACKAGE := mu
TARGET_OS := linux windows darwin
SRC_PACKAGES = provider workflows cli common templates e2e 

###
BRANCH := $(or $(TRAVIS_BRANCH), $(shell git rev-parse --abbrev-ref HEAD))
IS_MASTER := $(filter master, $(BRANCH))
VERSION := $(shell cat VERSION)$(if $(IS_MASTER),,-$(BRANCH))
SRC_FILES = $(foreach pkg, $(SRC_PACKAGES), ./$(pkg)/...) .
ARCH := $(shell go env GOARCH)
OS := $(shell go env GOOS)
BUILD_DIR = .release
BUILD_DIR = $(abspath $(if $(CIRCLE_WORKING_DIRECTORY),$(CIRCLE_WORKING_DIRECTORY)/artifacts,.release))
BUILD_FILES = $(foreach os, $(TARGET_OS), $(BUILD_DIR)/$(PACKAGE)-$(os)-$(ARCH))
UPLOAD_FILES = $(foreach os, $(TARGET_OS), $(PACKAGE)-$(os)-$(ARCH))
GOLDFLAGS = "-X main.version=$(VERSION)"
TAG_VERSION = v$(VERSION)

export PATH := $(GOPATH)/bin:$(PATH)

default: all

deps:
	@echo "=== preparing $(VERSION) from $(BRANCH) ==="
	go get "github.com/golang/dep/cmd/dep"
	go get "github.com/jteeuwen/go-bindata/..."
	go get "github.com/golang/lint/golint"
	go get "github.com/jstemmer/go-junit-report"
	go get "github.com/aktau/github-release"
	go get "github.com/fzipp/gocyclo"
	dep ensure
	patch -p1 < go-git.v4.patch

ifeq ($(CIRCLECI),true)
	gem install cfn-nag
else
	gem list | grep cfn-nag || sudo gem install cfn-nag
endif

gen:
	go generate $(SRC_FILES)

lint: fmt
	@echo "=== linting ==="
	go vet $(SRC_FILES)
	echo $(SRC_FILES) | xargs -n1 golint -set_exit_status

nag:
	@echo "=== cfn_nag ==="
	@mkdir -p $(BUILD_DIR)/cfn_nag
	@grep -l AWSTemplateFormatVersion: templates/assets/*.yml | while read -r line; do \
		filename=`basename $$line` ;\
		grep -v '{{' $$line > $(BUILD_DIR)/cfn_nag/$$filename ;\
		output=`cfn_nag_scan --input-path $(BUILD_DIR)/cfn_nag/$$filename 2>&1` ;\
		if [ $$? -ne 0 ]; then \
			echo "$$output\n" ;\
		fi ;\
	done | grep ".*" ;\
    if [ $$? -eq 0 ]; then \
    	exit 1 ;\
    fi

cyclo:
	@echo "=== cyclomatic complexity ==="
	@gocyclo -over 30 $(SRC_PACKAGES)
	@gocyclo -over 15 $(SRC_PACKAGES) || echo "WARNING: cyclomatic complexity is high"
	
test: lint gen nag cyclo
	@echo "=== testing ==="
ifneq ($(CIRCLE_WORKING_DIRECTORY),)
	mkdir -p $(CIRCLE_WORKING_DIRECTORY)/test-results/unit
	bash -co pipefail 'go test -v -cover $(filter-out ./e2e/..., $(SRC_FILES)) -short | go-junit-report > $(CIRCLE_WORKING_DIRECTORY)/test-results/unit/report.xml'
else
	go test -cover $(filter-out ./e2e/..., $(SRC_FILES)) -short
endif

e2e: gen stage keypair
	@echo "=== e2e testing ==="
	MU_VERSION=$(VERSION) MU_BASEURL=https://mu-staging-$$(aws sts get-caller-identity --output text --query 'Account').s3.amazonaws.com go test -v ./e2e -timeout 60m

build: gen $(BUILD_FILES)

$(BUILD_FILES):
	@echo "=== building $(VERSION) - $@ ==="
	mkdir -p $(BUILD_DIR)
	GOOS=$(word 2,$(subst -, ,$(notdir $@))) GOARCH=$(word 3,$(subst -, ,$(notdir $@))) go build -ldflags=$(GOLDFLAGS) -o '$@'

install: gen $(BUILD_DIR)/$(PACKAGE)-$(OS)-$(ARCH)
	@echo "=== installing $(VERSION) - $(PACKAGE)-$(OS)-$(ARCH) ==="
	cp $(BUILD_DIR)/$(PACKAGE)-$(OS)-$(ARCH) /usr/local/bin/mu
	chmod 755 /usr/local/bin/mu

keypair:
	@aws ec2 describe-key-pairs --key-names mu-e2e > /dev/null 2>&1; \
	if [ $$? -ne 0 ]; then \
		echo "=== creating keypair ==="; \
		aws ec2 create-key-pair --key-name mu-e2e --query "KeyMaterial" --output text > ~/.ssh/mu-e2e-$$(aws sts get-caller-identity --output text --query 'Account').pem; \
		chmod 600 ~/.ssh/mu-e2e-$$(aws sts get-caller-identity --output text --query 'Account').pem; \
	fi;

stage: fmt $(BUILD_DIR)/$(PACKAGE)-linux-$(ARCH)
	@echo "=== staging to S3 bucket ==="
	@export BUCKET_NAME=mu-staging-$$(aws sts get-caller-identity --output text --query 'Account') ;\
	aws s3 mb s3://$$BUCKET_NAME || echo "bucket exists" ;\
	aws s3 website --index-document index.html s3://$$BUCKET_NAME ;\
	aws s3 sync $(BUILD_DIR) s3://$$BUCKET_NAME/v$(VERSION)/ --acl public-read --exclude "*" --include "$(PACKAGE)-linux-*" ;\
	echo https://$$BUCKET_NAME.s3.amazonaws.com

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

formula:
	rm -rf homebrew-tap
	git clone git@github.com:stelligent/homebrew-tap.git

	$(eval MAC_URL := https://github.com/stelligent/mu/releases/download/$(TAG_VERSION)/mu-darwin-amd64)
	$(eval MAC_SHA256 := $(shell curl -L -s $(MAC_URL) | shasum -a 256 | cut -d' ' -f1))
	$(eval LINUX_URL := https://github.com/stelligent/mu/releases/download/$(TAG_VERSION)/mu-linux-amd64)
	$(eval LINUX_SHA256 := $(shell curl -L -s $(LINUX_URL) | shasum -a 256 | cut -d' ' -f1))

    # Update formula in mu-cli.rb
ifeq ($(OS),darwin)
	sed -i "" 's|.*\( # The MacOS '$(BRANCH)' url\)|    url "'$(MAC_URL)'"\1|g '  homebrew-tap/Formula/mu-cli.rb
	sed -i "" 's|.*\( # The MacOS '$(BRANCH)' sha256sum\)|    sha256 "'$(MAC_SHA256)'"\1|g' homebrew-tap/Formula/mu-cli.rb
	sed -i "" 's|.*\( # The Linux '$(BRANCH)' url\)|    url "'$(LINUX_URL)'"\1|g' homebrew-tap/Formula/mu-cli.rb
	sed -i "" 's|.*\( # The Linux '$(BRANCH)' sha256sum\)|    sha256 "'$(LINUX_SHA256)'"\1|g' homebrew-tap/Formula/mu-cli.rb
	sed -i "" 's|\(\s*version\).*\( # The '$(BRANCH)' version\)|\1 "'$(VERSION)'"\2|g' homebrew-tap/Formula/mu-cli.rb
else
	sed -i"" 's|.*\( # The MacOS '$(BRANCH)' url\)|    url "'$(MAC_URL)'"\1|g '  homebrew-tap/Formula/mu-cli.rb
	sed -i"" 's|.*\( # The MacOS '$(BRANCH)' sha256sum\)|    sha256 "'$(MAC_SHA256)'"\1|g' homebrew-tap/Formula/mu-cli.rb
	sed -i"" 's|.*\( # The Linux '$(BRANCH)' url\)|    url "'$(LINUX_URL)'"\1|g' homebrew-tap/Formula/mu-cli.rb
	sed -i"" 's|.*\( # The Linux '$(BRANCH)' sha256sum\)|    sha256 "'$(LINUX_SHA256)'"\1|g' homebrew-tap/Formula/mu-cli.rb
	sed -i"" 's|\(\s*version\).*\( # The '$(BRANCH)' version\)|\1 "'$(VERSION)'"\2|g' homebrew-tap/Formula/mu-cli.rb
endif

	git -C homebrew-tap add Formula/mu-cli.rb
	git -C homebrew-tap commit -m "auto updated the mu-cli formula for version $(TAG_VERSION) branch $(BRANCH)"
	git -C homebrew-tap push

clean:
	@echo "=== cleaning ==="
	rm -rf $(BUILD_DIR)
	rm -rf vendor

all: clean deps test build

fmt:
	@echo "=== formatting ==="
	go fmt $(SRC_FILES)

changelog:
	github_changelog_generator -u stelligent -p mu -t $(GITHUB_TOKEN)


.PHONY: default all lint test e2e build deps gen clean release-clean release-create dev-release release install $(UPLOAD_FILES) $(BUILD_FILES) $(TARGET_OS) keypair stage
