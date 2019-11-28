#Makefile
# yuuki.miyo@gmail.com

NAME     := $(shell basename `pwd`)
BINDIR   := bin
REPO     := github.com/yuukimiyo/$(NAME)
VERSION  := v0.0.1

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
	$(BINDIR)/$(NAME) -f /mnt/v01/resource/yuho/risktext/risk_20140801_20160731_categorysentenses/social_economic.csv $(args)
#	@$(BINDIR)/$(NAME) -f /mnt/v01/resource/wikipedia/jawiki/20191001/extract/jawiki-20191001-categorylinks.sql $(ARG)
#	@$(BINDIR)/$(NAME) -f /mnt/v01/resource/wikipedia/jawiki/20191001/extract/jawiki-20191001-pages-articles6.xml-p2534193p4013905
