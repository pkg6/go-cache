# Copyright 2023
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

# Test all Go packages and enable data competition detection
.PHONY: tr
tr:
	@echo "go test -race ..."
	@go test -race ./...

# Run the setup.sh script to set up the environment
.PHONY: setup
setup:
	@echo "sh ./script/setup.sh ..."
	@sh ./script/setup.sh

#GolangCI Litt code validation
.PHONY: lint
lint:
	@echo "golangci-lint run ..."
	@golangci-lint run

# Format all Go code using goimports
.PHONY: fmt
fmt:
	@echo "goimports ..."
	@goimports -l -w .

# Using go mod tidy to organize Go module dependencies
.PHONY: tidy
tidy:
	@echo "go mod tidy -v ..."
	@go mod tidy -v

# One click check: format code+organize dependencies
.PHONY: check
check:
	@echo "fmt AND tidy..."
	@$(make) fmt
	@$(make) tidy