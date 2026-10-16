# E70 keyboard MIDlet: CLDC 1.1 / MIDP 2.0.
#
#   make           build + package checks + reproducible rebuild
#   make build     -> dist/E70Kbd.jar, dist/E70Kbd.jad, dist/SHA256SUMS
#   make check     checks only (after build)
#   make clean
#
# Needs: a JDK 11+ (only runs the build tools), Python 3, curl, unzip.
# Build dependencies are downloaded to .deps/j2me and verified against deps.lock.

JAVA ?= java
PYTHON ?= python3

DEPS  ?= .deps/j2me
ECJ   := $(DEPS)/ecj-3.33.0.jar
CLDC  := $(DEPS)/cldcapi11-2.0.4.jar
MIDP  := $(DEPS)/midpapi20-2.0.4.jar
PROGUARD := $(DEPS)/proguard-7.9.1/lib/proguard.jar
# Zopfli sources (CafeUndZopfli, Apache-2.0) for tools/ZopfliDir.java: host only
ZOPFLI := $(DEPS)/cafeundzopfli-5cdf283

# class version 46.0 (as the MIDlets shipped on Series 40 3rd Edition phones)
SOURCE_LEVEL := 1.3
TARGET_LEVEL := 1.2

BUILD ?= build
DIST  ?= dist
JSR82 = $(BUILD)/stubs
FILE_BASE := $(shell sed -n 's/^FILE_BASE=//p' app.properties)

SRCS := $(shell find src -name '*.java' | sort)
STUB_SRCS := $(shell find stubs -name '*.java' | sort)

.PHONY: all build check deps clean

all: build check

deps:
	@tools/fetch_deps.sh $(DEPS)

build: deps
	rm -rf $(BUILD)/classes $(BUILD)/preverified $(BUILD)/res $(BUILD)/mapping.txt $(BUILD)/zopfli $(BUILD)/deflated $(JSR82) $(DIST)
	mkdir -p $(BUILD)/classes $(BUILD)/preverified $(BUILD)/res $(BUILD)/zopfli $(BUILD)/deflated $(JSR82)
	cp res/icon.png $(BUILD)/res/icon.png
	$(JAVA) -jar $(ECJ) -source $(SOURCE_LEVEL) -target $(TARGET_LEVEL) \
		-bootclasspath $(CLDC):$(MIDP) -encoding UTF-8 -g:none \
		-failOnWarning -d $(JSR82) $(STUB_SRCS)
	$(JAVA) -jar $(ECJ) -source $(SOURCE_LEVEL) -target $(TARGET_LEVEL) \
		-bootclasspath $(CLDC):$(MIDP):$(JSR82) -encoding UTF-8 -g:none \
		-failOnWarning -d $(BUILD)/classes $(SRCS)
	$(JAVA) -jar $(PROGUARD) @e70kbd.pro \
		-injars $(BUILD)/classes -outjars $(BUILD)/preverified \
		-libraryjars $(CLDC) -libraryjars $(MIDP) -libraryjars $(JSR82) \
		-printmapping $(BUILD)/mapping.txt
	@test -f $(BUILD)/mapping.txt || : > $(BUILD)/mapping.txt
	$(JAVA) -jar $(ECJ) -8 -nowarn -encoding UTF-8 -d $(BUILD)/zopfli $(ZOPFLI)/*.java tools/ZopfliDir.java
	$(JAVA) -cp $(BUILD)/zopfli ZopfliDir $(BUILD)/deflated $(BUILD)/preverified $(BUILD)/res
	$(PYTHON) tools/package.py app.properties $(BUILD)/preverified $(DIST) "" $(BUILD)/res $(BUILD)/deflated

check:
	@test -f $(DIST)/SHA256SUMS || { echo "run 'make build' first" >&2; exit 1; }
	$(PYTHON) tools/check.py app.properties $(DIST) $(BUILD)/mapping.txt $(CLDC) $(MIDP) $(JSR82)
	@echo "reproducibility: clean rebuild into $(BUILD)/repro"
	@rm -rf $(BUILD)/repro
	@$(MAKE) --no-print-directory BUILD=$(BUILD)/repro/build DIST=$(BUILD)/repro/dist build >$(BUILD)/repro.log 2>&1 \
		|| { cat $(BUILD)/repro.log; exit 1; }
	@cmp $(DIST)/SHA256SUMS $(BUILD)/repro/dist/SHA256SUMS \
		&& echo "PASS  reproducible: second build has identical SHA-256" \
		|| { echo "FAIL  reproducible: SHA-256 differs"; exit 1; }

clean:
	rm -rf $(BUILD) $(DIST)
