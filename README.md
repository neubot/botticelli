# Botticelli

Server for Neubot and NDT tests written by Simone Basso, et al, in
the context of the [neubot](https://github.com/neubot) project.
This software is not affiliated with the NDT project and NDT server
side has was written using as reference the [official
specification](https://github.com/ndt-project/ndt/wiki/NDTProtocol).

Botticelli is still experimental software. As far as the NDT server
is concerned, botticelli is known to work with [measurement-kit's
NDT client](https://github.com/measurement-kit/measurement-kit) and
does not work with NDT v3.7.0.2 (but [a patch to make it work with
the official NDT has been
submitted](https://github.com/ndt-project/ndt/pull/216)). As regards
Neubot, it MAY work with some Neubot versions, but for now the main
focus of botticelli is to implement NDT.

We currently use botticelli to implement
[neubot-server](https://github.com/neubot/neubot-server), where
botticelli is used ONLY to implement a NDT server (meaning
that botticelli's Neubot specific code is not enabled).

## Cloning botticelli

Botticelli is written in [Golang](https://golang.org/) and you need to
have it installed on your system to compile botticelli.

To checkout botticelli you first need to select your `GOPATH`. In my
systems I typically use my home for that:

    export GOPATH=$HOME

Then you need to create the directory where to clone botticelli:

    install -d $GOPATH/src/github.com/neubot

Then close like this:

    cd $GOPATH/src/github.com/neubot
    git clone https://github.com/neubot/botticelli

## Compiling and cross compiling

You need to enter into botticelli's root directory first:

    cd $GOPATH/src/github.com/neubot/botticelli

Then compile for your system and architecture:

    go get -u -v

To run botticelli, execute:

    $GOPATH/bin/botticelli

Because botticelli is written in Go, it is also quite easy to cross
compile it for other systems and architectures, e.g.:

    GOOS=linux GOARCH=386 go get -u -v

The cross compiled binary will be located at:

    $GOPATH/bin/linux_386/botticelli

Consult [Golang docs](
https://golang.org/doc/install/source#environment<Paste>) for more
info on supported `GOOS` and `GOARCH` combinations.
