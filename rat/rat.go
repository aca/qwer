package rat

import (
	"database/sql/driver"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
)

var DefaultPrecision = 8

type Rational struct {
	bigrat    big.Rat
	precision int
}

func RatMin(first any, args ...any) *Rational {
	firstrat := Rat(first)
	if len(args) == 0 {
		return firstrat
	}
	out := firstrat.Clone()
	for _, arg := range args {
		if out.Greater(arg) {
			out.Set(Rat(arg))
		}
	}
	return out
}

func RatMax(first any, args ...any) *Rational {
	firstrat := Rat(first)
	if len(args) == 0 {
		return firstrat
	}
	out := firstrat.Clone()
	for _, arg := range args {
		if out.Less(arg) {
			out.Set(Rat(arg))
		}
	}
	return out
}

func Rat(v any) *Rational {
	switch v := any(v).(type) {
	case float32:
		return parseFloat64(float64(v))
	case float64:
		return parseFloat64(float64(v))
	case int:
		return parseInt64(int64(v))
	case int8:
		return parseInt64(int64(v))
	case int16:
		return parseInt64(int64(v))
	case int32:
		return parseInt64(int64(v))
	case int64:
		return parseInt64(int64(v))
	case string:
		return parse(v)
	case *Rational:
		return v
	}
	return nil
}

func (r *Rational) UnmarshalJSON(data []byte) error {
	sdata := string(data)
	sdata = strings.ReplaceAll(sdata, "\"", "")
	*r = *Rat(sdata)
	return nil
}

func (r *Rational) Int() int {
	return r.FloorInt()
}

func (r *Rational) FloorInt() int {
	// Floor function: rounds DOWN to the nearest integer
	// Examples:
	//   floor(3.7) = 3      (rounds down)
	//   floor(3.0) = 3      (already integer)
	//   floor(-3.7) = -4    (rounds down, NOT -3!)
	//   floor(-3.0) = -3    (already integer)
	
	// IMPORTANT: We cannot use float64 conversion because it loses precision!
	// Example: -3.000000000000000000001 would become -3.0 in float64,
	// giving us floor(-3.0) = -3, but the correct answer is -4!
	
	// Step 1: Get the numerator and denominator of our rational number
	// If we have -3.1, it's stored as -31/10
	num := r.bigrat.Num()      // numerator (-31)
	denom := r.bigrat.Denom()   // denominator (10)
	
	// Step 2: Do integer division (this truncates towards zero)
	// -31 ÷ 10 = -3 (remainder -1, but Go's Div ignores remainder)
	// This is NOT floor yet! It's truncation.
	quotient := new(big.Int)
	quotient.Div(num, denom)
	
	// Step 3: Check if there was a remainder
	// We do this by calculating: quotient × denominator
	// If this equals numerator, there was no remainder
	temp := new(big.Int)
	temp.Mul(quotient, denom)  // -3 × 10 = -30
	
	// Step 4: Compare original with our multiplication result
	// If num < temp, we have a negative number with a remainder
	// Example: -31 < -30, so -3.1 had a remainder
	if num.Cmp(temp) < 0 {
		// For negative numbers with ANY remainder (even tiny ones),
		// we must subtract 1 to get the floor
		// -3 - 1 = -4, which is floor(-3.1)
		quotient.Sub(quotient, big.NewInt(1))
	}
	// Note: For positive numbers, truncation = floor, so no adjustment needed
	// For exact integers (no remainder), no adjustment needed
	
	// Step 5: Convert our big.Int result to regular int
	// WARNING: This can overflow in several ways:
	// 1. If quotient > MaxInt64, Int64() returns MaxInt64 (incorrect)
	// 2. If quotient < MinInt64, Int64() returns MinInt64 (incorrect)  
	// 3. On 32-bit systems, int(int64) can overflow if value > MaxInt32
	//
	// For most practical uses, the result should fit in an int.
	// If you need to handle very large numbers, consider using:
	// - A method that returns *big.Int directly
	// - A method that returns (int, bool) where bool indicates overflow
	// - Checking quotient.IsInt64() before conversion
	return int(quotient.Int64())
}

