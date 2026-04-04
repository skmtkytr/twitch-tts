APP_NAME := twitch-tts
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
DIST_DIR := dist
ARCHIVE := $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz

.PHONY: build install uninstall release clean

build:
	wails build

install: build
	cp build/bin/$(APP_NAME) $(APP_NAME)
	cp build/appicon.png appicon.png
	./install.sh
	rm -f $(APP_NAME) appicon.png

uninstall:
	./uninstall.sh

release: build
	@mkdir -p $(DIST_DIR)/$(APP_NAME)
	cp build/bin/$(APP_NAME) $(DIST_DIR)/$(APP_NAME)/
	cp build/appicon.png $(DIST_DIR)/$(APP_NAME)/
	cp install.sh $(DIST_DIR)/$(APP_NAME)/
	cp uninstall.sh $(DIST_DIR)/$(APP_NAME)/
	cp README.md $(DIST_DIR)/$(APP_NAME)/
	chmod +x $(DIST_DIR)/$(APP_NAME)/install.sh $(DIST_DIR)/$(APP_NAME)/uninstall.sh
	cd $(DIST_DIR) && tar czf $(APP_NAME)-$(VERSION)-linux-amd64.tar.gz $(APP_NAME)
	rm -rf $(DIST_DIR)/$(APP_NAME)
	@echo ""
	@echo "Release archive: $(ARCHIVE)"

clean:
	rm -rf $(DIST_DIR) build/bin
