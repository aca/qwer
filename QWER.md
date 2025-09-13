# QWER Commands

This is the QWER.md file with custom commands.

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
