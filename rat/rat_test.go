package rat

import (
	"encoding/json"
	"testing"
)

func Test10(t *testing.T) {
	a := Rat(1347)
	b := a.Quo(10).Floor().Mul(10)
	if b.String() != "1340" {
		t.Fatalf("expected 1340, got %s", b.String())
	}

	if a.String() != "1347" {
		t.Fatalf("expected 1347, got %s", a.String())
	}
}

func TestJson(t *testing.T) {
	var msg = `{ "balance": "1386929.37231066771348207123", "currency": "KRW" }`

	type Balance struct {
		Balance  *Rational `json:"balance"`
		Currency string    `json:"currency"`
	}

	v := &Balance{}
	err := json.Unmarshal([]byte(msg), v)
	if err != nil {
		t.Fatal(err)
	}

	v.Balance.precision = 20

	if v.Balance.String() != "1386929.37231066771348207123" {
		t.Fatalf("expected 1386929.37231066771348207123, got %s", v.Balance.String())
	}
}

func TestSub(t *testing.T) {
	a := Rat("2")
	if !a.Sub(-2).Equal(Rat(4)) {
		t.Fatalf("expected 4, got %s", a.Sub(-2))
	}

	assertEqual(t, a.String(), "2")
}

func TestParseFloat(t *testing.T) {
	// got := parseFloat64(1 / 100)
	got := parseFloat64(0.01)
	got.precision = 2
	if got.String() != "0.01" {
		t.Fatalf("expected 0.01, got %s", got.String())
	}
}

func TestAddNeg(t *testing.T) {
	got := parse("7.004")
	got = got.Add(Rat("0.001").Neg())
	if got.String() != "7.003" {
		t.Fatalf("expected 7.003, got %s", got.String())
	}
}

func TestCeil(t *testing.T) {
	// {
	// 	got := parse("1382.5")
	// 	if got.FloorInt() != 1382 {
	// 		t.Fatalf("expected 1382, got %d", got.FloorInt())
	// 	}
	// }
	// {
	// 	got := parse("-5.05")
	// 	if got.FloorInt() != -6 {
	// 		t.Fatalf("expected -6, got %d", got.FloorInt())
	// 	}
	// }
	{
		got := parse("0.95")
		if got.Ceil().Equal(Rat(1)) != true {
			t.Fatalf("expected 1382, got %v", got.Floor())
		}
	}
	{
		got := parse("4")
		if got.Ceil().Equal(Rat(4)) != true {
			t.Fatalf("expected -6, got %v", got.Floor())
		}
	}
	{
		got := parse("7.004")
		if got.Ceil().Equal(Rat(8)) != true {
			t.Fatalf("expected -6, got %v", got.Floor())
		}
	}
	{
		got := parse("-7.004")
		if got.Ceil().Equal(Rat(-7)) != true {
			t.Fatalf("expected -6, got %v", got.Floor())
		}
	}
}

func TestFloor(t *testing.T) {
	{
		got := parse("1382.5")
		if got.FloorInt() != 1382 {
			t.Fatalf("expected 1382, got %d", got.FloorInt())
		}
	}
	{
		got := parse("-5.05")
		if got.FloorInt() != -6 {
			t.Fatalf("expected -6, got %d", got.FloorInt())
		}
	}
	{
		got := parse("1382.5")
		if got.Floor().Equal(Rat(1382)) != true {
			t.Fatalf("expected 1382, got %v", got.Floor())
		}
	}
	{
		got := parse("-5.05")
		if got.Floor().Equal(Rat(-6)) != true {
			t.Fatalf("expected -6, got %v", got.Floor())
		}
	}
}

func TestNeg(t *testing.T) {
	a := parse("1382")
	if a.Neg().String() != "-1382" {
		t.Fatalf("expected -1382, got %s", a.Neg().String())
	}

	assertEqual(t, a.String(), "1382")
}

func TestMin(t *testing.T) {
	a := parse("1382")
	b := parse("1380")
	c := parse("1381")

	if RatMin(a, b, c).String() != "1380" {
		t.Fatalf("expected 1380, got %s", RatMin(a, b, c).String())
	}
	if a.String() != "1382" {
		t.Fatalf("expected 1380, got %s", a.String())
	}

	if b.String() != "1380" {
		t.Fatalf("expected 1381, got %s", b.String())
	}
	if c.String() != "1381" {
		t.Fatalf("expected 1382, got %s", c.String())
	}
}

func TestParcentage(t *testing.T) {
	a := Rat(10000).Mul(Rat("3%"))
	assertEqual(t, a.String(), "300")
}

