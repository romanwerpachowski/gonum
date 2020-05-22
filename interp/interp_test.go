// Copyright ©2020 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interp

import (
	"fmt"
	"math"
	"testing"
)

func panics(fn func()) (panicked bool, message string) {
	defer func() {
		r := recover()
		panicked = r != nil
		message = fmt.Sprint(r)
	}()
	fn()
	return
}

func TestNewConstInterpolator1D(t *testing.T) {
	t.Parallel()
	const value = 42.0
	i1d := NewConstInterpolator1D(value)
	if i1d.begin() != math.Inf(-1) {
		t.Errorf("unexpected begin() value: got: %g want: %g", i1d.begin(), math.Inf(-1))
	}
	if i1d.end() != math.Inf(1) {
		t.Errorf("unexpected end() value: got: %g want: %g", i1d.end(), math.Inf(1))
	}
}

func TestConstInterpolator1DEval(t *testing.T) {
	t.Parallel()
	const value = 42.0
	i1d := NewConstInterpolator1D(value)
	xs := [...]float64{math.Inf(-1), -11, 0.4, 1e9, math.Inf(1)}
	for _, x := range xs {
		y := i1d.Eval(x)
		if y != value {
			t.Errorf("unexpected Eval(%g) value: got: %g want: %g", x, y, value)
		}
	}
}

func TestFindSegment(t *testing.T) {
	t.Parallel()
	xs := []float64{0, 1, 2}
	type params struct {
		x         float64
		expectedI int
		expectedX float64
	}
	paramSets := [...]params{{0, 0, 0}, {0.3, 0, 0}, {1, 1, 1}, {1.5, 1, 1}, {2, 2, 2}}
	for _, param := range paramSets {
		i, x := findSegment(xs, param.x)
		if i != param.expectedI || x != param.expectedX {
			t.Errorf("unexpected value of findSegment(xs, %g): got %d, %g want: %d, %g", param.x, i, x, param.expectedI, param.expectedX)
		}
	}
	panicXs := [...]float64{-0.5, 2.1}
	expectedMessages := [...]string{
		"interp: x value -0.5 below lower bound 0",
		"interp: x value 2.1 above upper bound 2",
	}
	for i, x := range panicXs {
		panicked, message := panics(func() { findSegment(xs, x) })
		if !panicked || message != expectedMessages[i] {
			t.Errorf("expected panic with message '%s' for evaluating at invalid x: %g", expectedMessages[i], x)
		}
	}
}

func BenchmarkFindSegment(b *testing.B) {
	xs := []float64{0, 1.5, 3, 4.5, 6, 7.5, 9, 12, 13.5, 16.5}
	for i := 0; i < b.N; i++ {
		findSegment(xs, 0)
		findSegment(xs, 16.5)
		findSegment(xs, 8.25)
		findSegment(xs, 4.125)
		findSegment(xs, 13.6)
		findSegment(xs, 13.5)
		findSegment(xs, 6)
		findSegment(xs, 4.5)
	}
}

// testPiecewiseInterpolator1DCreation tests common functionality in creating piecewise 1D interpolators.
func testPiecewiseInterpolator1DCreation(t *testing.T, create func(xs []float64, ys []float64) Interpolator1D) {
	xs := []float64{0, 1, 2}
	i1d := create(xs, []float64{-0.5, 1.5, 1})
	if xs[0] != i1d.begin() {
		t.Errorf("unexpected begin() value: got %g: want: %g", i1d.begin(), xs[0])
	}
	if xs[2] != i1d.end() {
		t.Errorf("unexpected end() value: got %g: want: %g", i1d.end(), xs[2])
	}
	type panicParams struct {
		xs              []float64
		ys              []float64
		expectedMessage string
	}
	panicParamSets := [...]panicParams{
		{xs, []float64{-0.5, 1.5}, "xs and ys have different lengths"},
		{[]float64{0.3}, []float64{0}, "too few points for interpolation"},
		{[]float64{0.3, 0.3}, []float64{0, 0}, "xs values not strictly increasing"},
		{[]float64{0.3, -0.3}, []float64{0, 0}, "xs values not strictly increasing"},
	}
	for _, params := range panicParamSets {
		panicked, message := panics(func() { create(params.xs, params.ys) })
		expectedMessage := fmt.Sprintf("interp: %s", params.expectedMessage)
		if !panicked || message != expectedMessage {
			t.Errorf("expected panic for xs: %v and ys: %v with message: %s", params.xs, params.ys, expectedMessage)
		}
	}
}

func TestNewLinearInterpolator1D(t *testing.T) {
	t.Parallel()
	testPiecewiseInterpolator1DCreation(t, func(xs []float64, ys []float64) Interpolator1D { return NewLinearInterpolator1D(xs, ys) })
}

