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

# --- Release & AUR ---

AUR_DIR := /tmp/twitch-tts-bin
AUR_REPO := ssh://aur@aur.archlinux.org/twitch-tts-bin.git

.PHONY: release-tag aur

# Tag and push to trigger CI release
# Usage: make release-tag VERSION=0.3.0
release-tag:
ifndef VERSION
	$(error VERSION is required. Usage: make release-tag VERSION=0.3.0)
endif
	git tag v$(VERSION)
	git push origin master
	git push origin v$(VERSION)
	@echo ""
	@echo "Tagged v$(VERSION) and pushed. CI will create the release."
	@echo "Run 'gh run watch' to monitor, then 'make aur VERSION=$(VERSION)' to update AUR."

# Update PKGBUILD sha256, push to AUR
# Usage: make aur VERSION=0.3.0
aur:
ifndef VERSION
	$(error VERSION is required. Usage: make aur VERSION=0.3.0)
endif
	@set -e; \
	echo "Downloading v$(VERSION) release to compute sha256..."; \
	TMP=$$(mktemp -d); \
	gh release download v$(VERSION) -p '*.tar.gz' -D "$$TMP"; \
	SHA=$$(sha256sum "$$TMP/twitch-tts-v$(VERSION)-linux-amd64.tar.gz" | cut -d' ' -f1); \
	rm -rf "$$TMP"; \
	echo "sha256: $$SHA"; \
	sed -i "s/^pkgver=.*/pkgver=$(VERSION)/" aur/PKGBUILD; \
	sed -i "s/^sha256sums=.*/sha256sums=('$$SHA')/" aur/PKGBUILD; \
	git add aur/PKGBUILD; \
	git commit -m "update PKGBUILD for v$(VERSION)"; \
	git push origin master; \
	echo "Pushing to AUR..."; \
	rm -rf $(AUR_DIR); \
	git clone $(AUR_REPO) $(AUR_DIR); \
	cp aur/PKGBUILD $(AUR_DIR)/; \
	cd $(AUR_DIR) && makepkg --printsrcinfo > .SRCINFO; \
	cd $(AUR_DIR) && git add PKGBUILD .SRCINFO && git commit -m "Update to v$(VERSION)" && git push; \
	rm -rf $(AUR_DIR); \
	echo ""; \
	echo "AUR updated to v$(VERSION)"
