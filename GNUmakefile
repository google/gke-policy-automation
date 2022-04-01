#Copyright 2022 Google LLC
#
#Licensed under the Apache License, Version 2.0 (the "License");
#you may not use this file except in compliance with the License.
#You may obtain a copy of the License at
#
#    https://www.apache.org/licenses/LICENSE-2.0
#
#Unless required by applicable law or agreed to in writing, software
#distributed under the License is distributed on an "AS IS" BASIS,
#WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#See the License for the specific language governing permissions and
#limitations under the License.

GOCMD		=go
TEST		?=$$(go list ./... |grep -v 'vendor')
BINARY		=gke-policy
APP_MODULE	=github.com/google/gke-policy-automation/internal/app
GOFMT_FILES	?=$$(find . -name '*.go' |grep -v vendor)
COMMIT_SHA  =$(shell g rev-parse --short HEAD)
LDFLAGS     =-s -w

ifneq (, $(shell git 2>/dev/null))
	COMMIT_SHA	=$(shell git rev-parse --short HEAD)
	LDFLAGS		+= -X ${APP_MODULE}.Version=git-${COMMIT_SHA}
endif

default: clean build test

all: default

build:
	${GOCMD} build -ldflags "${LDFLAGS}" -o ${BINARY}

test: fmtcheck
	echo $(TEST) | \
		xargs -t ${GOCMD} test -v -cover

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

clean:
	${GOCMD} clean

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@./scripts/gofmtcheck.sh

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

.PHONY: build test clean fmt fmtcheck vet