// testInterpolator1DEval tests evaluation of a 1D interpolator.
func testInterpolator1DEval(t *testing.T, i1d Interpolator1D, xs []float64, expectedYs []float64, tol float64) {
	for i, x := range xs {
		y := i1d.Eval(x)
		yErr := math.Abs(y - expectedYs[i])
		if yErr > tol {
			if tol == 0 {
				t.Errorf("unexpected Eval(%g) value: got: %g want: %g", x, y, expectedYs[i])
			} else {
				t.Errorf("unexpected Eval(%g) value: got: %g want: %g with tolerance: %g", x, y, expectedYs[i], tol)
			}
		}
	}
}

func TestLinearInterpolator1DEval(t *testing.T) {
	t.Parallel()
	xs := []float64{0, 1, 2}
	ys := []float64{-0.5, 1.5, 1}
	i1d := NewLinearInterpolator1D(xs, ys)
	testInterpolator1DEval(t, i1d, xs, ys, 0)
	testXs := []float64{0.1, 0.5, 0.8, 1.2}
	expectedYs := []float64{-0.3, 0.5, 1.1, 1.4}
	testInterpolator1DEval(t, i1d, testXs, expectedYs, 1e-15)
}

func BenchmarkNewLinearInterpolator1D(b *testing.B) {
	xs := []float64{0, 1.5, 3, 4.5, 6, 7.5, 9, 12, 13.5, 16.5}
	ys := []float64{0, 1, 2, 2.5, 2, 1.5, 4, 10, -2, 2}
	for i := 0; i < b.N; i++ {
		NewLinearInterpolator1D(xs, ys)
	}
}

func BenchmarkLinearInterpolator1DEval(b *testing.B) {
	xs := []float64{0, 1.5, 3, 4.5, 6, 7.5, 9, 12, 13.5, 16.5}
	ys := []float64{0, 1, 2, 2.5, 2, 1.5, 4, 10, -2, 2}
	i1d := NewLinearInterpolator1D(xs, ys)
	for i := 0; i < b.N; i++ {
		i1d.Eval(0)
		i1d.Eval(16.5)
		i1d.Eval(4)
		i1d.Eval(7.32)
		i1d.Eval(9.0001)
		i1d.Eval(1.4)
		i1d.Eval(1.6)
		i1d.Eval(13.5)
		i1d.Eval(4.5)
	}
}

func TestNewPiecewiseConstInterpolator1D(t *testing.T) {
	testPiecewiseInterpolator1DCreation(t, func(xs []float64, ys []float64) Interpolator1D { return NewPiecewiseConstInterpolator1D(xs, ys, true) })
	testPiecewiseInterpolator1DCreation(t, func(xs []float64, ys []float64) Interpolator1D { return NewPiecewiseConstInterpolator1D(xs, ys, false) })
}

func benchmarkPiecewiseConstInterpolator1DEval(b *testing.B, leftContinuous bool) {
	xs := []float64{0, 1.5, 3, 4.5, 6, 7.5, 9, 12, 13.5, 16.5}
	ys := []float64{0, 1, 2, 2.5, 2, 1.5, 4, 10, -2, 2}
	i1d := NewPiecewiseConstInterpolator1D(xs, ys, leftContinuous)
	for i := 0; i < b.N; i++ {
		i1d.Eval(0)
		i1d.Eval(16.5)
		i1d.Eval(4)
		i1d.Eval(7.32)
		i1d.Eval(9.0001)
		i1d.Eval(1.4)
		i1d.Eval(1.6)
		i1d.Eval(13.5)
		i1d.Eval(4.5)
	}
}

func BenchmarkPiecewiseConstInterpolator1DLeftContinuousEval(b *testing.B) {
	benchmarkPiecewiseConstInterpolator1DEval(b, true)
}

func BenchmarkPiecewiseConstInterpolator1DRightContinuousEval(b *testing.B) {
	benchmarkPiecewiseConstInterpolator1DEval(b, false)
}

func TestPiecewiseConstInterpolator1DEval(t *testing.T) {
	t.Parallel()
	xs := []float64{0, 1, 2}
	ys := []float64{-0.5, 1.5, 1}
	iLeft := NewPiecewiseConstInterpolator1D(xs, ys, true)
	iRight := NewPiecewiseConstInterpolator1D(xs, ys, false)
	testInterpolator1DEval(t, iLeft, xs, ys, 0)
	testInterpolator1DEval(t, iRight, xs, ys, 0)
	testXs := []float64{0.1, 0.5, 0.8, 1.2}
	leftYs := []float64{1.5, 1.5, 1.5, 1}
	rightYs := []float64{-0.5, -0.5, -0.5, 1.5}
	testInterpolator1DEval(t, iLeft, testXs, leftYs, 0)
	testInterpolator1DEval(t, iRight, testXs, rightYs, 0)
}
