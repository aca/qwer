# qwer

Markdown defined command runner alternative to make, just, xc.

## rationale

I needed simple command runner that just works out of the box.
There's several options out there that tries to replace old "make".

[just](https://github.com/casey/just)
- it is now making similar mistakes as make with it's own [recipes](https://just.systems/man/en/expressions-and-substitutions.html).
- treesitter support doesn't work quite well, which is important when writing file that may contains arbitary languages.

[xc](https://github.com/joerdav/xc)
- it uses markdown, but limited to read README.md only, doesn't support other than 'sh'

## usage

Given qwer.md

    ## echo
    ```
    echo echo "$@"
    ```

    ## echo2
    script can be written in any language supported by shebang
    ```
    #!/usr/bin/env -S deno run -A
    console.log("echo2", Deno.args);
    ```

    ### sub
    this provides subcommand `echo2 sub`
    ```
    #!/usr/bin/env -S deno run -A
    console.log("echosub", Deno.args);
    ```

This provices commands `echo`, `echo2`, `echo2 sub`:
```
qwer echo hello world
qwer echo2 hello world
qwer echo2 sub hello world
```

fuzzy match is supported:
```
qwer e2 hello world
```


## nix

```
nix --extra-experimental-features "nix-command flakes" run github:aca/qwer/main -- --list
```
