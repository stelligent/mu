ORG := stelligent
PACKAGE := mu
SRC_PACKAGES = provider workflows cli common templates e2e
SNAPSHOT_SUFFIX := develop

###
SRC_FILES = $(foreach pkg, $(SRC_PACKAGES), ./$(pkg)/...) .
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
LATEST_VERSION := $(shell git tag -l --sort=creatordate | grep "^v[0-9]*.[0-9]*.[0-9]*$$" | tail -1 | cut -c 2-)

ifdef CIRCLE_TAG
VERSION = $(CIRCLE_TAG:v%=%)
else
ifeq "$(shell git tag -l v$(LATEST_VERSION) --points-at HEAD)" "v$(LATEST_VERSION)"
### latest tag points to current commit, this is a release build
VERSION ?= $(LATEST_VERSION)
else
### latest tag points to prior commit, this is a snapshot build
MAJOR_VERSION := $(word 1, $(subst ., ,$(LATEST_VERSION)))
MINOR_VERSION := $(word 2, $(subst ., ,$(LATEST_VERSION)))
PATCH_VERSION := $(word 3, $(subst ., ,$(LATEST_VERSION)))
VERSION ?= $(MAJOR_VERSION).$(MINOR_VERSION).$(shell echo $$(( $(PATCH_VERSION) + 1)) )-$(SNAPSHOT_SUFFIX)
endif
endif
IS_SNAPSHOT = $(if $(findstring -, $(VERSION)),true,false)
TAG_VERSION = v$(VERSION)

ARCH := $(shell go env GOARCH)
OS := $(shell go env GOOS)
GEM := $(shell command -v gem 2> /dev/null)

export PATH := $(GOPATH)/bin:$(PATH)


default: all

info:
	@echo "VERSION:$(VERSION) IS_SNAPSHOT:$(IS_SNAPSHOT) BRANCH:$(BRANCH) LATEST_VERSION:$(LATEST_VERSION)"

deps:
	@echo "=== dependencies ==="
	go get "github.com/golang/dep/cmd/dep"
	@dep ensure -vendor-only -v
	@git apply -p1 go-git.v4.patch

fmt:
	@echo "=== formatting ==="
	@go fmt $(SRC_FILES)

lint: fmt
	@echo "=== linting ==="
	@go vet $(SRC_FILES)

	@go get -u golang.org/x/lint/golint
	@echo $(SRC_FILES) | xargs -n1 golint -set_exit_status

gen:
	@echo "=== generating ==="
	@go get "github.com/gobuffalo/packr/packr"
	@go generate $(SRC_FILES)

