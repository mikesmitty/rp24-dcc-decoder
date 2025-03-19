#!/bin/sh

rm -fv cover.out cover.html
GOCACHE=$(mktemp -d)
go env -w GOCACHE=$GOCACHE
go test -v -coverprofile cover.out ./pkg/... && go tool cover -html cover.out -o cover.html && open cover.html
go env -u GOCACHE
