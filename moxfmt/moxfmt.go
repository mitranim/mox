package main

/*
CLI for formatting code.

After cloning the repo, install this globally:

	go install ./moxfmt

Run it:

	moxfmt

To allow the shell to find the executable, make sure the shell profile file
(`~/.profile` on my system) contains this:

	export GOPATH=~/go
	export PATH=$PATH:$GOPATH/bin

To integrate with Sublime Text, use https://github.com/mitranim/sublime-fmt.
*/

import (
	"io"
	"mox"
	"os"
	"strings"
)

func main() {
	var buf strings.Builder
	_, err := io.Copy(&buf, os.Stdin)
	if err != nil {
		panic(err)
	}

	ast, err := mox.Parse(buf.String())
	if err != nil {
		panic(err)
	}

	mox.FmtMox(ast)

	_, err = os.Stdout.Write([]byte(mox.Format(ast)))
	if err != nil {
		panic(err)
	}
}
