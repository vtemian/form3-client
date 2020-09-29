#!/usr/bin/env sh

HOST=${TEST_API_HOST:-"http://localhost:8080"}

go get github.com/onsi/ginkgo/ginkgo
go mod download
make check
make tests