func TestBasics(t *testing.T) {
	a := Rat("2")
	b := Rat("3")
	astr := a.String()
	bstr := b.String()

	b.Neg() // no-op
	a.Neg() // no-op

	assertEqual(t, "15", a.Add(b).Mul(b).IntString())
	assertEqual(t, "1.5", b.Quo(a).String())

	assertEqual(t, astr, a.String())
	assertEqual(t, bstr, b.String())
}

func TestQuo(t *testing.T) {
	a := parse("2")
	b := parse("4")
	if RatQuo(a, b).String() != "0.5" {
		t.Fatal("Quo")
	}
}

func TestPowInt(t *testing.T) {
	tests := []struct {
		name     string
		base     any
		exp      int
		expected string
	}{
		{"2^0", "2", 0, "1"},
		{"2^1", "2", 1, "2"},
		{"2^2", "2", 2, "4"},
		{"2^3", "2", 3, "8"},
		{"2^10", "2", 10, "1024"},
		{"3^3", "3", 3, "27"},
		{"0.5^2", "0.5", 2, "0.25"},
		{"1/2^3", "1/2", 3, "0.125"},
		{"10^-1", "10", -1, "0.1"},
		{"2^-2", "2", -2, "0.25"},
		{"0.5^-2", "0.5", -2, "4"},
		{"negative base", "-2", 3, "-8"},
		{"negative base even exp", "-2", 4, "16"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Rat(tt.base).PowInt(tt.exp)
			if result.String() != tt.expected {
				t.Errorf("Rat(%v).PowInt(%d) = %s, want %s", tt.base, tt.exp, result.String(), tt.expected)
			}
		})
	}
	
	// Test that original is not modified
	t.Run("immutability", func(t *testing.T) {
		a := Rat("2")
		b := a.PowInt(10)
		if a.String() != "2" {
			t.Errorf("Original value modified: got %s, want 2", a.String())
		}
		if b.String() != "1024" {
			t.Errorf("Result incorrect: got %s, want 1024", b.String())
		}
	})
}

func TestFloorCeilCorrectness(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedFloor string
		expectedCeil  string
		expectedFloorInt int
	}{
		// Positive numbers
		{"positive integer", "5", "5", "5", 5},
		{"positive decimal small", "5.1", "5", "6", 5},
		{"positive decimal large", "5.9", "5", "6", 5},
		
		// Negative numbers - this is where the bug likely is
		{"negative integer", "-5", "-5", "-5", -5},
		{"negative decimal small", "-5.1", "-6", "-5", -6}, // Floor should be -6, Ceil should be -5
		{"negative decimal large", "-5.9", "-6", "-5", -6}, // Floor should be -6, Ceil should be -5
		{"negative close to zero", "-0.1", "-1", "0", -1},   // Floor should be -1, Ceil should be 0
		
		// Zero and close to zero
		{"zero", "0", "0", "0", 0},
		{"positive close to zero", "0.1", "0", "1", 0},
		
		// The specific example from TODO comment
		{"todo example", "-7.004", "-8", "-7", -8}, // Current TODO says should be -7, but Floor(-7.004) should be -8
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Rat(tt.input)
			
			// Test FloorInt
			floorInt := r.FloorInt()
			if floorInt != tt.expectedFloorInt {
				t.Errorf("FloorInt(%s) = %d, want %d", tt.input, floorInt, tt.expectedFloorInt)
			}
			
			// Test Floor
			floor := r.Floor()
			if floor.String() != tt.expectedFloor {
				t.Errorf("Floor(%s) = %s, want %s", tt.input, floor.String(), tt.expectedFloor)
			}
			
			// Test Ceil
			ceil := r.Ceil()
			if ceil.String() != tt.expectedCeil {
				t.Errorf("Ceil(%s) = %s, want %s", tt.input, ceil.String(), tt.expectedCeil)
			}
			
			// Mathematical properties that should always hold:
			// 1. Floor(x) <= x <= Ceil(x)
			if !floor.Less(r) && !floor.Equal(r) {
				t.Errorf("Floor(%s) = %s should be <= %s", tt.input, floor.String(), tt.input)
			}
			if !r.Less(ceil) && !r.Equal(ceil) {
				t.Errorf("%s should be <= Ceil(%s) = %s", tt.input, tt.input, ceil.String())
			}
			
			// 2. If x is not an integer, then Ceil(x) = Floor(x) + 1
			if !r.bigrat.IsInt() {
				expectedCeilFromFloor := floor.Add(1)
				if !ceil.Equal(expectedCeilFromFloor) {
					t.Errorf("For non-integer %s: Ceil should equal Floor + 1. Got Ceil=%s, Floor+1=%s", 
						tt.input, ceil.String(), expectedCeilFromFloor.String())
				}
			}
		})
	}
}

