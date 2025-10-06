#!/usr/bin/env elvish

use str

# Simulate the completion function
fn complete {|@args|
    var partial-cmd = [(drop 1 $args)]
    var partial-input = ""
    var has-empty = $false

    # Check if last arg is empty (user pressed space)
    if (and (> (count $partial-cmd) 0) (eq $partial-cmd[(- (count $partial-cmd) 1)] "")) {
        set has-empty = $true
        set partial-cmd = [(take (- (count $partial-cmd) 1) $partial-cmd)]
    } elif (> (count $partial-cmd) 0) {
        # Last arg is partial input being typed
        set partial-input = $partial-cmd[(- (count $partial-cmd) 1)]
        set partial-cmd = [(take (- (count $partial-cmd) 1) $partial-cmd)]
    }

    var prefix = (str:join " " $partial-cmd)
    var next-index = (count $partial-cmd)

    # Add space if we have completed commands
    if (> $next-index 0) {
        set prefix = $prefix" "
    }

    echo "Input args:" $args
    echo "partial-cmd:" $partial-cmd
    echo "partial-input: '"$partial-input"'"
    echo "prefix: '"$prefix"'"
    echo "next-index:" $next-index
    echo "Completions:"

    var seen = [&]
    qwer --list | each {|line|
        if (str:has-prefix $line $prefix) {
            var parts = [(str:split " " $line)]
            if (>= (count $parts) (+ $next-index 1)) {
                var candidate = $parts[$next-index]
                # Filter by partial input
                if (str:has-prefix $candidate $partial-input) {
                    if (not (has-key $seen $candidate)) {
                        echo "  -> "$candidate
                        set seen[$candidate] = $true
                    }
                }
            }
        }
    }
}

# Test cases - simulating what elvish passes to the completer
echo "=== Test 1: typing 'qwer ' (with space) ==="
complete qwer ""

echo "\n=== Test 2: typing 'qwer e' ==="
complete qwer e

echo "\n=== Test 3: typing 'qwer echo2 ' (with space) ==="
complete qwer echo2 ""

echo "\n=== Test 4: typing 'qwer echo2 s' ==="
complete qwer echo2 s
