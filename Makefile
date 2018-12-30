
PREFIX ?= /usr/local

DIST := \
	COPYING \
	Makefile \
	mtmediasrv.service \
	mtmediasrv.yaml \
	nginx-server.conf \
	readme.md \
	mtmediasrv.conf

PROJECT := mtmediasrv
VERSION = 5
BUILD = `git describe --tags --always`

$(PROJECT): main.go
	go build -ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)" -o $(PROJECT)

build: $(PROJECT)

install: $(PROJECT)
	mkdir -p $(DESTDIR)$(PREFIX)/bin
	install -m0755 $(PROJECT) $(DESTDIR)$(PREFIX)/bin/$(PROJECT)
	mkdir -p $(DESTDIR)$(PREFIX)/lib/tmpfiles.d
	install -m0644 mtmediasrv.conf $(DESTDIR)$(PREFIX)/lib/tmpfiles.d/
	mkdir -p $(DESTDIR)$(PREFIX)/lib/systemd/system
	install -m0644 mtmediasrv.service $(DESTDIR)$(PREFIX)/lib/systemd/system
	mkdir -p $(DESTDIR)$(PREFIX)/share/mtmediasrv
	install -m0644 mtmediasrv.yaml $(DESTDIR)$(PREFIX)/share/mtmediasrv/mtmediasrv.yaml.example
	@echo 'Installation complete. You may have to run:'
	@echo '`sudo systemd-tmpfiles --create`'
	@echo '`sudo systemctl daemon-reload`'
	@echo '`sudo systemctl enable mtmediasrv`'
	@echo '`sudo systemctl start mtmediasrv`'

clean:
	go clean

dist:
	rm -rf $(PROJECT)-$(BUILD)
	mkdir $(PROJECT)-$(BUILD)
	cp $(DIST) $(PROJECT)-$(BUILD)/
	rm -f $(PROJECT)-$(BUILD)/$(PROJECT)
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)" -o $(PROJECT) -o $(PROJECT)-$(BUILD)/$(PROJECT)
	zip -r $(PROJECT)-$(BUILD)-x86_64.zip $(PROJECT)-$(BUILD)/

