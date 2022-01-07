# File: Makefile
# Author: Stalker-lee
# DateTime: 2022-01-06 18:00:12
#
# Description:  Makefile artifact from source code
# Usage:
#       make help|build-linux|build-darwin|build-windows|build-docker

# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /usr/bin/env bash

# specify go env parameters
GO           ?= go
GOFMT        ?= $(GO)fmt
FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
GOOPTS       ?=
GOHOSTOS     ?= $(shell $(GO) env GOHOSTOS)
GOHOSTARCH   ?= $(shell $(GO) env GOHOSTARCH)

GO_VERSION        ?= $(shell $(GO) version)
GO_VERSION_NUMBER ?= $(word 3, $(GO_VERSION))
PRE_GO_111        ?= $(shell echo $(GO_VERSION_NUMBER) | grep -E 'go1\.(10|[0-9])\.')

GOVENDOR :=
GO111MODULE :=

ifeq (, $(PRE_GO_111))
	ifneq (,$(wildcard go.mod))
		# Enforce Go modules support just in case the directory is inside GOPATH (and for Travis CI).
		GO111MODULE := on

		ifneq (,$(wildcard vendor))
			# Always use the local vendor/ directory to satisfy the dependencies.
			GOOPTS := $(GOOPTS) -mod=vendor
		endif
	endif
else
	ifneq (,$(wildcard go.mod))
		ifneq (,$(wildcard vendor))
$(warning This repository requires Go >= 1.11 because of Go modules)
$(warning Some recipes may not work as expected as the current Go runtime is '$(GO_VERSION_NUMBER)')
		endif
	else
		# This repository isn't using Go modules (yet).
		GOVENDOR := $(FIRST_GOPATH)/bin/govendor
	endif
endif

ifeq (arm, $(GOHOSTARCH))
	GOHOSTARM ?= $(shell GOARM= $(GO) env GOARM)
	GO_BUILD_PLATFORM ?= $(GOHOSTOS)-$(GOHOSTARCH)v$(GOHOSTARM)
else
	GO_BUILD_PLATFORM ?= $(GOHOSTOS)-$(GOHOSTARCH)
endif

# 关于 docker 设置
DOCKER ?= $(shell command -v "docker")

# The name of the binary component
NAME_OF_ARTIFACT := owl-engine

# The directory where the artifact is stored
STORE_ARTIFACT_DIR := $(CURDIR)/owl-engine

# The storage directory of the generated script
STORE_GENERATE_SCRIPTS_DIR := $(CURDIR)/docs/shell

.PHONY: fmt
fmt:
	@go list -f {{.Dir}} ./... | xargs gofmt -w -s -d

.PHONY: test
test: ## execute unit test for go
	@go test



.PHONY: fmt test build-darwin
build-darwin: ## build darwin artifact on x86 platform
	@bash $(STORE_GENERATE_SCRIPTS_DIR)/darwin.sh $(NAME_OF_ARTIFACT) $(CURDIR) $(STORE_ARTIFACT_DIR) $(STORE_GENERATE_SCRIPTS_DIR)

.PHONY: fmt test build-linux
build-linux: ## build linux artifact on x86 platform
	@bash $(STORE_GENERATE_SCRIPTS_DIR)/linux.sh $(NAME_OF_ARTIFACT) $(CURDIR) $(STORE_ARTIFACT_DIR) $(STORE_GENERATE_SCRIPTS_DIR)

.PHONY: fmt test build-windows
build-windows: ## build windows artifact on x86 platform
	@bash $(STORE_GENERATE_SCRIPTS_DIR)/windows.sh $(NAME_OF_ARTIFACT) $(CURDIR) $(STORE_ARTIFACT_DIR) $(STORE_GENERATE_SCRIPTS_DIR)

.PHONY: fmt test build-docker
build-docker: ## build docker images
	# docker 是否安装
	if [[ -n `command -v docker` ]]; then                 					    										\
		is_running=$$(docker info | grep -E "Is the docker daemon running?") ;  											\
		if [[ -z $$is_running ]]; then 									   												\
			containerID=$$(docker images -q --filter reference=$(NAME_OF_ARTIFACT) | sort | uniq) ; 						\
			if [[ -n $$is_exist ]]; then																					\
				docker rmi -f $$containerID ;																				\
				echo "owl-engine image already exists. delete it now" ;														\
			fi ;																											\
																															\
			docker build --build-arg MODE=$(MODE) -t $(NAME_OF_ARTIFACT):v1.0.0 $(CURDIR) ;								\
			if [[ $$? -eq 0 ]]; then																						\
				cd $(CURDIR) ;																								\
				docker save -o $(NAME_OF_ARTIFACT)_v1.0.0.tar $(NAME_OF_ARTIFACT):v1.0.0 ; 								\
				echo "$(CURDIR)/$(NAME_OF_ARTIFACT)_v1.0.0.tar image was successfully built, now you can distribute it!" ; \
			fi ; 																											\
		else																												\
			echo "Is the docker daemon running?" ;													        				\
		fi ;																    											\
	fi

.PHONY: help
help: ## show make targets
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\n",sprintf("\n%22c"," "), $$2);printf " \033[36m%-20s\033[0m  %s\n", $$1, $$2}' $(MAKEFILE_LIST)