func TestString(t *testing.T) {
	{
		a := parse("0.5/1")
		if a.String() != "0.5" {
			t.Fatal(a.String())
		}

		if _, exact := a.bigrat.FloatPrec(); !exact {
			t.Fatal("not exact")
		}
	}

	{
		a := parse("1/0.5")
		if a.String() != "2" {
			t.Fatal(a.String())
		}
		if _, exact := a.bigrat.FloatPrec(); !exact {
			t.Fatal("not exact")
		}
	}

	{
		a := parse("1/3")
		a.precision = 8
		if a.String() != "0.33333333" {
			t.Fatal(a.String())
		}
		if _, exact := a.bigrat.FloatPrec(); exact {
			t.Fatal("should not be exact")
		}
	}
}

// func TestCopy(t *testing.T) {
// 	a := parse("1/2")
// 	b := a.Clone()
// 	if a.Add(parse("1")).String() != "1.5" {
// 		t.Fatal()
// 	}
// 	if b.String() != "0.5" {
// 		t.Fatal()
// 	}
// }
//
// func TestZero(t *testing.T) {
// 	z := Zero()
// 	if z.String() != "0" {
// 		t.Fatal(z.String())
// 	}
// }
//
// func TestCmp(t *testing.T) {
// 	a := parse("1/2")
// 	b := parse("1/3")
// 	if a.IsLessThan(b) {
// 		t.Fatal()
// 	}
// }
//
// func TestInt(t *testing.T) {
// 	a := parse("1/2")
// 	if a.Int() != 0 {
// 		t.Fatal()
// 	}
//
// 	t.Log(Set("1/2").AddInt(3).String())
// 	t.Log(Set("1/2").AddInt(3).Int())
//
// 	if Set("1/2").AddInt(3).Int() != 3 {
// 		t.Fatal()
// 	}
// }
//
// func TestImu(t *testing.T) {
// 	d := parse("1")
//
// 	exp := parse("4")
// 	got := d.MinusInt(1).AddInt(2).MulInt(6).QuoInt(3) // 4
//
// 	if d.String() != "4" {
// 		t.Fatal(d.String())
// 	}
//
// 	if got.IsEqual(exp) != true {
// 		t.Fatal("fatal")
// 	}
// }

// func ExampleAdd() {
// 	a := Parse("0.1")
// 	fmt.Println(a.Add(Parse("0.2")))                      // 0.3
// 	fmt.Println(a.Add(Parse("0.2")).Equal(BigRat(3, 10))) // true
// 	fmt.Println(a)                                        // 0.1
// 	// Output: 0.3
// 	// true
// 	// 0.1
// }
//

func assertEqual[T comparable](t *testing.T, a T, b T) {
	if a != b {
		t.Fatalf("assert fail %v %v", a, b)
	}
}

// Test Rat constructor with various types
func TestRatConstructor(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		// Integer types
		{"int", 10, "10"},
		{"int8", int8(10), "10"},
		{"int16", int16(10), "10"},
		{"int32", int32(10), "10"},
		{"int64", int64(10), "10"},
		{"negative int", -10, "-10"},
		
		// Float types
		{"float32", float32(10.5), "10.5"},
		{"float64", float64(10.5), "10.5"},
		{"float with decimals", 0.1, "0.10000000"},
		{"negative float", -10.5, "-10.5"},
		
		// String types
		{"string integer", "10", "10"},
		{"string decimal", "10.5", "10.5"},
		{"string fraction", "1/2", "0.5"},
		{"string fraction complex", "3/4", "0.75"},
		{"string percentage", "50%", "0.5"},
		{"string percentage decimal", "12.5%", "0.125"},
		
		// Rational type
		{"rational", Rat("10.5"), "10.5"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Rat(tt.input)
			if r == nil {
				t.Fatalf("Rat(%v) returned nil", tt.input)
			}
			if r.String() != tt.expected {
				t.Errorf("Rat(%v).String() = %s, want %s", tt.input, r.String(), tt.expected)
			}
		})
	}
	
	// Test nil case
	t.Run("unsupported type", func(t *testing.T) {
		r := Rat(struct{}{})
		if r != nil {
			t.Errorf("Rat(struct{}{}) = %v, want nil", r)
		}
	})
}

