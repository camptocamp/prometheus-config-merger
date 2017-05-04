DEPS = $(wildcard */*.go)
VERSION = $(shell git describe --always --dirty)

all: prometheus-config-merger

prometheus-config-merger: main.go $(DEPS)
	CGO_ENABLED=0 GOOS=linux \
	  go build -a \
		  -ldflags="-X main.version=$(VERSION)" \
	    -installsuffix cgo -o $@ $<
	strip $@

clean:
	rm -f prometheus-config-merger

.PHONY: all clean
