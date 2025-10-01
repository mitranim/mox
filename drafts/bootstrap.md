Since our design requirements include gradual self-bootstrapping of the compiler, it seems natural to try and start from scratch, namely from a small assembly core, hand-codable and verifiable, like certain parts of GNU Mes.

In the short term, we could instead bootstrap from C / Rust / JS (I know what the latter sounds like!).

## From scratch

At a high level:
- Stage 0: write a minimal compiler/runner in assembly; it runs Stage 1.
- Stage 1: write a minimal compiler/runner in our language; it runs Stage 2.
- Stage 2: compiler core executed by Stage 1; gradually evolves into the full language.

Codegen in 0+1 is simple and naive. Once we're up and running natively, we switch to an optimizing codegen library.

Eventually, Stage 2 compiles itself for real.

Bootstrap could be done in an emulator (container / Qemu / Blinkenlights) for both safety and portability; just target Linux x64 and never have to rewrite the assembly.

We don't just bootstrap once, we _stay_ bootstrappable, and we must do this without having to modify 0+1 ever again. They must provide the right kinds of "intrinsics" to the programs they compile/run (0→1 and 1→2), sufficient to allow all further evolution.

This means a Stage 2 compiler core 50 years in the future, evolved with many bells and whistles, as well as all standard library modules on which it depends, should _still_ be runnable via the same Stage 1, or at least with only minimal changes. We want to make it possible to define all those bells and whistles (like numeric types and arithmetic operations.....) on the fly _and use them_ immediately.

This is equivalent to saying that 50 years in the future, it should be possible to copy-paste compiler and stdlib code into some other random program which uses Stage 2 bells and whistles unsupported by Stage 1, _run it with Stage 1_, and _it should work_ (at a slow crawl).

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

## From scratch via LLVM

Similar to the "from scratch" plan but keeps native assembly to a minimum:
- Stage 0: entry point written in native assembly.
  - Loads LLVM libraries, or is statically linked with them.
  - Reads the Stage 1 file and uses LLVM (via C-style ABI) to JIT-compile and execute it.
- Stage 1: compiler core written in LLVM IR.
- Stage 2: the rest of the language.

The same approach can be used with any codegen library that has a decently "high level" IR with a text representation.

## From C or Rust

Similar to the JS plan, but:
- Start by implementing the native backend.
- Generating native code is easier.
- Requires more upfront work before being able to actually run the language. The core part is getting JIT and syscalls/IO working.

Using a portable IR + codegen library seems attractive. Some of the options:
- [LLVM](https://llvm.org) (more complicated to integrate; known for slow compilation; mature optimizations)
- [Cranelift](https://cranelift.dev) (known for fast compilation)

Relevant demo:
- https://github.com/bytecodealliance/cranelift-jit-demo

The following could be a useful read, but seems unsuitable for our purposes, since it only takes and emits text:
- [QBE](https://c9x.me/compile)

## From JS

Why?
- JS is one of our compilation targets / backends anyway. (For browsers.)
- Plugging new code into running compiler is trivial.
- Bootstrap requires little code. We can rewrite the language in itself early and avoid having to translate it later.

At a high level:
- Implement lang-to-JS in JS
- Rewrite in lang
- Implement lang-to-native in JS environment
- Implement lang-to-native in native environment

Steps:
- Implement our compiler in TS/JS, with support for only one compilation target: JS. Native comes later.
  - Follow the same self-bootstrapping principles as in the from-scratch bootstrap.
  - Has the same unresolved questions.
- Rewrite it in our language.
- Use it to recompile itself to JS.
  - The output must be nicely formatted and easily readable.
  - This is our "binary blob" (but readable).
  - We keep this "blob" in source control and drop the original JS code. Further compiler updates also update the "blob".
  - Optional: emit TS instead of plain JS.
- Organize the standard library.
  - The compiler is simply one of many modules in the stdlib.
  - The CLI entry point is simply a tiny wrapper / interface to it.
  - Programs written in our language should be able to import the compiler and use it at their runtime.
    - This should also be possible from other languages, but that's a lofty goal for later; see C/C++ interop.
- Isolate most of the standard library from target specifics. The following components are target-specific and used conditionally:
  - JS IO.
  - Syscalls and OS IO (NYI at this point).
  - Native codegen in JS environment; several options are available:
    - Find or write an FFI for native interop (can't work in browsers).
    - Call into WASM (TBD; more portable, probably more ergonomic, should work in browsers).
    - Write text files and shell out (can't work in browsers).
  - Native codegen in native environment: FFI with codegen library.
  - FFI in native environment (also involves linking).
  - Core language primitives for JS.
  - Core language primitives for native.
    - Memory management.
    - Core data types.
  - Compile-time execution for JS (works by creating a JS module and importing it).
  - Compile-time execution for native (works by JIT-compiling machine code, marking it executable, and jumping).
- Implement all this stuff for the native target (that's quite a lot...). We'll also need to handle linking.
- Run the compiler (still in JS) to compile itself for the native target.
- Run the new executable to recompile itself again, this time both for JS and for the native target.
- Test, debug, etc.

Don't forget to sandbox the whole thing; we don't want an accidental `rm -rf /`-style syscall from some misplaced asm instruction.

Since we want to stay bootstrappable, this requires keeping the compiler runnable in JS until native bootstrap is implemented. On the plus side, this capability is nice for web playgrounds. (The alternative is WASM; see notes below.)

## WASM

Eventually we'll be targeting WASM as one of the backends. Portable codegen libs make it easy. WASM programs can use IO in WASI-supporting runtimes, and be compiled to executables. See https://github.com/bytecodealliance/wasm-tools.

It's worth investigating targeting WASM in compiler bootstrap. There are significant differences. This replaces syscalls with WASI calls, and changes our strategy for compile-time execution. Instead of dumping machine code into executable memory and issuing a jump, this involves creating new WASM modules, loading them, exporting / importing functions, and sharing memory. Like dynamic linking, but with extra steps (and overheads).

Targeting WASM is also simpler in the sense that codegen is portable right off the bat without having to link with codegen libs. Could learn lessons from https://porffor.dev.

## C

Some languages, most notably Nim, just print C code and shell out to a C compiler. This is unattractive for various reasons:
- No native on-the-fly execution.
- Involves depending on C forever.

Our goals include support for loading arbitrary code at runtime not just in development, but in production, and running it without significant overhead. This rules out interpretation. It's possible to compile C on the fly to a dylib and load that dylib, but this has way more latency than the native JIT approach, so we'd rather not.

## Misc

Languages which start by targeting an existing backend, such as LLVM or C, eventually develop their own higher-level IR (intermediate representation). Recent examples include Rust and Nim. This makes language-specific optimizations easier. As a side effect, it also makes compiler backends easier to swap. LLVM devs have noticed this, which resulted in [MLIR](https://mlir.llvm.org), which we should look into.

Long-lived languages also tend to develop alternatives to LLVM for faster compilation. Moderately recent examples include JSC's B3 (9 years old at the time of writing), also Cranelift.
