# QWER Commands

This is the QWER.md file with custom commands.

## hello
```
echo "Hello from QWER.md!"
```

## test
```
echo "Running tests from QWER.md"
go test -v ./...
```

## version
```
#!/usr/bin/env bash
echo "QWER Command Runner"
echo "File: QWER.md"
echo "Arguments: $@"
```

## date
```
date '+%Y-%m-%d %H:%M:%S'
```