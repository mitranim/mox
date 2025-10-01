Misc notes and thoughts.

## Comments in the AST

Since we intend to generate docs from comments, the parser should link comment nodes with expression nodes. Doing this at the parser level seems natural because this association is newline-sensitive.

Any other code dealing with the AST should be able to jump from any declared symbol to its comment, if any.
