
PREFIX ?= /usr/local

DIST := COPYING Makefile mtmediasrv.service mtmediasrv.yaml nginx-server.conf readme.md

PROJECT := mtmediasrv
VERSION = 1
BUILD = `git describe --tags --always`

$(PROJECT): main.go
	go build -ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)" -o $(PROJECT)

build: $(PROJECT)

install: $(PROJECT)
	go install
	mkdir -p $(DESTDIR)$(PREFIX)/bin
	install -m0755 $(PROJECT) $(DESTDIR)$(PREFIX)/bin/$(PROJECT)

clean:
	go clean

dist:
	rm -rf $(PROJECT)-$(BUILD)
	mkdir $(PROJECT)-$(BUILD)
	cp $(DIST) $(PROJECT)-$(BUILD)/
	GOOS=linux GOARCH=386 go build -ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)" -o $(PROJECT) -o $(PROJECT)-$(BUILD)/$(PROJECT)
	zip -r $(PROJECT)-$(BUILD)-ia32.zip $(PROJECT)-$(BUILD)/
	rm -f $(PROJECT)-$(BUILD)/$(PROJECT)
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)" -o $(PROJECT) -o $(PROJECT)-$(BUILD)/$(PROJECT)
	zip -r $(PROJECT)-$(BUILD)-x86_64.zip $(PROJECT)-$(BUILD)/

