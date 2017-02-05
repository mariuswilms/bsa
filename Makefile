# Copyright 2014 David Persson. All rights reserved.
#
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

PREFIX ?= /usr/local
VERSION ?= head-$(shell git rev-parse --short HEAD)

PROG_GOFLAGS = -X main.Version=$(VERSION)

.PHONY: install
install: $(PREFIX)/sbin/bsa

.PHONY: uninstall
uninstall:
	rm $(PREFIX)/sbin/bsa

.PHONY: clean
clean:
	if [ -d ./dist ]; then rm -r ./dist; fi

.PHONY: dist
dist: dist/bsa dist/bsa-darwin-amd64 dist/bsa-linux-amd64

$(PREFIX)/sbin/%: dist/%
	install -m 555 $< $@

dist/%-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(PROG_GOFLAGS)" -o $@

dist/%-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(PROG_GOFLAGS)" -o $@

dist/%:
	go build -ldflags "$(PROG_GOFLAGS)" -o $@
