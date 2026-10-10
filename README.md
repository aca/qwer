# qwer

A markdown-defined command runner, an alternative to make, just, and xc.

## rationale

I wanted a command runner that works out of the box, with no new language to learn.
Several tools try to replace make, but none fit:

- [just](https://github.com/casey/just) is growing its own [expression language](https://just.systems/man/en/expressions-and-substitutions.html), repeating make's mistake.
  Its tree-sitter support is also weak, which hurts when recipes are written in many languages.
- [xc](https://github.com/joerdav/xc) uses markdown, but only reads README.md and only runs `sh`.

qwer keeps the markdown idea and drops the limits: commands live in plain markdown,
so any editor highlights each code block in its own language, and each block can run
any interpreter through a shebang.

## usage

Given qwer.md

    ## echo
    ```
    echo echo "$@"
    ```

    ## echo2
    scripts can be written in any language via a shebang
    ```
    #!/usr/bin/env -S deno run -A
    console.log("echo2", Deno.args);
    ```

    ### sub
    this provides the subcommand `echo2 sub`
    ```
    #!/usr/bin/env -S deno run -A
    console.log("echosub", Deno.args);
    ```

This provides the commands `echo`, `echo2`, and `echo2 sub`:
```
qwer echo hello world
qwer echo2 hello world
qwer echo2 sub hello world
```

Fuzzy matching is supported:
```
qwer e2 hello world
```

## build

Requires zig 0.17. [md4c](https://github.com/mity/md4c) is fetched as a zig package.

```
zig build -Doptimize=ReleaseSmall
```

The flake installs prebuilt binaries from `dist/` (x86_64-linux, aarch64-darwin).
After changing the source, regenerate them with `./dist.sh`. `nix build .#source`
builds from source and should be byte-identical to `dist/qwer-x86_64-linux`.

The original Go implementation lives in `go/` and is deprecated.

## cachiex

```
./dist.sh
nix run nixpkgs#cachix -- push kata $(nix build --print-out-paths)
```
