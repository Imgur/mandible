#!/bin/bash
# The script does automatic checking on a Go package and its sub-packages, including:
# 1. gofmt         (http://golang.org/cmd/gofmt/)
# 2. goimports     (https://github.com/bradfitz/goimports)
# 3. golint        (https://github.com/golang/lint)
# 4. go vet        (http://golang.org/cmd/vet)
# 5. race detector (http://blog.golang.org/race-detector)
# 6. test coverage (http://blog.golang.org/cover)

export GO15VENDOREXPERIMENT=1

set -e

PROJECTS="./uploadedfile ./server ./imageprocessor ./imagestore ./config ."

# Automatic checks
test -z "$(gofmt -l -w .     | tee /dev/stderr)"
# test -z "$(goimports -l -w . | tee /dev/stderr)"
# test -z "$(golint .          | tee /dev/stderr)"
godep go vet $PROJECTS
godep go test -race $PROJECTS

# Run test coverage on each subdirectories and merge the coverage profile.

echo "mode: count" > profile.cov

# Standard go tooling behavior is to ignore dirs with leading underscors
for dir in $PROJECTS
do
if ls $dir/*.go &> /dev/null; then
    godep go test -covermode=count -coverprofile=$dir/profile.tmp $dir
    if [ -f $dir/profile.tmp ]
    then
        cat $dir/profile.tmp | tail -n +2 >> profile.cov
        rm $dir/profile.tmp
    fi
fi
done

godep go tool cover -func profile.cov

[ ${COVERALLS_TOKEN} ] && goveralls -coverprofile=profile.cov -service travis-ci -repotoken $COVERALLS_TOKEN
