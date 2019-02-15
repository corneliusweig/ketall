# Copyright 2019 Cornelius Weig
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

PROJECT   ?= ketall
REPOPATH  ?= github.com/corneliusweig/$(PROJECT)
PLATFORMS ?= linux windows darwin

TARGETS   := $(patsubst %,$(PROJECT)-%-amd64,$(PLATFORMS))
CHECKSUMS := $(patsubst %,%.sha256,$(TARGETS))

VERSION_PACKAGE := $(REPOPATH)/pkg/version
COMMIT          := $(shell git rev-parse HEAD)
VERSION         ?= $(shell git describe --always --tags --dirty)

GO_LDFLAGS :="
GO_LDFLAGS += -X $(VERSION_PACKAGE).version=$(VERSION)
GO_LDFLAGS += -X $(VERSION_PACKAGE).buildDate=$(shell date +'%Y-%m-%dT%H:%M:%SZ')
GO_LDFLAGS += -X $(VERSION_PACKAGE).gitCommit=$(COMMIT)
GO_LDFLAGS +="

GO_FILES  := $(shell find . -type f -name '*.go')

test:
	GO111MODULE=on go test ./...

all: $(TARGETS)

dev: ketall-linux-amd64
	@mv $< $(PROJECT)

$(PROJECT)-%-amd64: $(GO_FILES)
	GO111MODULE=on GOARCH=amd64 CGO_ENABLED=0 GOOS=$* go build -ldflags $(GO_LDFLAGS) -o $@ main.go

.PHONY: deploy
deploy: $(TARGETS)
	for x in $^; do sha256sum $$x > $${x}.sha256;done
	git archive --format=tar.gz HEAD > $(VERSION).tar.gz

.PHONY: dist
dist: $(TARGETS)
	tar czf $(PROJECT).tar.gz $(TARGETS) Makefile plugin.yaml

.PHONY: clean
clean:
	$(RM) $(TARGETS) $(CHECKSUMS) $(VERSION).tar.gz
