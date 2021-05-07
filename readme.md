## Overview

Scratch for an experimental language syntax.

The Sublime Text syntax implementation **requires ST >= 4075**. To enable, symlink this folder into packages. MacOS version:

    ln -sf "$(pwd)/sublime" "$HOME/Library/Application Support/Sublime Text/Packages/mox"

## Levels

The syntax has two levels:

* Language-independent data notation. See `examples/reference_data_notation.mox`.

* Some hypothetical language. See `examples/reference_language*`.

## Inspiration

In no particular order: Lisps, Rebol, Erlang, Go, Haskell, Rust, and more.

## Semantics

This repo is dedicated to syntax. Language semantics are mentioned only where relevant for syntax design.

## Syntax

Syntax and conventions are designed using _objective_ metrics: less thinking, less typing, fewer typing errors, easy-to-modify code.

Only one delimiter (`[]`) → less thinking, easier typing.

No unnecessary punctuation → less thinking, less typing, fewer errors, easier to modify.

Explicit delimiters → easier to modify, because of special editor support for delimiters.

Placing closing delimiters on separate lines → easier to modify intermediary code.

Only prefix calls → less thinking, no infix precedence errors.

Simple universal structure → better for DSLs.

Choice of `[]` over `()` and `{}`: easier to type on most keyboard layouts.
