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

func RatMax(first *Rational, args ...*Rational) *Rational {
	if len(args) == 0 {
		return first
	}
	ret := first.Clone()
	for _, arg := range args {
		if ret.Less(arg) {
			ret.Set(arg)
		}
	}
	return ret
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

func (r *Rational) FloorInt() int {
	// Optimize: use big.Rat methods directly instead of string conversion
	if r.bigrat.IsInt() {
		// For integers, convert directly
		f, _ := r.bigrat.Float64()
		return int(f)
	}
	
	// For non-integers, we need proper floor division
	// The issue is that big.Int.Div truncates towards zero, not towards negative infinity
	
	// Simple approach: use the mathematical definition
	// Floor(x) = x - (x mod 1) if x >= 0
	// Floor(x) = x - (x mod 1) - 1 if x < 0 and (x mod 1) != 0
	
	f, _ := r.bigrat.Float64()
	if f >= 0 {
		return int(f) // For positive numbers, truncation equals floor
	} else {
		// For negative numbers, check if there's a fractional part
		truncated := int(f)
		if float64(truncated) == f {
			return truncated // It's exactly an integer
		} else {
			return truncated - 1 // Floor of negative non-integer
		}
	}
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
		// TODO: fix this, math.ceil should be math.cel(-7.004) should be -7
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
		out.bigrat.Mul(&r.bigrat, &Rat(v).bigrat)
	case int32:
		out.bigrat.Mul(&r.bigrat, &Rat(v).bigrat)
	case int64:
		out.bigrat.Mul(&r.bigrat, &Rat(v).bigrat)
	case float32:
		out.bigrat.Mul(&r.bigrat, &Rat(v).bigrat)
	case float64:
		out.bigrat.Mul(&r.bigrat, &Rat(v).bigrat)
	case string:
		out.bigrat.Mul(&r.bigrat, &Rat(v).bigrat)
	case *Rational:
		out.bigrat.Mul(&r.bigrat, &v.bigrat)
	default:
		panic("rat: add invalid type")
	}
	return out
}

func (r *Rational) Quo(in any) *Rational {
	out := r.Clone()

	switch v := in.(type) {
	case int:
		out.bigrat.Quo(&r.bigrat, &Rat(v).bigrat)
	case int32:
		out.bigrat.Quo(&r.bigrat, &Rat(v).bigrat)
	case int64:
		out.bigrat.Quo(&r.bigrat, &Rat(v).bigrat)
	case float32:
		out.bigrat.Quo(&r.bigrat, &Rat(v).bigrat)
	case float64:
		out.bigrat.Quo(&r.bigrat, &Rat(v).bigrat)
	case string:
		out.bigrat.Quo(&r.bigrat, &Rat(v).bigrat)
	case *Rational:
		out.bigrat.Quo(&r.bigrat, &v.bigrat)
	default:
		panic("rat: add invalid type")
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
	inrat := new(Rational)
	switch v := in.(type) {
	case int:
		inrat = Rat(v)
	case int8:
		inrat = Rat(v)
	case int32:
		inrat = Rat(v)
	case int64:
		inrat = Rat(v)
	case float32:
		inrat = Rat(v)
	case float64:
		inrat = Rat(v)
	case string:
		inrat = Rat(v)
	case *Rational:
		inrat = v
	}

	if r.bigrat.Cmp(&inrat.bigrat) == -1 {
		return true
	}
	return false
}

func (r *Rational) Greater(in any) bool {
	inrat := new(Rational)
	switch v := in.(type) {
	case int:
		inrat = Rat(v)
	case int8:
		inrat = Rat(v)
	case int32:
		inrat = Rat(v)
	case int64:
		inrat = Rat(v)
	case float32:
		inrat = Rat(v)
	case float64:
		inrat = Rat(v)
	case string:
		inrat = Rat(v)
	case *Rational:
		inrat = v
	}

	if r.bigrat.Cmp(&inrat.bigrat) == 1 {
		return true
	}
	return false
}

func (r *Rational) Cmp(b *Rational) int {
	return r.bigrat.Cmp(&b.bigrat)
}

func (r *Rational) Equal(in any) bool {
	inrat := new(Rational)
	switch v := in.(type) {
	case int:
		inrat = Rat(v)
	case int8:
		inrat = Rat(v)
	case int32:
		inrat = Rat(v)
	case int64:
		inrat = Rat(v)
	case float32:
		inrat = Rat(v)
	case float64:
		inrat = Rat(v)
	case string:
		inrat = Rat(v)
	case *Rational:
		inrat = v
	}

	if r.bigrat.Cmp(&inrat.bigrat) == 0 {
		return true
	}
	return false
}

func (r *Rational) SetPrecision(v int) *Rational {
	r.precision = v
	return r
}

func RatQuo(a *Rational, b *Rational) *Rational {
	return a.Quo(b)
}

func RatMul(a *Rational, b *Rational) *Rational {
	return a.Mul(b)
}

func RatAdd(a *Rational, b *Rational) *Rational {
	return a.Add(b)
}

func RatNeg(a *Rational) *Rational {
	c := big.Rat{}
	c.Neg(&a.bigrat)
	return &Rational{
		bigrat:    c,
		precision: a.precision,
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