// Test arithmetic operations comprehensively
func TestArithmeticOperations(t *testing.T) {
	t.Run("Add multiple values", func(t *testing.T) {
		tests := []struct {
			name     string
			a        any
			b        []any
			expected string
		}{
			{"integers", "10", []any{5}, "15"},
			{"decimals", "0.1", []any{"0.2"}, "0.3"},
			{"fractions", "1/4", []any{"1/4"}, "0.5"},
			{"mixed types", "0.1", []any{"1/10", "0.1"}, "0.3"},
			{"multiple args", "1", []any{2, 3, 4}, "10"},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r := Rat(tt.a).Add(tt.b...)
				if r.String() != tt.expected {
					t.Errorf("Rat(%v).Add(%v) = %s, want %s", tt.a, tt.b, r.String(), tt.expected)
				}
			})
		}
	})
	
	t.Run("Sub multiple values", func(t *testing.T) {
		tests := []struct {
			name     string
			a        any
			b        []any
			expected string
		}{
			{"integers", "10", []any{5}, "5"},
			{"decimals", "0.3", []any{"0.1"}, "0.2"},
			{"fractions", "3/4", []any{"1/4"}, "0.5"},
			{"negative result", "5", []any{10}, "-5"},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r := Rat(tt.a).Sub(tt.b...)
				if r.String() != tt.expected {
					t.Errorf("Rat(%v).Sub(%v) = %s, want %s", tt.a, tt.b, r.String(), tt.expected)
				}
			})
		}
	})
	
	t.Run("Mul operations", func(t *testing.T) {
		tests := []struct {
			name     string
			a        any
			b        any
			expected string
		}{
			{"integers", "10", 5, "50"},
			{"decimals", "0.1", "0.1", "0.01"},
			{"fractions", "1/2", "1/2", "0.25"},
			{"by zero", "10", 0, "0"},
			{"negative", "-5", 3, "-15"},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r := Rat(tt.a).Mul(tt.b)
				if r.String() != tt.expected {
					t.Errorf("Rat(%v).Mul(%v) = %s, want %s", tt.a, tt.b, r.String(), tt.expected)
				}
			})
		}
	})
	
	t.Run("Quo operations", func(t *testing.T) {
		tests := []struct {
			name     string
			a        any
			b        any
			expected string
		}{
			{"integers", "10", 5, "2"},
			{"decimals", "0.1", "0.1", "1"},
			{"fractions", "1/2", "1/4", "2"},
			{"non-exact", "10", 3, "3.33333333"},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r := Rat(tt.a).Quo(tt.b)
				if r.String() != tt.expected {
					t.Errorf("Rat(%v).Quo(%v) = %s, want %s", tt.a, tt.b, r.String(), tt.expected)
				}
			})
		}
	})
	
	t.Run("Neg operations", func(t *testing.T) {
		tests := []struct {
			name     string
			a        any
			expected string
		}{
			{"positive", "10", "-10"},
			{"negative", "-10", "10"},
			{"decimal", "10.5", "-10.5"},
			{"zero", "0", "0"},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r := Rat(tt.a).Neg()
				if r.String() != tt.expected {
					t.Errorf("Rat(%v).Neg() = %s, want %s", tt.a, r.String(), tt.expected)
				}
			})
		}
	})
}

