# PROG:=hardblame
CONFIG:=hardblame.yaml

VERSION:=$(shell git describe --dirty=+WiP --always)
GOFLAGS:=-v -ldflags "-X app.version=$(VERSION) -v"

UNAME:=$(shell uname -m)
HOSTARCH:=$(shell if [ "${UNAME}" == "x86_64" ]; then echo amd64; else echo ${UNAME}; fi)

GOOS ?= $(shell uname -s | tr A-Z a-z)
# GOARCH:=arm64

# GO:=GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=1 go
GO:=GOOS=$(GOOS) GOARCH=$(HOSTARCH) go

CERTDIR:=etc/certs

# default: ${PROG}

all:
	@if [ ! -e ${CERTDIR}/RootCA.crt ] ; then make certs; fi
	$(MAKE) -C sharkd

# ${PROG}: build

# build:	
# 	$(GO) build $(GOFLAGS) -o ${PROG}

arch:
	@echo Host architecture: ${HOSTARCH} \(uname: ${UNAME}\)

install: ${PROG}
	@mkdir -p ./etc ./sbin ./data
	if [ ! -e ./etc/${CONFIG} ] ; then install -c ${CONFIG}.sample ../etc/${CONFIG}; fi

test:
	$(GO) test -v -cover

clean:
	@rm -f $(PROG)

certs:
	mkdir -p "${CERTDIR}"
	openssl req -x509 -nodes -new -sha256 -days 1024 -newkey rsa:2048 -keyout "${CERTDIR}/RootCA.key" -out "${CERTDIR}/RootCA.pem" -subj "/C=US/CN=HealthRoot-CA"
	openssl x509 -outform pem -in "${CERTDIR}/RootCA.pem" -out "${CERTDIR}/RootCA.crt"
	openssl req -new -nodes -newkey rsa:2048 -keyout "${CERTDIR}/localhost.key" -out "${CERTDIR}/localhost.csr" -subj "/C=SE/ST=Confusion/L=Lost/O=HealthCertificates/CN=localhost.local"
	openssl x509 -req -sha256 -days 1024 -in "${CERTDIR}/localhost.csr" -CA "${CERTDIR}/RootCA.pem" -CAkey "${CERTDIR}/RootCA.key" -CAcreateserial -extfile domains.ext -out "${CERTDIR}/localhost.crt"


.PHONY: build clean generate