func (r *Rational) IntString() string {
	return fmt.Sprint(r.FloorInt())
}

func (r *Rational) Round() *Rational {
	return Rat(r.Add(Rat("0.5")).FloorInt())
}

func (r *Rational) Ceil() *Rational {
	if r.bigrat.IsInt() {
		return r.Clone()
	} else {
		// For non-integers, Ceil(x) = Floor(x) + 1
		return Rat(r.FloorInt() + 1)
	}
}

func (r *Rational) Floor() *Rational {
	if r.bigrat.IsInt() {
		return r.Clone()
	} else {
		// For non-integers, return the floor as a Rational
		// Example: floor(-7.004) = -8 (rounds down toward negative infinity)
		return Rat(r.FloorInt())
	}
}

func (r *Rational) Float64() float64 {
	out, _ := r.bigrat.Float64()
	return out
}

func (r *Rational) Sub(ins ...any) *Rational {
	out := r.Clone()

	for _, in := range ins {
		switch v := in.(type) {
		case int:
			// Optimize: convert int directly to big.Rat without creating Rational
			var temp big.Rat
			temp.SetInt64(int64(v))
			out.bigrat.Sub(&out.bigrat, &temp)
		case int8:
			var temp big.Rat
			temp.SetInt64(int64(v))
			out.bigrat.Sub(&out.bigrat, &temp)
		case int16:
			var temp big.Rat
			temp.SetInt64(int64(v))
			out.bigrat.Sub(&out.bigrat, &temp)
		case int32:
			var temp big.Rat
			temp.SetInt64(int64(v))
			out.bigrat.Sub(&out.bigrat, &temp)
		case int64:
			var temp big.Rat
			temp.SetInt64(v)
			out.bigrat.Sub(&out.bigrat, &temp)
		case float32:
			var temp big.Rat
			temp.SetFloat64(float64(v))
			out.bigrat.Sub(&out.bigrat, &temp)
		case float64:
			var temp big.Rat
			temp.SetFloat64(v)
			out.bigrat.Sub(&out.bigrat, &temp)
		case string:
			out.bigrat.Sub(&out.bigrat, &Rat(v).bigrat)
		case *Rational:
			out.bigrat.Sub(&out.bigrat, &v.bigrat)
		default:
			slog.Error("rat: sub invalid type")
			return nil
		}
	}
	return out
}

func (r *Rational) Add(ins ...any) *Rational {
	out := r.Clone()

	for _, in := range ins {
		switch v := in.(type) {
		case int:
			// Optimize: convert int directly to big.Rat without creating Rational
			var temp big.Rat
			temp.SetInt64(int64(v))
			out.bigrat.Add(&out.bigrat, &temp)
		case int8:
			var temp big.Rat
			temp.SetInt64(int64(v))
			out.bigrat.Add(&out.bigrat, &temp)
		case int16:
			var temp big.Rat
			temp.SetInt64(int64(v))
			out.bigrat.Add(&out.bigrat, &temp)
		case int32:
			var temp big.Rat
			temp.SetInt64(int64(v))
			out.bigrat.Add(&out.bigrat, &temp)
		case int64:
			var temp big.Rat
			temp.SetInt64(v)
			out.bigrat.Add(&out.bigrat, &temp)
		case float32:
			var temp big.Rat
			temp.SetFloat64(float64(v))
			out.bigrat.Add(&out.bigrat, &temp)
		case float64:
			var temp big.Rat
			temp.SetFloat64(v)
			out.bigrat.Add(&out.bigrat, &temp)
		case string:
			out.bigrat.Add(&out.bigrat, &Rat(v).bigrat)
		case *Rational:
			out.bigrat.Add(&out.bigrat, &v.bigrat)
		default:
			slog.Error("rat: add invalid type")
			return nil
		}
	}
	return out
}

func (r *Rational) Neg() *Rational {
	out := r.Clone()
	out.bigrat.Neg(&r.bigrat)
	return out
}

