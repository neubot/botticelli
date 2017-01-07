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

## Compiling and cross compiling

Botticelli is written in Go and you need to have Go installed on
your system to compile it.

Make sure the `GOPATH` environment variable is set. The typical setup
that I use is the following:

- GOPATH is set to $HOME
- botticelli is checked out in `$GOPATH/src/github.com/neubot/botticelli`

Adjust `GOPATH` to your needs and you should be okay.

With these settings, compiling botticelli is as simple as running the
following command from `$GOPATH/src/github.com/neubot/botticelli`:

    go get -u -v

Because botticelli is written in Go, it is also quite easy to cross
compile it for other systems and architectures, e.g.:

    GOOS=linux GOARCH=386 go get -u -v