nag:
	@echo "=== cfn_nag ==="
	@gem list -i cfn-nag --silent || gem install cfn-nag || gem install cfn-nag --user-install

	@mkdir -p .cfn_nag
	@grep -l AWSTemplateFormatVersion: templates/assets/cloudformation/*.yml | while read -r line; do \
		filename=`basename $$line` ;\
		grep -v '{{' $$line > .cfn_nag/$$filename ;\
		output=`cfn_nag_scan --input-path .cfn_nag/$$filename 2>&1` ;\
		if [ $$? -ne 0 ]; then \
			echo $$filename ;\
			echo "$$output\n" ;\
		fi ;\
	done | grep ".*" ;\
    if [ $$? -eq 0 ]; then \
    	exit 1 ;\
    fi

cyclo:
	@echo "=== cyclomatic complexity ==="
	@go get "github.com/fzipp/gocyclo"
	@gocyclo -over 15 $(SRC_PACKAGES)
	@gocyclo -over 12 $(SRC_PACKAGES) || echo "WARNING: cyclomatic complexity is high"

ifdef GEM
test: nag
endif

test: info lint gen cyclo
	@echo "=== testing ==="
ifneq ($(CIRCLE_WORKING_DIRECTORY),)
	@mkdir -p $(CIRCLE_WORKING_DIRECTORY)/test-results/unit
	@go get "github.com/jstemmer/go-junit-report"
	@bash -co pipefail 'go test -v -cover $(filter-out ./e2e/..., $(SRC_FILES)) -short | go-junit-report > $(CIRCLE_WORKING_DIRECTORY)/test-results/unit/report.xml'
else
	@go test -cover $(filter-out ./e2e/..., $(SRC_FILES)) -short
endif


build: info gen
	@go get github.com/goreleaser/goreleaser
	$(eval export SNAPSHOT_VERSION=$(VERSION))
	@goreleaser --snapshot --rm-dist

install: build
	@echo "=== installing $(PACKAGE)-$(OS)-$(ARCH) ==="
	@cp dist/$(OS)_$(ARCH)/$(PACKAGE) /usr/local/bin/mu
	@chmod 755 /usr/local/bin/mu
	@mu -v

stage: fmt build
	@echo "=== staging to S3 bucket ==="
	@export BUCKET_NAME=mu-staging-$$(aws sts get-caller-identity --output text --query 'Account') ;\
	aws s3 mb s3://$$BUCKET_NAME || echo "bucket exists" ;\
	aws s3 website --index-document index.html s3://$$BUCKET_NAME ;\
	aws s3 cp dist/linux_amd64/mu s3://$$BUCKET_NAME/$(TAG_VERSION)/$(PACKAGE)-linux-amd64 --acl public-read ;\
	echo https://$$BUCKET_NAME.s3.amazonaws.com

keypair:
	@aws ec2 describe-key-pairs --key-names mu-e2e > /dev/null 2>&1; \
	if [ $$? -ne 0 ]; then \
		echo "=== creating keypair ==="; \
		aws ec2 create-key-pair --key-name mu-e2e --query "KeyMaterial" --output text > ~/.ssh/mu-e2e-$$(aws sts get-caller-identity --output text --query 'Account').pem; \
		chmod 600 ~/.ssh/mu-e2e-$$(aws sts get-caller-identity --output text --query 'Account').pem; \
	fi;

e2e: gen stage keypair
	@echo "=== e2e testing ==="
	@MU_VERSION=$(VERSION) MU_BASEURL=https://mu-staging-$$(aws sts get-caller-identity --output text --query 'Account').s3.amazonaws.com go test -v ./e2e -timeout 60m

e2e_basic: gen stage keypair
	@echo "=== e2e-basic testing ==="
	@INCLUDE_TESTS=e2e-basic MU_VERSION=$(VERSION) MU_BASEURL=https://mu-staging-$$(aws sts get-caller-identity --output text --query 'Account').s3.amazonaws.com go test -v ./e2e -timeout 60m

e2e_ec2: gen stage keypair
	@echo "=== e2e-ec2 testing ==="
	@INCLUDE_TESTS=e2e-ec2 MU_VERSION=$(VERSION) MU_BASEURL=https://mu-staging-$$(aws sts get-caller-identity --output text --query 'Account').s3.amazonaws.com go test -v ./e2e -timeout 60m

e2e_eks: gen stage keypair
	@echo "=== e2e-eks testing ==="
	@INCLUDE_TESTS=e2e-eks MU_VERSION=$(VERSION) MU_BASEURL=https://mu-staging-$$(aws sts get-caller-identity --output text --query 'Account').s3.amazonaws.com go test -v ./e2e -timeout 60m

e2e_fargate: gen stage keypair
	@echo "=== e2e-fargate testing ==="
	@INCLUDE_TESTS=e2e-fargate MU_VERSION=$(VERSION) MU_BASEURL=https://mu-staging-$$(aws sts get-caller-identity --output text --query 'Account').s3.amazonaws.com go test -v ./e2e -timeout 60m

check_github_token:
ifndef GITHUB_TOKEN
	@echo GITHUB_TOKEN is undefined
	@echo Create one at https://github.com/settings/tokens
	@exit 1
endif


changelog: check_github_token
	@echo "=== generating changelog ==="
	@rm -f CHANGELOG.md
	@go get github.com/Songmu/ghch/cmd/ghch
ifeq ($(IS_SNAPSHOT),true)
	@ghch --format=markdown --from=v$(LATEST_VERSION) -w
else
	@ghch --format=markdown --latest -w
endif

github_release: check_github_token gen changelog
	@echo "=== generating github release '$(TAG_VERSION)' ==="
	@go get github.com/goreleaser/goreleaser
ifeq ($(IS_SNAPSHOT),true)
	@go get github.com/aktau/github-release
	@github-release delete -u stelligent -r mu -t $(TAG_VERSION) || echo "already deleted"
endif
	@goreleaser --rm-dist --release-notes CHANGELOG.md
ifeq ($(IS_SNAPSHOT),true)
	@github-release edit -u stelligent -r mu -t $(TAG_VERSION) -p -d - < CHANGELOG.md
endif

formula:
	rm -rf homebrew-tap
	git clone git@github.com:stelligent/homebrew-tap.git

	$(eval MAC_URL := https://github.com/stelligent/mu/releases/download/$(TAG_VERSION)/mu-darwin-amd64)
	$(eval MAC_SHA256 := $(shell curl -L -s $(MAC_URL) | shasum -a 256 | cut -d' ' -f1))
	$(eval LINUX_URL := https://github.com/stelligent/mu/releases/download/$(TAG_VERSION)/mu-linux-amd64)
	$(eval LINUX_SHA256 := $(shell curl -L -s $(LINUX_URL) | shasum -a 256 | cut -d' ' -f1))

    # Update formula in mu-cli.rb
ifeq ($(IS_SNAPSHOT),true)
	$(eval BREW_BRANCH := $(SNAPSHOT_SUFFIX))
else
	$(eval BREW_BRANCH := 'master')
endif

ifeq ($(OS),darwin)
	sed -i "" 's|.*\( # The MacOS '$(BREW_BRANCH)' url\)|    url "'$(MAC_URL)'"\1|g '  homebrew-tap/Formula/mu-cli.rb
	sed -i "" 's|.*\( # The MacOS '$(BREW_BRANCH)' sha256sum\)|    sha256 "'$(MAC_SHA256)'"\1|g' homebrew-tap/Formula/mu-cli.rb
	sed -i "" 's|.*\( # The Linux '$(BREW_BRANCH)' url\)|    url "'$(LINUX_URL)'"\1|g' homebrew-tap/Formula/mu-cli.rb
	sed -i "" 's|.*\( # The Linux '$(BREW_BRANCH)' sha256sum\)|    sha256 "'$(LINUX_SHA256)'"\1|g' homebrew-tap/Formula/mu-cli.rb
	sed -i "" 's|\(\s*version\).*\( # The '$(BREW_BRANCH)' version\)|\1 "'$(VERSION)'"\2|g' homebrew-tap/Formula/mu-cli.rb
else
	sed -i"" 's|.*\( # The MacOS '$(BREW_BRANCH)' url\)|    url "'$(MAC_URL)'"\1|g '  homebrew-tap/Formula/mu-cli.rb
	sed -i"" 's|.*\( # The MacOS '$(BREW_BRANCH)' sha256sum\)|    sha256 "'$(MAC_SHA256)'"\1|g' homebrew-tap/Formula/mu-cli.rb
	sed -i"" 's|.*\( # The Linux '$(BREW_BRANCH)' url\)|    url "'$(LINUX_URL)'"\1|g' homebrew-tap/Formula/mu-cli.rb
	sed -i"" 's|.*\( # The Linux '$(BREW_BRANCH)' sha256sum\)|    sha256 "'$(LINUX_SHA256)'"\1|g' homebrew-tap/Formula/mu-cli.rb
	sed -i"" 's|\(\s*version\).*\( # The '$(BREW_BRANCH)' version\)|\1 "'$(VERSION)'"\2|g' homebrew-tap/Formula/mu-cli.rb
endif

	git -C homebrew-tap add Formula/mu-cli.rb
	git -C homebrew-tap commit -m "auto updated the mu-cli formula for version $(TAG_VERSION) branch $(BRANCH)"
	git -C homebrew-tap push

release: info github_release formula

image:
	docker build -t stelligent/mu:$(VERSION) .

clean:
	@echo "=== cleaning ==="
	rm -rf vendor
	rm -rf .cfn_nag
	rm -rf dist
	rm -f templates/*-packr.go

all: clean deps test build

depromote: info check_github_token
	@echo "Depromoting $(LATEST_VERSION)"
	@github-release delete -u stelligent -r mu -t v$(LATEST_VERSION)
	git tag --delete v$(LATEST_VERSION)
	@git push -d origin v$(LATEST_VERSION)

promote: info
ifeq (false,$(IS_SNAPSHOT))
	@echo "Unable to promote a non-snapshot"
	@exit 1
endif
ifneq ($(shell git status -s),)
	@echo "Unable to promote a dirty workspace"
	@exit 1
endif
	$(eval NEW_VERSION := $(word 1,$(subst -, , $(TAG_VERSION))))
	@git tag -a -m "releasing $(NEW_VERSION)" $(NEW_VERSION)
	@git push origin $(NEW_VERSION)

promote-develop:
ifneq ($(shell git status -s),)
	@echo "Unable to promote a dirty workspace"
	@exit 1
endif
	@echo "=== creating tag '$(TAG_VERSION)' ==="
	@git tag --force -a -m "releasing $(TAG_VERSION)" $(TAG_VERSION)
	@git push --force origin $(TAG_VERSION)

.PHONY: default all lint test e2e build deps gen clean release install keypair stage promote formula github_release changelog tag_release check_github_token