func (r *Rational) Mul(in any) *Rational {
	out := r.Clone()

	switch v := in.(type) {
	case int:
		// Optimize: convert int directly to big.Rat without creating Rational
		var temp big.Rat
		temp.SetInt64(int64(v))
		out.bigrat.Mul(&out.bigrat, &temp)
	case int8:
		var temp big.Rat
		temp.SetInt64(int64(v))
		out.bigrat.Mul(&out.bigrat, &temp)
	case int16:
		var temp big.Rat
		temp.SetInt64(int64(v))
		out.bigrat.Mul(&out.bigrat, &temp)
	case int32:
		var temp big.Rat
		temp.SetInt64(int64(v))
		out.bigrat.Mul(&out.bigrat, &temp)
	case int64:
		var temp big.Rat
		temp.SetInt64(v)
		out.bigrat.Mul(&out.bigrat, &temp)
	case float32:
		var temp big.Rat
		temp.SetFloat64(float64(v))
		out.bigrat.Mul(&out.bigrat, &temp)
	case float64:
		var temp big.Rat
		temp.SetFloat64(v)
		out.bigrat.Mul(&out.bigrat, &temp)
	case string:
		out.bigrat.Mul(&out.bigrat, &Rat(v).bigrat)
	case *Rational:
		out.bigrat.Mul(&out.bigrat, &v.bigrat)
	default:
		panic("rat: mul invalid type")
	}
	return out
}

func (r *Rational) Quo(in any) *Rational {
	out := r.Clone()

	switch v := in.(type) {
	case int:
		// Optimize: convert int directly to big.Rat without creating Rational
		var temp big.Rat
		temp.SetInt64(int64(v))
		out.bigrat.Quo(&out.bigrat, &temp)
	case int8:
		var temp big.Rat
		temp.SetInt64(int64(v))
		out.bigrat.Quo(&out.bigrat, &temp)
	case int16:
		var temp big.Rat
		temp.SetInt64(int64(v))
		out.bigrat.Quo(&out.bigrat, &temp)
	case int32:
		var temp big.Rat
		temp.SetInt64(int64(v))
		out.bigrat.Quo(&out.bigrat, &temp)
	case int64:
		var temp big.Rat
		temp.SetInt64(v)
		out.bigrat.Quo(&out.bigrat, &temp)
	case float32:
		var temp big.Rat
		temp.SetFloat64(float64(v))
		out.bigrat.Quo(&out.bigrat, &temp)
	case float64:
		var temp big.Rat
		temp.SetFloat64(v)
		out.bigrat.Quo(&out.bigrat, &temp)
	case string:
		out.bigrat.Quo(&out.bigrat, &Rat(v).bigrat)
	case *Rational:
		out.bigrat.Quo(&out.bigrat, &v.bigrat)
	default:
		panic("rat: quo invalid type")
	}
	return out
}

func (r *Rational) PowInt(exp int) *Rational {
	if exp == 0 {
		// Any number to the power of 0 is 1
		return Rat(1)
	}

	// Optimize: use exponentiation by squaring for better performance
	if exp < 0 {
		// For negative exponents, calculate 1 / (r^|exp|)
		posResult := r.powIntPositive(-exp)
		one := Rat(1)
		return one.Quo(posResult)
	}

	return r.powIntPositive(exp)
}

// Helper function for positive integer exponentiation using binary exponentiation
func (r *Rational) powIntPositive(exp int) *Rational {
	if exp == 1 {
		return r.Clone()
	}

	result := Rat(1)
	base := r.Clone()

	// Binary exponentiation: O(log n) instead of O(n)
	for exp > 0 {
		if exp%2 == 1 {
			result = result.Mul(base)
		}
		base = base.Mul(base)
		exp /= 2
	}

	return result
}

func (r *Rational) String() string {
	n, exact := r.bigrat.FloatPrec()
	if exact {
		return r.bigrat.FloatString(min(r.precision, n))
		// return r.bigrat.FloatString(min(DefaultPrecision, n))
	}
	return r.bigrat.FloatString(r.precision)
}

