# Understanding the Floor Function

## What is the Floor Function?

The floor function takes any number and rounds it DOWN to the nearest integer. Think of it like this:
- You're standing on a number line
- The floor function tells you which integer is directly below you (or exactly where you are if you're already on an integer)

## Visual Examples

### Positive Numbers
```
Number line:  0----1----2----3----4----5
              
2.7 → floor(2.7) = 2     (rounds DOWN to 2)
3.0 → floor(3.0) = 3     (already an integer, stays 3)
3.1 → floor(3.1) = 3     (rounds DOWN to 3)
3.9999 → floor(3.9999) = 3  (rounds DOWN to 3, even though it's very close to 4)
```

### Negative Numbers (This is where it gets tricky!)
```
Number line:  -5----(-4)----(-3)----(-2)----(-1)----0
              
-2.7 → floor(-2.7) = -3    (rounds DOWN to -3, NOT up to -2)
-3.0 → floor(-3.0) = -3    (already an integer, stays -3)
-3.1 → floor(-3.1) = -4    (rounds DOWN to -4, NOT up to -3)
-3.0001 → floor(-3.0001) = -4  (even tiny amounts make it round DOWN to -4)
```

## Why Negative Numbers are Confusing

Many people think floor should "remove the decimal part", but that's actually TRUNCATION:
- Truncate(3.7) = 3 ✓
- Truncate(-3.7) = -3 ✗ (This is WRONG for floor!)

Floor ALWAYS rounds toward negative infinity:
- Floor(3.7) = 3 (same as truncate for positive)
- Floor(-3.7) = -4 (different from truncate for negative!)

## The Problem with Float64

When we convert rational numbers to float64, we lose precision:

```
-3.000000000000000000001 as a rational = -3000000000000000000001/1000000000000000000000
                                         (exact representation)

-3.000000000000000000001 as a float64 ≈ -3.0
                                         (loses the tiny .000000000000000000001 part!)
```

This is why the old implementation was wrong - it would think -3.000000000000000000001 was exactly -3.0!

## How Our Algorithm Works

1. **We have a fraction**: numerator/denominator
   - Example: -3.1 = -31/10

2. **Integer division**: Divide numerator by denominator, ignoring remainder
   - -31 ÷ 10 = -3 (with remainder)
   - This gives us the "truncated" value

3. **Check if there was a remainder**: 
   - We multiply our result back: -3 × 10 = -30
   - Compare with original: -31 < -30? YES, there was a remainder!

4. **For negative numbers with remainder**: Subtract 1
   - -3 - 1 = -4
   - This is our floor value!

## Step-by-Step Examples

### Example 1: floor(-3.1)
```
1. Convert to fraction: -3.1 = -31/10
2. Integer division: -31 ÷ 10 = -3 (truncated)
3. Check remainder: -3 × 10 = -30, and -31 < -30, so there IS a remainder
4. Since negative with remainder: -3 - 1 = -4
Result: floor(-3.1) = -4
```

### Example 2: floor(3.1)
```
1. Convert to fraction: 3.1 = 31/10
2. Integer division: 31 ÷ 10 = 3 (truncated)
3. Check remainder: 3 × 10 = 30, and 31 > 30, so there IS a remainder
4. Since positive, we don't subtract: stays 3
Result: floor(3.1) = 3
```

### Example 3: floor(-3.0)
```
1. Convert to fraction: -3.0 = -3/1
2. Integer division: -3 ÷ 1 = -3 (exact)
3. Check remainder: -3 × 1 = -3, and -3 = -3, so NO remainder
4. No remainder, so no adjustment needed
Result: floor(-3.0) = -3
```

### Example 4: floor(-3.000000000000000000001)
```
1. Convert to fraction: -3000000000000000000001/1000000000000000000000
2. Integer division: -3000000000000000000001 ÷ 1000000000000000000000 = -3
3. Check remainder: -3 × 1000000000000000000000 = -3000000000000000000000
   Original -3000000000000000000001 < -3000000000000000000000, so there IS a remainder
4. Since negative with remainder: -3 - 1 = -4
Result: floor(-3.000000000000000000001) = -4
```

This is why we needed to fix the function - using float64 would lose that tiny .000000000000000000001 part!