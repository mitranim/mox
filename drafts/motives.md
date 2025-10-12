Most languages tend to be great for some use cases, but suddenly fall flat for some others. I am tired of having to juggle a bazillion of different tools. Some people have one favorite, but there are reasons, other than legacy and inertia, why no single language has "won the world" (C hasn't either). I refuse to believe that a winning design is impossible.

The world needs a final programming language for everything.

* All the targets:
  * Systems-level programming.
  * Easy scripting.
  * Web (WASM _and_ compact JS).
  * GPU programming (see Mojo; same code, multiple targets). Can be delayed, but must be planned for.
  * Embedded devices. Can be delayed, but must be planned for.
* Statically typed.
* AOT compilation out of the box.
* Tiered JIT out of the box.
  * May involve an interpreter, may involve bytecode and a VM for "cold" paths. The final JIT tier and AOT compilation should generate the same code.
  * Running a script does not involve compiling a binary to disk and running it; it simply uses JIT. Startup must be fast.
  * Calling a new executable has multiple issues:
    * Huge first-time latency on MacOS as the OS validates the file; takes seconds for large executables.
    * Disk pollution.
* Arbitrary compile-time evaluation (via tiered JIT).
* Runtime code loading and JIT compilation in both development _and_ production (with static AOT in production builds).
  * Use cases:
    * Fast development (with hot code reloading).
    * Mods.
    * Plugins.
    * Compiler ships with a watch + HCR mode, which is preferred for development.
  * The compiler is part of the stdlib; any app can import the compiler, ship with it, and load more code at runtime (mods / plugins); loaded code is JIT-compiled and dynamically linked (not interpreted).
  * Regular code, using regular imports, is:
    * In development: lazily loaded (only if used), JIT-compiled, and dynamically linked into the current process.
    * In production: AOT compiled and statically linked.
  * Ideally (if doable): fine-grained sandboxing of loaded code:
    * No APIs by default, like in WASM.
    * Everything opt-in at the level of the host app.
    * Fine-grained IO permissions: FS = specific dirs, network = specific hosts, ports, etc.
  * Easy export of "header" files declaring the host APIs, for "offline" type-checked development of external code; testing still requires a live host.
  * Should be zero-overhead after first call, like dylibs.
* Lazy code loading and analysis in development: load _only_ the code which is actually used in current execution; avoid analyzing and compiling code which is loaded but not being used.
* Easy and zero-overhead interop with C/C++ and equivalent.
* RAII and guaranteed deinit (drop / destruct / destructor, names vary).
  * Deinit-on-drop is the preferred way of freeing resources, and is tracked and guaranteed by the language.
  * See Rust (`drop` is commonly used), Swift (`deinit` is supported, needs investigation), Nim (`=destroy` is opt-in), and more.
  * Default: shared memory within one thread, with deterministically-timed reference counting with cycle detection (possible or not? ARC has the former, ORC has the latter; needs investigation); sending between threads requires either synchronization with move semantics, or copying.
    * Source code identical to a GC language, but with predictable timing and memory safety.
  * Opt-in: move and borrow semantics.
  * Consider eager drop: every value is deinited immediately after its last use, not when it goes out of scope. See Mojo: https://docs.modular.com/mojo/manual/lifecycle/death/.
* Exceptions with traces by default, with explicit opt-out.
  * Exception-based control flow and error traces are orthogonal and have separate opt-outs.
  * Can opt out at file level.
  * Can opt out as module level.
  * Can opt out as package level (just once per repo).
  * Can opt out of traces in release builds (vs dev / debug builds). Can be toggled at runtime or via env vars (like in Rust).
  * For code which has opted out of exceptions:
    * The language guarantees that control flow is always visible. It's not possible to have a hidden exception or a hidden return, regardless of metaprogramming features.
    * The language requires explicit error handling when calling code which uses exceptions.
    * The language may have some fatal exceptions which always crash the program, but should strive to make it possible to handle all errors, including OOM, division by zero, etc. The latter requires math operations to have explicit errors in their signatures, in this mode.
* Unified threading model:
  * It should be possible to switch between different threading implementations, such as OS-level threads vs language-level coroutines, without changing the code.
  * No "function coloring" with async/await.
  * See Nim 3.
* No scheduler overhead by default. Opt-in language-level concurrency.
  * Code for embedded devices just runs the main function / main loop.
  * Code for regular OSes runs on a regular OS thread by default.
  * Opt-in support for lightweight coroutines, with the intent of using OS-level async IO whenever possible.
    * User code can choose a scheduler/reactor. See Rust `async/await` where this is customizable.
    * Crucial: no `await` keyword or equivalent. No function color. Every call "blocks" by default. Instead, it's an _opt-out_: when calling a function which returns a future, prefixing the call with `async` (or similar) opts out of the "await"/"yield".
    * A preemptive scheduler is a later goal. Begin with a cooperative scheduler.
    * When stack traces are enabled, trace all the way up to the ancestor (main entry point of the program).
    * Context-sensitive IO may be desirable: thread-level code uses sync IO, coroutine code uses async IO (when available), with no changes in the code. Again, we must avoid function coloring. Threads were invented for a reason. Go is quite possibly the only major language so far that has avoided inventing a _different version_ of threads, and having both versions in one language, and calling this "progress" (maybe Haskell too, but it's not quite as major).
* Unified fields / methods / functions / operators. All 4 represent the exact same concept, and should be syntactically unified and interchangeable.
  * See Nim for the field / method / function equivalence.
* Self-hosting and self-bootstrappable (no C dependence even for bootstrap).
* Very simple and minimal syntax.
  * As little punctuation as possible.
  * As much type inference as possible.
  * Many opt-ins for lower-level behaviors (various compiler pragmas).
  * See Nim for many good ideas.
  * No commas or semicolons (other than as comment delimiters).
* It should be possible to write terse code, which relies on type inference and default pragmas, to look almost like a dynamic language.
  * However, it should also be possible to explicitly declare all types and pragmas, and it should be possible to lock down a particular file / package / project such that the compiler _requires_ explicit declarations.
  * The terse inferred syntax is for scripts, minimal apps, and high-level business logic.
  * The fully-explicit syntax is for code which is in danger or being error-prone otherwise. In some projects, the entire codebase could be like that, which is why the lock-down option could be useful.
* Built-in support for implicit context / contextual parameters / dynamic variables (the concepts are closely related), which works for threads _and_ coroutines, and is inherited by default.
  * Presumably implemented with TLS for threads, and custom mechanisms for coroutines.
  * Functions should be able to declare "this parameter / dynamic variable must be set", and have this checked at compile time throughout their call chain. Maybe be compatible with only static dispatch, not dynamic.
* Support for compiling to compact, near-plain JS:
  * Requires implicit memory management in the core language; source code should look GC-able-like.
  * May need the ability to transparently switch default string encoding.
    * Motive: save space and performance in JS, while using UTF-8 on native targets.
    * Native target: string literals and `Str` are UTF-8.
    * JS target: string literals and `Str` are UTF-16.
    * Encoding differences are encapsulated in `Str` which is defined differently depending on the target.
    * `Str` abstracts away everything encoding-specific:
      * count graphemes
      * count grapheme clusters
      * iterate graphemes
      * iterate grapheme clusters
      * get grapheme at index
      * get grapheme cluster at index
      * convert to wire format (always UTF-8; should be skipped in JS operations where the engine handles that, like `fetch`)
    * Note that correctly handling graphemes requires iteration even in UTF-16 due to surrogate pairs. Ironically, they defeat the whole purpose of UTF-16, which was simplicity and O(1) length and lookup by index. (Also, grapheme clusters have this effect on every encoding, but for UTF-8 it's nothing new.)
* As homoiconic as practically possible.
* The language is its own data format.
  * Quoted code is data.
  * No separate format for config files.
  * There may be a separate file extension, but it's a strict subset of the language.
* Compiling _to_ C/C++ libraries:
  * Compiler produces `.h` and `.o` files.
  * `.h` files have documentation comments from source code.
  * Compiler also produces `.h`-like files with the _original_ syntax, containing declarations with documentation comments. This is the preferred documentation format. It could even be usable for web docs as-is (but with syntax highlighing and hyperlinks).
  * Documentation comments:
    * Like in Go, any comment attached to a declaration becomes its doc.
    * Like in Go, no special comment syntax / prefix is needed.
    * Like in Rust and Swift, comments use a variant of Markdown, with support for symbol-reference hyperlinks.
  * This automatically makes the language embeddable.
* Optional but convenient feature: cosmopolitan executables.
  * See https://justine.lol/ape.html
  * See https://justine.lol/cosmopolitan/
  * APE: "actually portable executable". This format allows the same file to execute on every major OS, as well as being bootable.
  * Cosmopolitan Libc allows C programs to run on every major OS, abstracting away syscall differences. We could do something similar. This involves bundling all syscall implementations in the executable, and detecting at runtime what to use.
* Support compilation to multiple targets (platforms, architectures) in a single run of the compiler (parse / analyze / optimize just once, proceed to emit code).

See `./lang.mox` for more.
