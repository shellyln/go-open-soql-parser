

# GOOS   - darwin windows linux ...
# GOARCH - 386 amd64 arm arm64 ...
# GOARM  - 5 6 7 ...

DEVNUL      := /dev/null
DIRSEP      := /
SEP         := :
RM_F        := rm -f
RM_RF       := rm -rf
CP          := cp
CP_FLAGS    :=
CP_R        := cp -RT
CP_R_FLAGS  :=
FINDFILE    := find . -type f -name
WHICH       := which
PRINTENV    := printenv
GOXOS       := darwin
GOXARCH     := arm64
GOXARM      :=
GOCMD       := go
GOBUILD     := $(GOCMD) build
GOTIDY      := $(GOCMD) mod tidy
GOCLEAN     := $(GOCMD) clean
GOTEST      := $(GOCMD) test
GOVET       := $(GOCMD) vet
GOSHADOW    := $(shell $(GOCMD) env GOPATH)/bin/shadow
GOLINT      := $(shell $(GOCMD) env GOPATH)/bin/staticcheck
TINYGOCMD   := tinygo
SRCS        :=
CLI_NAME    := app
LIB_NAME    := lib
TARGET_CLI  := ./
TARGET_LIB  := ./lib
TARGET_WASM := ./wasm
BIN_CLI     := $(CLI_NAME)
BIN_LIB     := $(LIB_NAME).a
BIN_SO      := $(LIB_NAME).so
BIN_WASM    := web/go.wasm


ifeq ($(OS),Windows_NT)
    BIN_CLI := $(CLI_NAME).exe
    BIN_LIB := $(LIB_NAME).a
    BIN_SO  := $(LIB_NAME).dll
    ifeq ($(MSYSTEM),)
        SHELL      := cmd.exe
        DEVNUL     := NUL
        DIRSEP     := \\
        SEP        := ;
        RM_F       := del /Q
        RM_RF      := rmdir /S /Q
        CP         := copy
        CP_FLAGS   := /Y
        CP_R       := xcopy
        CP_R_FLAGS := /E /I /Y
        FINDFILE   := cmd.exe /C "where /r . "
        WHICH      := where
        PRINTENV   := set
    endif
endif

define normalize_dirsep
    $(subst /,$(DIRSEP),$1)
endef

define find_file
    $(subst $(subst \,/,$(CURDIR)),.,$(subst \,/,$(shell $(FINDFILE) $1)))
endef


# Usage of cp -R and cp
# $(CP_R) $(call normalize_dirsep,path/to/src) $(call normalize_dirsep,path/to/dest) $(CP_R_FLAGS)
# $(CP)   $(call normalize_dirsep,path/to/src) $(call normalize_dirsep,path/to/dest) $(CP_FLAGS)


SRCS     := $(call find_file,"*.go")
VERSION  := $(shell git describe --tags --abbrev=0 2> $(DEVNUL) || echo "0.0.0-alpha.1")
REVISION := $(shell git rev-parse --short HEAD)

LDFLAGS        := -ldflags="-s -w -buildid= -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\" -extldflags \"-static\""
LDFLAGS_SHARED := -ldflags="-s -w -buildid= -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\""


.PHONY: printenv clean cleantest upgrade tidy test testinfo cover lint wasm tinywasm doc
all: clean test build


printenv:
	@echo SHELL      : $(SHELL)
	@echo CURDIR     : $(CURDIR)
	@echo DEVNUL     : $(DEVNUL)
	@echo DIRSEP     : "$(DIRSEP)"
	@echo SEP        : "$(SEP)"
	@echo WHICH GO   : $(shell $(WHICH) $(GOCMD))
	@echo GOXOS      : $(GOXOS)
	@echo GOXARCH    : $(GOXARCH)
	@echo GOXARM     : $(GOXARM)
	@echo VERSION    : $(VERSION)
	@echo REVISION   : $(REVISION)
	@echo SRCS       : $(SRCS)
	@echo LDFLAGS    : $(LDFLAGS)
	@echo TARGET_CLI : $(TARGET_CLI)
	@echo BIN_CLI    : $(BIN_CLI)
	@echo BIN_LIB    : $(BIN_LIB)
	@echo BIN_SO     : $(BIN_SO)


clean:
	$(GOCLEAN)
	-$(RM_F) $(BIN_WASM)

cleantest:
	$(GOCLEAN) -testcache


upgrade:
	$(GOCMD) get -u && $(GOTIDY)

tidy:
	$(GOTIDY)


test:
	$(GOTEST) ./...

test+info:
	$(GOTEST) -gcflags=-m ./...

cover:
	$(GOTEST) -cover ./...

lint:
	@echo "Run go vet..."
	$(GOVET) ./...
	@echo "Run shadow..."
	$(GOVET) -vettool="$(GOSHADOW)" ./...
	@echo "Run staticcheck..."
	$(GOLINT) ./...


wasm: export GOOS:=js
wasm: export GOARCH:=wasm
wasm:
	$(CP) "$(shell $(GOCMD) env GOROOT)/misc/wasm/wasm_exec.js" web/.
	$(GOBUILD) \
	    -a -tags wasm \
	    -trimpath \
	    -buildvcs=false \
	    $(LDFLAGS) \
	    -o $(BIN_WASM) $(TARGET_WASM)


tinywasm: export GOOS:=js
tinywasm: export GOARCH:=wasm
tinywasm:
	$(CP) "$(shell $(TINYGOCMD) env TINYGOROOT)/targets/wasm_exec.js" web/.
	$(TINYGOCMD) build \
	    -tags wasm \
	    -no-debug \
	    -stack-size 512kB \
	    -ldflags="-X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\"" \
	    -o $(BIN_WASM) $(TARGET_WASM)


doc:
	godoc -http=:6060
