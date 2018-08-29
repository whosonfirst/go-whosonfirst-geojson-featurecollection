CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test -d src; then rm -rf src; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-geojson-featurecollection
	cp -r encode src/github.com/whosonfirst/go-whosonfirst-geojson-featurecollection/
	cp *.go src/github.com/whosonfirst/go-whosonfirst-geojson-featurecollection/
	cp -r vendor/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

deps:
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-geojson-v2"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-index"
	mv src/github.com/whosonfirst/go-whosonfirst-geojson-v2/vendor/github.com/whosonfirst/warning src/github.com/whosonfirst/
	# mv src/github.com/whosonfirst/go-whosonfirst-geojson-v2/vendor/github.com/whosonfirst/go-whosonfirst-spr src/github.com/whosonfirst/
	# mv src/github.com/whosonfirst/go-whosonfirst-geojson-v2/vendor/github.com/whosonfirst/go-whosonfirst-flags src/github.com/whosonfirst/

vendor-deps: rmdeps deps
	if test ! -d vendor; then mkdir vendor; fi
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt *.go
	go fmt cmd/*.go
	go fmt encode/*.go

bin: 	self
	rm -rf bin/*
	@GOPATH=$(GOPATH) go build -o bin/wof-encode-featurecollection cmd/wof-encode-featurecollection.go