func (r *Rational) Set(v *Rational) *Rational {
	r.bigrat.Set(&v.bigrat)
	return r
}

func (r *Rational) Clone() *Rational {
	newr := new(Rational)
	newr.bigrat.Set(&r.bigrat)
	newr.precision = r.precision
	return newr
}

func (r *Rational) Less(in any) bool {
	inrat := Rat(in)

	if r.bigrat.Cmp(&inrat.bigrat) == -1 {
		return true
	}
	return false
}

func (r *Rational) Greater(in any) bool {
	inrat := Rat(in)

	if r.bigrat.Cmp(&inrat.bigrat) == 1 {
		return true
	}
	return false
}

func (r *Rational) Cmp(b any) int {
	br := Rat(b)
	return r.bigrat.Cmp(&br.bigrat)
}

func (r *Rational) Equal(in any) bool {
	inrat := Rat(in)
	if inrat == nil {
		return false
	}
	return r.bigrat.Cmp(&inrat.bigrat) == 0
}

func (r *Rational) SetPrecision(v int) *Rational {
	r.precision = v
	return r
}

func RatQuo(a any, b any) *Rational {
	return Rat(a).Quo(b)
}

func RatMul(a any, b *Rational) *Rational {
	return Rat(a).Mul(b)
}

func RatAdd(a any, b *Rational) *Rational {
	return Rat(a).Add(b)
}

func RatNeg(a any) *Rational {
	ra := Rat(a)
	c := big.Rat{}
	c.Neg(&ra.bigrat)
	return &Rational{
		bigrat:    c,
		precision: ra.precision,
	}
}

func RatZero() *Rational {
	return &Rational{
		precision: DefaultPrecision,
	}
}

func RatClone(r *Rational) *Rational {
	newr := new(Rational)
	newr.bigrat.Set(&r.bigrat)
	newr.precision = r.precision
	return newr
}

func parseFloat64(a float64) *Rational {
	out := new(Rational)
	out.bigrat.SetFloat64(a)
	out.precision = DefaultPrecision
	return out
}

func parseInt64(a int64) *Rational {
	out := new(Rational)
	out.bigrat.SetInt64(a)
	out.precision = DefaultPrecision
	return out
}

func parse(v string) (out *Rational) {
	defer func() {
		if out == nil {
			slog.Error("rat: parse failed, out is nil")
		}
	}()

	if strings.HasSuffix(v, "%") {
		defer func() {
			if out == nil {
				return
			}
			// 1% => 0.01
			out = out.Quo(100)
		}()
		v = v[0 : len(v)-1]
	}

	if strings.Contains(v, "/") {
		split := strings.Split(v, "/")
		if len(split) != 2 {
			slog.Error("rat: invalid rat string " + v)
			return nil
		}

		return Rat(split[0]).Quo(Rat(split[1]))
	}

	out = new(Rational)
	_, ok := out.bigrat.SetString(v)
	if !ok {
		return nil
	}
	out.precision = DefaultPrecision
	return out
}

func (r *Rational) Scan(src any) error {
	r.precision = DefaultPrecision
	switch v := src.(type) {
	case string:
		r.Set(Rat(v))
		return nil
	case []byte:
		r.Set(Rat(string(v)))
		return nil
	case int32:
		r.Set(Rat(v))
		return nil
	case int64:
		r.Set(Rat(v))
		return nil
	case float32:
		r.Set(Rat(v))
		return nil
	case float64:
		r.Set(Rat(v))
		return nil
	default:
		return fmt.Errorf("rat: scan err: invalid type %T", src)
	}
}

func (r *Rational) Value() (driver.Value, error) {
	return r.bigrat.RatString(), nil
}

func (r *Rational) GobEncode() ([]byte, error) {
	return r.bigrat.GobEncode()
}

func (r *Rational) GobDecode(buf []byte) error {
	r.precision = DefaultPrecision
	return r.bigrat.GobDecode(buf)
}
