GOOS=linux
GOARCH=amd64

all: ilno-server

VERSION := 0.1.0
BUILD_DATE := `date +%FT%T%z`
LD_FLAGS := "-X 'github.com/fiatjaf/ilno/version.Version=$(VERSION)' -X 'github.com/fiatjaf/ilno/version.BuildTime=$(BUILD_DATE)'"


ilno-server: $(shell ag -l --go) server/bindata.go
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags $(LD_FLAGS) -o ilno-server

server/bindata.go: static/js/embed.min.js
	go-bindata -fs -o server/bindata.go -pkg server -prefix "static/" static/js/...

static/js/%.min.js: js/%.js $(shell find ./js)
	./node_modules/esbuild/bin/esbuild --bundle --minify --sourcemap --outfile=$@ --define:process.env.NODE_ENV='"production"' $<

static/js/%.js: js/%.js $(shell find ./js)
	./node_modules/esbuild/bin/esbuild --bundle --sourcemap --outfile=$@ --define:process.env.NODE_ENV='"development"' $<
