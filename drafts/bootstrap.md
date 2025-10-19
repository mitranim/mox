Most languages have to bootstrap from a full compiler written in another language. That's because their execution model does not permit on-the-fly self-bootstrapping. We aim for something different.

- Flexible execution model: no boundary between interpretation and compilation, from code perspective (only internal differences).
- Flexible compiler: the compiler / interpreter can be manipulated and extended on the fly, by the code it's compiling / running.

This may allow us to bootstrap from a tiny external core, fixed in size and complexity, and _stay_ bootstrappable that way, every time, without needing a full-blown alternative compiler.

Current plan, at a high level:

- Write a small boot-core in another language, doesn't matter which.
- The boot-core is a simple interpreter which provides a small number of core intrinsics:
  - Essential basics: declarations, procedures, maybe a few types.
  - Extensibility: the state of the running compiler, and of the program being compiled, is accessible and can be manipulated to extend the language.
- The boot-core has one job: _run_ our language's CLI frontend, instructing it to _build_ our language's CLI frontend, with all its dependencies.
- The CLI frontend file begins by importing the language prologue.
- The prologue defines many essential primitives, such as integer types and operations on them.
- The prologue defines more complicated things, such as type kinds (structs, unions, etc.), and _compiler_ logic for them.
- The prologue defines many essential compiler behaviors, such as type checking, ownership tracking, memory management, and more, and plugs them into the currently running compiler / interpreter, which runs these behaviors on all code from now on.
- The CLI frontend then imports a new compiler core, rewritten in our language.
- The new compiler core also has an interpreter, is extensible, and provides similar intrinsics as the boot-core. However, it also supports building executables.
- The boot-core interprets the CLI frontend, which runs the new compiler, pointed it self, to _build_ the CLI frontend.

From there, we proceed with the actual implementation, optimizations, etc.

For "hot" code, we'll be replacing interpretation with JIT compilation, but that's not relevant for bootstrapping.

The plan is to _stay_ bootstrappable while keeping the boot-core fixed in size and rarely having to modify it. It needs to provide the right kinds of instrinsics to allow all further evolution.

This means our compiler 50 years in the future, evolved with many bells and whistles, as well as all standard library modules on which it depends, should _still_ be runnable via the same boot-core, or at least with only minimal changes. We want to make it possible to define all those bells and whistles (like numeric types and arithmetic operations.....) on the fly _and use them_ immediately.

This is equivalent to saying that 50 years in the future, it should be possible to copy-paste the prologue and stdlib into some other random program which uses bells and whistles not intrinsically supported by the boot-core, _run it with the boot-core_, and _it should work_ (at a slow crawl).

This is equivalent to saying that the language is self-modifiable at the runtime of the compiler. While other languages keep adding intrinsics to their compilers, we rely on a small fixed set of compiler intrinsics and implement everything else as library code.

This is equivalent to saying that language modules should instruct the compiler exactly how to compile them and/or all later code. This means changing codegen on the fly, adding components to the already-running compiler, and more.

By bells and whistles here, we mean various intrinsic features which are usually done inside compilers and shipped inside the compiler binary. Features which are often impossible to do in library code and involve special syntax, compiler pragmas, and so on.

Other languages:

```
+-----------------+
|    compiler     |    +----------+
|        +        | <- | app code |
| 🔔 language 🎶  |    +----------+
+-----------------+
```

Us (conceptually):

```
                +---------------------+   +-----------+
+----------+    | 🔔 language code 🎶 |   | libraries |
| compiler | <- |          +          | + | with more |
+----------+    |       app code      |   |   🔔🎶    |
                +---------------------+   +-----------+

Or:

+----------+    +---------+    +---------+    +----------+
| compiler | <- | lang 🔔 | <- | lang 🎶 | <- | app code |
+----------+    +---------+    +---------+    +----------+
```

Naturally, we ship the whole language AOT-compiled and optimized. But it should be simultaneously possible to _opt out_ and _opt more in_ without changing the compiler.

Can this be done?
