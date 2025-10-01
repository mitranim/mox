## Overview

Drafts / wishlists for a hypothetical programming language that doesn't exist but needs to. It could be built one day.

See [motives.md](drafts/motives.md) and [lang.mox](drafts/lang.mox). See other files in [drafts](drafts) for condensed examples.

The lang draft explains the syntax first (at length) because it's different. Semantics are explained later and form a larger part of the motivation.

The aim is to properly balance efficiency between _every_ aspect of the design. Runtime efficiency, compilation efficiency, cognitive efficiency, even typing efficiency, should be balanced. Some parts should also be configurable, with tuning knobs for different projects and preferences.

Most languages tend to be great for some use cases, but suddenly fall flat for some others. I am tired of having to juggle a bazillion of different tools. Some people have one favorite, but there are reasons, other than legacy and inertia, why no single language has "won the world" (C hasn't either). I refuse to believe that a winning design is impossible.

## Comparison and criticism

At the moment, comparisons with existing projects are meaningless because they exist and this doesn't. In any case, no major language has _all_ of our goals (see [motives.md](drafts/motives.md)). The closest match seems to be Nim.

[Nim](https://nim-lang.org): extremely nice replacement for C/C++; positives: too many to list here, go try it out; negatives include:
* Compile-time execution is interpreter-only; no JIT-to-native.
* More complicated syntax than we want; not based on a data notation.

Since [Rust](https://rust-lang.org) is all the hype these days, a comparison is warranted. Positives:
* Nice replacement for C/C++.
* Very powerful yet safe.
* Supports compile-time execution.

Negatives:
* User code can't freely switch between AOT and JIT.
  * Focused on AOT compilation, no JIT support.
  * CTFE is interpreter-only.
  * Procedural macros can only transform code.
* Extremely complicated syntax; barrier to entry; unsuitable for quick scripting and high level business logic.
* The borrow system adds noise.
* The ownership and borrow system often stands in the way of high level logic, requiring workarounds.

## Sublime Text

The directory `./sublime` provides syntax highlighting and other niceties for the [Sublime Text](https://sublimetext.com) editor. Requires ST >= 4075. To enable, symlink that folder into ST's `Packages`. Example for MacOS:

```sh
ln -sf "$(pwd)/sublime" "$HOME/Library/Application Support/Sublime Text/Packages/mox"
```
