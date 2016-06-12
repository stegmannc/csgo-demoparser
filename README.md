csgo-demoparser
===============

A Golang utility to parse CSGO demo files.

This is under active development, does not fully function yet.

## Requirements

You need a normal golang setup. See https://golang.org/doc/install
For OSX users install go via homebrew with `brew install go` and set all path variables:

		export GOPATH=$gopath
		export GOROOT=/usr/local/opt/go/libexec
		export PATH=$PATH:$GOPATH/bin
		export PATH=$PATH:$GOROOT/bin

## Build and Run
`go install`

`csgo-demoparser -f demo.dem`
