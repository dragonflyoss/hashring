# Copyright 2026 The Dragonfly Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

all: help

# Run code lint of golang and rust
lint: lint-go lint-rust
.PHONY: lint

# Run code lint of golang
lint-go:
	@echo "Begin to golangci-lint."
	@golangci-lint run
.PHONY: lint-go

# Run code lint of rust
lint-rust:
	@echo "Begin to cargo fmt and clippy."
	@cargo fmt --all -- --check
	@cargo clippy --all --all-targets --all-features -- -D warnings
.PHONY: lint-rust

# Run tests of golang and rust
test: test-go test-rust
.PHONY: test

# Run tests of golang
test-go:
	@echo "Begin to go test."
	@go test -race ./...
.PHONY: test-go

# Run tests of rust
test-rust:
	@echo "Begin to cargo test."
	@cargo test --all-features
.PHONY: test-rust

# Clear compiled files
clean:
	@go clean
	@cargo clean
.PHONY: clean

help:
	@echo "make lint                           run code lint of golang and rust"
	@echo "make lint-go                        run code lint of golang"
	@echo "make lint-rust                      run code lint of rust"
	@echo "make test                           run tests of golang and rust"
	@echo "make test-go                        run tests of golang"
	@echo "make test-rust                      run tests of rust"
	@echo "make clean                          clean"
