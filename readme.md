## Overview

Scratch for experimental language syntax.

To enable in Sublime Text, symlink this folder into packages. MacOS version:

    ln -sf "$(pwd)" "$HOME/Library/Application Support/Sublime Text 3/Packages/"

## Levels

The syntax has two levels:

* Language-independent data notation. See `reference_data_notation.mox`.

* Some hypothetical language. See `reference_language.nox`.

## Inspiration

In no particular order: Clojure and other Lisps, Erlang, Go, Rust, and more.

## Semantics

This repo is dedicated to syntax. Language semantics are mentioned only where relevant for syntax design.

## Syntax

Syntax and conventions are designed using _objective_ metrics: less thinking, less typing, fewer typing errors, easy-to-modify code.

Only one delimiter (`{}`) → less thinking, easier typing.

No punctuation → less thinking, less typing, fewer errors, easier to modify.

Explicit delimiters → easier to modify, because of special editor support for delimiters.

Placing delimiters on separate lines → easier to modify intermediary code.

Only prefix notation → less thinking, no infix precedence errors.

Simple universal structure → better for DSLs.