// Test comparison operations comprehensively
func TestComparisonOperations(t *testing.T) {
	t.Run("Equal", func(t *testing.T) {
		tests := []struct {
			name     string
			a        any
			b        any
			expected bool
		}{
			{"equal integers", "10", "10", true},
			{"equal decimals", "0.1", "0.1", true},
			{"equal fractions", "1/2", "0.5", true},
			{"not equal", "10", "11", false},
			{"different types equal", 10, "10", true},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Rat(tt.a).Equal(Rat(tt.b))
				if result != tt.expected {
					t.Errorf("Rat(%v).Equal(Rat(%v)) = %v, want %v", tt.a, tt.b, result, tt.expected)
				}
			})
		}
	})
	
	t.Run("Equal with different input types", func(t *testing.T) {
		tests := []struct {
			name     string
			a        any
			b        any
			expected bool
		}{
			{"rat vs int", "10", 10, true},
			{"rat vs float", "10.5", 10.5, true},
			{"rat vs string", "10", "10", true},
			{"rat vs fraction", "0.5", "1/2", true},
			{"rat vs percentage", "0.5", "50%", true},
			{"different values", "10", 11, false},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Test both directions
				result1 := Rat(tt.a).Equal(tt.b)
				result2 := Rat(tt.b).Equal(tt.a)
				if result1 != tt.expected {
					t.Errorf("Rat(%v).Equal(%v) = %v, want %v", tt.a, tt.b, result1, tt.expected)
				}
				if result2 != tt.expected {
					t.Errorf("Rat(%v).Equal(%v) = %v, want %v", tt.b, tt.a, result2, tt.expected)
				}
			})
		}
	})
	
	t.Run("Less", func(t *testing.T) {
		tests := []struct {
			name     string
			a        any
			b        any
			expected bool
		}{
			{"less integers", "10", 11, true},
			{"less decimals", "0.1", 0.2, true},
			{"not less equal", "10", 10, false},
			{"not less greater", "11", 10, false},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Rat(tt.a).Less(tt.b)
				if result != tt.expected {
					t.Errorf("Rat(%v).Less(%v) = %v, want %v", tt.a, tt.b, result, tt.expected)
				}
			})
		}
	})
	
	t.Run("Greater", func(t *testing.T) {
		tests := []struct {
			name     string
			a        any
			b        any
			expected bool
		}{
			{"greater integers", "11", 10, true},
			{"greater decimals", "0.2", 0.1, true},
			{"not greater equal", "10", 10, false},
			{"not greater less", "10", 11, false},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Rat(tt.a).Greater(tt.b)
				if result != tt.expected {
					t.Errorf("Rat(%v).Greater(%v) = %v, want %v", tt.a, tt.b, result, tt.expected)
				}
			})
		}
	})
	
	t.Run("Cmp", func(t *testing.T) {
		tests := []struct {
			name     string
			a        any
			b        any
			expected int
		}{
			{"less", "10", "11", -1},
			{"equal", "10", "10", 0},
			{"greater", "11", "10", 1},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Rat(tt.a).Cmp(Rat(tt.b))
				if result != tt.expected {
					t.Errorf("Rat(%v).Cmp(Rat(%v)) = %v, want %v", tt.a, tt.b, result, tt.expected)
				}
			})
		}
	})
}

// Test utility functions more comprehensively
func TestUtilityFunctions(t *testing.T) {
	t.Run("Round", func(t *testing.T) {
		tests := []struct {
			name     string
			input    any
			expected string
		}{
			{"round down", "10.3", "10"},
			{"round up", "10.7", "11"},
			{"exact half", "10.5", "11"},
			{"negative round down", "-10.3", "-10"},
			{"negative round up", "-10.7", "-11"},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r := Rat(tt.input).Round()
				if r.String() != tt.expected {
					t.Errorf("Rat(%v).Round() = %s, want %s", tt.input, r.String(), tt.expected)
				}
			})
		}
	})
	
	t.Run("Float64", func(t *testing.T) {
		tests := []struct {
			name     string
			input    any
			expected float64
		}{
			{"integer", "10", 10.0},
			{"decimal", "10.5", 10.5},
			{"fraction", "1/2", 0.5},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Rat(tt.input).Float64()
				if result != tt.expected {
					t.Errorf("Rat(%v).Float64() = %v, want %v", tt.input, result, tt.expected)
				}
			})
		}
	})
}

// Test other methods
func TestOtherMethods(t *testing.T) {
	t.Run("Clone", func(t *testing.T) {
		original := Rat("10.5")
		clone := original.Clone()
		
		// Check values are equal
		if !original.Equal(clone) {
			t.Errorf("Clone() value mismatch: original = %s, clone = %s", original.String(), clone.String())
		}
		
		// Modify clone and ensure original is unchanged
		modified := clone.Add("1")
		if original.Equal(modified) {
			t.Error("Clone() created a shallow copy - modifying clone affected original")
		}
		// Also verify clone and original are still equal
		if !original.Equal(clone) {
			t.Error("Original clone was modified")
		}
	})
	
	t.Run("Set", func(t *testing.T) {
		r1 := Rat("10")
		r2 := Rat("20")
		r1.Set(r2)
		
		if !r1.Equal(r2) {
			t.Errorf("Set() failed: r1 = %s, r2 = %s", r1.String(), r2.String())
		}
	})
	
	t.Run("SetPrecision", func(t *testing.T) {
		r := Rat("10").Quo("3")
		r.SetPrecision(3)
		expected := "3.333"
		if r.String() != expected {
			t.Errorf("SetPrecision(3) = %s, want %s", r.String(), expected)
		}
	})
}

// Test example from README
func TestREADMEExample(t *testing.T) {
	result := Rat("0.1").Add("1/10", "0.1").Equal(Rat("0.3"))
	if !result {
		t.Error("README example failed")
	}
}
