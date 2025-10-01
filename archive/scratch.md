Known issues:

* Prefix vs infix vs postfix. Prefix is required. Postfix may be better for chaining.
* Delimiter-free calls with fixed arity. Everything nullary/unary/binary. Tuples should be rare.
  * Getter/setter calls: arity mismatch: unary vs binary.
* Reserve best syntax for methods. Methods take priority over global functions.
* Code vs markup. Embedding. Mixing.
  * Embedding is not enough, mixing must be well supported.
  * It should be possible to "invert" the priority.
  * It should be possible to treat a markdown file as a markup Mox file with embedded Mox.
    * Embedding syntax should be compatible and flexible.
* Reserve good syntax for chaining.
* Whitespace sensitivity.
  * We were never going to be completely whitespace-insensitive.
  * All languages are sensitive to spaces.
  * Newline sensitivity may be considered.
  * Indentation sensitivity may be considered, but delimiters are more attractive.

------

How does one go about inverting code and markup? (See: "literate programming".)

## Table of Contents

[[md_to_toc]]

## Some other content

[[img_box_link `/images/one.jpg` `Two` `https://three.four/five`]]

[[
  img_box_link
  src `/images/one.jpg`
  alt `Two`
  href `https://three.four/five`
]]

[[img_box_link src `/images/one.jpg` alt `Two` href `https://three.four/five`]]

[[img_box_link, src `/images/one.jpg`, alt `Two`, href `https://three.four/five`]]

[[img_box_link [src `/images/one.jpg`] [alt `Two`] [href `https://three.four/five`]]]

[[img_box_link, src = `/images/one.jpg`, alt = `Two`, href = `https://three.four/five`]]
