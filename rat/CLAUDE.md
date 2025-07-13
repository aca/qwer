# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a Go package that provides an alternative API for rational arithmetic, wrapping Go's `math/big.Rat` with a more convenient interface. The package supports various input types (int, float, string, fractions, percentages) and provides immutable operations.

## Commands

### Testing
```bash
# Run all tests
go test

# Run tests with verbose output
go test -v

# Run a specific test
go test -run TestPowInt -v

# Run tests with coverage
go test -cover
```

### Building
```bash
# Build the package
go build

# Check for compilation errors
go build ./...
```

### Development
```bash
# Format code
go fmt ./...

# Run go vet
go vet ./...
```

## Architecture

The package centers around the `Rational` struct which wraps `big.Rat` and adds precision control. Key design decisions:

1. **Immutable Operations**: All arithmetic methods (`Add`, `Sub`, `Mul`, `Quo`, `PowInt`) return new `Rational` instances rather than modifying the receiver. This is critical for the API design.

2. **Flexible Input**: The `Rat()` constructor accepts multiple types (int variants, float32/64, string, *Rational). String inputs support:
   - Regular numbers: "10", "10.5"
   - Fractions: "1/2", "3/4"
   - Percentages: "50%", "12.5%"

3. **Method Naming**: 
   - Comparison methods are `Equal()`, `Less()`, `Greater()` (NOT `IsEqual`, `IsLessThan`, `IsGreaterThan`)
   - All comparison methods now accept `any` type as input, consistent with the flexible input design
   - This inconsistency existed in `RatMin` and `RatMax` which incorrectly used the `Is*` variants - these have been fixed

4. **Precision Control**: Each `Rational` has its own precision setting (default 8) which affects string output but not internal calculations.

## Important Implementation Details

- **Add/Sub Bug**: The original implementation had a bug where `Add` and `Sub` would return after processing only the first argument. This has been fixed to process all variadic arguments.

- **Float Precision**: Float inputs like `0.1` will display with full precision (e.g., "0.10000000") due to float64 representation limitations.

- **Error Handling**: Methods panic on invalid types rather than returning errors, following the pattern of the standard library's math operations.

## Performance Optimizations

Several performance optimizations have been implemented:

1. **Arithmetic Operations**: `Add` and `Sub` methods now convert primitive types (int, float) directly to `big.Rat` instead of creating intermediate `Rational` objects, reducing allocations.

2. **FloorInt Method**: Replaced string parsing with direct `big.Rat.Float64()` conversion, improving performance by ~6.5x (from 359ns to 55ns).

3. **PowInt Method**: Implemented binary exponentiation (exponentiation by squaring) instead of naive repeated multiplication, providing O(log n) complexity instead of O(n).

## Performance Results

Key improvements from benchmarks:
- **Add with integers**: ~1.8x faster (497ns vs 896ns) with 40% fewer allocations
- **FloorInt**: ~6.5x faster (55ns vs 360ns) with 87% fewer allocations  
- **PowInt**: More efficient for large exponents due to O(log n) algorithm

## Known Issues

- The error message in `Sub()` says "rat: add invalid type" instead of "rat: sub invalid type" (minor issue).