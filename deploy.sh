#!/bin/sh
set -e

prefix=/usr/local
sysconfdir=/etc
bindir=$prefix/bin
libdir=$prefix/lib

botticelli=botticelli-linux-amd64

install -d $bindir
install -v $botticelli $bindir/botticelli
install -v -d $libdir/systemd/system
install -v -m644 botticelli.service $libdir/systemd/system
install -d $sysconfdir
install -v rc.local $sysconfdir
