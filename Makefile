.PHONY: build timefind-man-pages

SHELL=/bin/bash

#
# XXX make sure to tag your releases! See CONTRIBUTING.
#
VERSION=$(shell git describe --tags --abbrev=0 | cut -c 2-)
VERSION_FULL=$(shell git describe --tags | cut -c 2-)

# test equality
VERSION_MATCH=$(shell ([ "$(VERSION)" == "$(VERSION_FULL)" ] && echo "1") || echo "0")

# NOTE: DESTDIR must include a trailing /
DESTDIR=
PREFIX=/usr
BINDIR=$(DESTDIR)$(PREFIX)/bin
DOCDIR=$(DESTDIR)$(PREFIX)/share/timefind-$(VERSION)
MANDIR=$(DESTDIR)$(PREFIX)/share/man
MAN1DIR=$(DESTDIR)$(PREFIX)/share/man/man1

export GOPATH := ${PWD}

SYMLINK_VENDOR=0
VENDOR_DIRS=

all:: bin/timefind bin/timefind_indexer timefind-man-pages

test:
	pushd src/timefind; go test ./...

bin/timefind: clean
	go build -ldflags "-X main.TimefindTimestamp=`date -u +%Y-%m-%dT%H:%M:%S` -X main.TimefindCommit=`git rev-parse HEAD`" -o ./bin/timefind ./src/timefind

bin/timefind_indexer: clean
	go build -ldflags "-X main.IndexerTimestamp=`date -u +%Y-%m-%dT%H:%M:%S` -X main.IndexerCommit=`git rev-parse HEAD`" -o ./bin/timefind_indexer ./src/timefind/indexer

# build is really "install.local"
# and for if you run locally
build:
	make install_programs BINDIR=./bin
	make install_READMEs DOCDIR=./bin


#
# install is what is used for the .rpm
#
install: install_programs install_READMEs install_LICENSE install_man_pages

install_programs: bin/timefind bin/timefind_indexer
	-mkdir -p $(BINDIR)
	cp ./bin/timefind $(BINDIR)
	cp ./bin/timefind_indexer $(BINDIR)/timefind_indexer
	cp ./src/timefind_lander_indexer/timefind_lander_indexer $(BINDIR)

install_READMEs:
	-mkdir -p $(DOCDIR)
	cp ./src/timefind/README $(DOCDIR)/README.timefind
	cp ./src/timefind/indexer/README $(DOCDIR)/README.timefind_indexer

install_LICENSE:
	-mkdir -p $(DOCDIR)
	cp ./COPYRIGHT ./LICENSE $(DOCDIR)

install_man_pages: timefind-man-pages
	-mkdir -p $(MAN1DIR)
	cp ./src/timefind/timefind.1 $(MAN1DIR)
	cp ./src/timefind/indexer/timefind_indexer.1 $(MAN1DIR)

clean:
	rm -rf bin/timefind bin/timefind_indexer

#
# release stuff
#
TV=timefind-$(VERSION)
tar.gz:
ifneq ($(VERSION_MATCH),1)
	$(error Repository tag is "$(VERSION_FULL)"! Expecting something like "$(VERSION)". Did you properly tag your release? (See CONTRIBUTING for more information))

else
	ln -s . $(TV)
	tar \
		--transform='flags=r;s|README\.timefind|README|' \
		--exclude "$(TV)/src/timefind/timefind" \
		--exclude "$(TV)/src/timefind/indexer/indexer" \
		--exclude "$(TV)/src/timefind/indexer/tests" \
		-czvf timefind-$(VERSION).tar.gz \
		$(TV)/CHANGELOG \
		$(TV)/CONTRIBUTORS \
		$(TV)/COPYRIGHT \
		$(TV)/LICENSE \
		$(TV)/Makefile \
		$(TV)/README.timefind \
		$(TV)/src/timefind \
		$(TV)/src/timefind_lander_indexer
	rm -f $(TV)
endif

RPM_DIST=$(shell rpm --eval '%{dist}')

rpms:
	cp timefind-$(VERSION).tar.gz $$HOME/rpmbuild/SOURCES
	cp timefind.spec  $$HOME/rpmbuild/SPECS
	( cd $$HOME/rpmbuild; rpmbuild -ba SPECS/timefind.spec; )
	cp $$HOME/rpmbuild/RPMS/x86_64/timefind-$(VERSION)-1$(RPM_DIST).x86_64.rpm .
	cp $$HOME/rpmbuild/SRPMS/timefind-$(VERSION)-1$(RPM_DIST).src.rpm .

#cp $$HOME/rpmbuild/RPMS/noarch/timefind-$(VERSION)-1$(RPM_DIST).noarch.rpm .
#	cp $$HOME/rpmbuild/SRPMS/timefind-$(VERSION)-1$(RPM_DIST).src.rpm .

RELEASE_FILES=timefind-$(VERSION).tar.gz \
		timefind-$(VERSION)-1$(RPM_DIST).x86_64.rpm \
		timefind-$(VERSION)-1$(RPM_DIST).src.rpm
release:
	cp $(RELEASE_FILES) $$HOME/WORKING/ANT/WWW/ant_2015/software/timefind
	cd $$HOME/WORKING/ANT/WWW/ant_2015/software/timefind && git add $(RELEASE_FILES)
	mv $(RELEASE_FILES) RELEASE

#
# man pages:
#
src/timefind/timefind.1: src/timefind/README
	pandoc -s -f markdown -t man -o $@ $< 

src/timefind/indexer/timefind_indexer.1: src/timefind/indexer/README
	pandoc -s -f markdown -t man -o $@ $< 

.PHONY:
timefind-man-pages: src/timefind/indexer/timefind_indexer.1 src/timefind/timefind.1
