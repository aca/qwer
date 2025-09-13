# Project Commands

This README defines commands that can be executed using the markdown runner.
Commands can be written in any language using shebang notation.

## build
```
go build -o mdrun main.go
echo "Build completed"
```

## test
```
go test ./...
```

### unit
```
go test -short ./...
```

### integration
```
go test -run Integration ./...
```

### coverage
```
go test -cover ./...
```

## lint
```
go fmt ./...
go vet ./...
```

## clean
```
rm -f mdrun
rm -f *.test
echo "Cleanup completed"
```

## dev

### setup
```
go mod download
go mod tidy
echo "Development environment ready"
```

### watch
```
echo "Watching for file changes..."
ls -la *.go
```

## deploy

### staging
```
echo "Deploying to staging..."
echo "Build app..."
echo "Upload to staging server..."
echo "Staging deployment complete!"
```

### production
```
echo "Deploying to production..."
echo "Running tests first..."
echo "Build optimized version..."
echo "Upload to production server..."
echo "Production deployment complete!"
```

## info
```
echo "Markdown Runner"
echo "==============="
go version
pwd
ls -la
```

## echo

should echo

```
#!/usr/bin/env bash
# Echo command that accepts arguments
if [ $# -eq 0 ]; then
    echo "No arguments provided"
else
    echo "Arguments received:"
    for arg in "$@"; do
        echo "  - $arg"
    done
fi
```

## grep-files
```
#!/usr/bin/env bash
# Grep wrapper that accepts pattern and files
if [ $# -lt 1 ]; then
    echo "Usage: grep-files <pattern> [files...]"
    exit 1
fi

pattern="$1"
shift

echo "Searching for: $pattern"
if [ $# -eq 0 ]; then
    echo "Searching in all .go and .md files:"
    grep -n "$pattern" *.go *.md 2>/dev/null || echo "No matches found"
else
    echo "Searching in specified files:"
    grep -n "$pattern" "$@" 2>/dev/null || echo "No matches found"
fi
```

## hello
```
echo "Hello from markdown runner!"
date
```

### world
```
echo "Hello, World!"
echo "Current time: $(date '+%Y-%m-%d %H:%M:%S')"
```

### user
```
echo "Hello, $USER!"
echo "Your home directory: $HOME"
```

## scripts

### bash
```
#!/usr/bin/env bash
echo "Bash Script with Shebang"
echo "========================"
echo "Bash version: $BASH_VERSION"
echo "Current directory: $(pwd)"
echo "User: $USER"
echo "Script arguments: $@"
echo "Number of arguments: $#"

# Show all arguments
if [ $# -gt 0 ]; then
    echo "Arguments received:"
    for arg in "$@"; do
        echo "  - $arg"
    done
fi

# Array example
arr=(apple banana cherry)
echo "Fruits: ${arr[@]}"

for fruit in "${arr[@]}"; do
    echo "  - $fruit"
done

# Function example
greet() {
    echo "Hello, $1!"
}

greet "World"
```

### python
```
#!/usr/bin/env python3
import sys
import datetime

print("Python Script Execution")
print("=" * 30)
print(f"Python version: {sys.version}")
print(f"Current time: {datetime.datetime.now()}")
print(f"Platform: {sys.platform}")

# Calculate something
numbers = [1, 2, 3, 4, 5]
total = sum(numbers)
print(f"Sum of {numbers} = {total}")
```

### node
```
#!/usr/bin/env node
console.log("Node.js Script Execution");
console.log("=".repeat(30));
console.log(`Node version: ${process.version}`);
console.log(`Platform: ${process.platform}`);
console.log(`Current directory: ${process.cwd()}`);

// Some JavaScript logic
const numbers = [1, 2, 3, 4, 5];
const sum = numbers.reduce((a, b) => a + b, 0);
console.log(`Sum of [${numbers}] = ${sum}`);

// JSON output
const data = {
    timestamp: new Date().toISOString(),
    environment: process.env.NODE_ENV || 'development'
};
console.log("Data:", JSON.stringify(data, null, 2));
```

### ruby
```
#!/usr/bin/env ruby
puts "Ruby Script Execution"
puts "=" * 30
puts "Ruby version: #{RUBY_VERSION}"
puts "Platform: #{RUBY_PLATFORM}"
puts "Current time: #{Time.now}"

# Ruby logic
numbers = [1, 2, 3, 4, 5]
sum = numbers.sum
puts "Sum of #{numbers} = #{sum}"

# Hash example
data = {
  user: ENV['USER'],
  home: ENV['HOME'],
  timestamp: Time.now.to_i
}
puts "Data: #{data}"
```

### perl
```
#!/usr/bin/env perl
use strict;
use warnings;

print "Perl Script Execution\n";
print "=" x 30 . "\n";
print "Perl version: $]\n";

# Get current time
my $time = localtime();
print "Current time: $time\n";

# Perl logic
my @numbers = (1, 2, 3, 4, 5);
my $sum = 0;
$sum += $_ for @numbers;
print "Sum of (@numbers) = $sum\n";

# Environment
print "User: $ENV{USER}\n" if $ENV{USER};
print "Home: $ENV{HOME}\n" if $ENV{HOME};
```

### go-script
```
#!/usr/bin/env -S go run
package main

import (
    "fmt"
    "runtime"
    "time"
)

func main() {
    fmt.Println("Go Script Execution")
    fmt.Println("==============================")
    fmt.Printf("Go version: %s\n", runtime.Version())
    fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
    fmt.Printf("Current time: %s\n", time.Now().Format(time.RFC3339))

    // Go logic
    numbers := []int{1, 2, 3, 4, 5}
    sum := 0
    for _, n := range numbers {
        sum += n
    }
    fmt.Printf("Sum of %v = %d\n", numbers, sum)
}
```
