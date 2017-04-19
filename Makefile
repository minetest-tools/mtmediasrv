
PREFIX ?= /usr/local

PROJECT := mtmediasrv
VERSION = 1
BUILD = $(PROJECT)-`git describe --tags --always`

$(PROJECT): main.go
	go build -ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)" -o $(PROJECT)

build: $(PROJECT)

install: $(PROJECT)
	go install
	mkdir -p $(DESTDIR)$(PREFIX)/bin
	install -m0755 $(PROJECT) $(DESTDIR)$(PREFIX)/bin/$(PROJECT)

clean:
	go clean

