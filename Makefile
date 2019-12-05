#Makefile
# yuuki.miyo@gmail.com

NAME     := $(shell basename `pwd`)
BINDIR   := bin
REPO     := github.com/yuukimiyo/$(NAME)
VERSION  := v0.1

args	:=
SRCS    := $(shell find . -type f -name '*.go')

.PHONY: build
build:
	go build -o $(BINDIR)/$(NAME)

.PHONY: clean
clean:
	rm -f $(BINDIR)/*
	rm -fr vendor

.PHONY: init
init: clean
	rm -f Gopkg.*
	dep init
	dep ensure

.PHONY: ensure
ensure:
	dep ensure

.PHONY: dev
dev:
	@go build -o $(BINDIR)/$(NAME)
	$(BINDIR)/$(NAME) $(args)
