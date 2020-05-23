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

func TestConstant1DInterval(t *testing.T) {
	t.Parallel()
	const value = 42.0
	ci := Constant1D{value}
	interval := ci.Interval()
	if !math.IsInf(interval.Min, -1) {
		t.Errorf("unexpected begin() value: got: %g want: %g", interval.Min, math.Inf(-1))
	}
	if !math.IsInf(interval.Max, 1) {
		t.Errorf("unexpected end() value: got: %g want: %g", interval.Max, math.Inf(1))
	}
}

func TestConstant1DValueAt(t *testing.T) {
	t.Parallel()
	const value = 42.0
	ci := Constant1D{value}
	xs := []float64{math.Inf(-1), -11, 0.4, 1e9, math.Inf(1)}
	for _, x := range xs {
		y := ci.ValueAt(x)
		if y != value {
			t.Errorf("unexpected ValueAt(%g) value: got: %g want: %g", x, y, value)
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
	paramSets := []params{{0, 0, 0}, {0.3, 0, 0}, {1, 1, 1}, {1.5, 1, 1}, {2, 2, 2}}
	for _, param := range paramSets {
		i, x := findSegment(xs, param.x)
		if i != param.expectedI || x != param.expectedX {
			t.Errorf("unexpected value of findSegment(xs, %g): got %d, %g want: %d, %g", param.x, i, x, param.expectedI, param.expectedX)
		}
	}
	panicXs := []float64{-0.5, 2.1}
	expectedMessages := []string{
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
	interval := i1d.Interval()
	if xs[0] != interval.Min {
		t.Errorf("unexpected begin() value: got %g: want: %g", interval.Min, xs[0])
	}
	if xs[2] != interval.Max {
		t.Errorf("unexpected end() value: got %g: want: %g", interval.Max, xs[2])
	}
	type panicParams struct {
		xs              []float64
		ys              []float64
		expectedMessage string
	}
	panicParamSets := []panicParams{
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

func TestNewLinear1D(t *testing.T) {
	t.Parallel()
	testPiecewiseInterpolator1DCreation(t, func(xs []float64, ys []float64) Interpolator1D { return NewLinear1D(xs, ys) })
}

// testInterpolator1DValueAt tests evaluation of a 1D interpolator.
func testInterpolator1DValueAt(t *testing.T, i1d Interpolator1D, xs []float64, expectedYs []float64, tol float64) {
	for i, x := range xs {
		y := i1d.ValueAt(x)
		yErr := math.Abs(y - expectedYs[i])
		if yErr > tol {
			if tol == 0 {
				t.Errorf("unexpected ValueAt(%g) value: got: %g want: %g", x, y, expectedYs[i])
			} else {
				t.Errorf("unexpected ValueAt(%g) value: got: %g want: %g with tolerance: %g", x, y, expectedYs[i], tol)
			}
		}
	}
}

func TestLinear1DValueAt(t *testing.T) {
	t.Parallel()
	xs := []float64{0, 1, 2}
	ys := []float64{-0.5, 1.5, 1}
	li := NewLinear1D(xs, ys)
	testInterpolator1DValueAt(t, li, xs, ys, 0)
	testXs := []float64{0.1, 0.5, 0.8, 1.2}
	expectedYs := []float64{-0.3, 0.5, 1.1, 1.4}
	testInterpolator1DValueAt(t, li, testXs, expectedYs, 1e-15)
}

func BenchmarkNewLinear1D(b *testing.B) {
	xs := []float64{0, 1.5, 3, 4.5, 6, 7.5, 9, 12, 13.5, 16.5}
	ys := []float64{0, 1, 2, 2.5, 2, 1.5, 4, 10, -2, 2}
	for i := 0; i < b.N; i++ {
		NewLinear1D(xs, ys)
	}
}

func BenchmarkLinear1DValueAt(b *testing.B) {
	xs := []float64{0, 1.5, 3, 4.5, 6, 7.5, 9, 12, 13.5, 16.5}
	ys := []float64{0, 1, 2, 2.5, 2, 1.5, 4, 10, -2, 2}
	li := NewLinear1D(xs, ys)
	for i := 0; i < b.N; i++ {
		li.ValueAt(0)
		li.ValueAt(16.5)
		li.ValueAt(4)
		li.ValueAt(7.32)
		li.ValueAt(9.0001)
		li.ValueAt(1.4)
		li.ValueAt(1.6)
		li.ValueAt(13.5)
		li.ValueAt(4.5)
	}
}

func TestNewPiecewiseConstant1D(t *testing.T) {
	testPiecewiseInterpolator1DCreation(t, func(xs []float64, ys []float64) Interpolator1D { return NewPiecewiseConstant1D(xs, ys, true) })
	testPiecewiseInterpolator1DCreation(t, func(xs []float64, ys []float64) Interpolator1D { return NewPiecewiseConstant1D(xs, ys, false) })
}

func benchmarkPiecewiseConstant1DValueAt(b *testing.B, leftContinuous bool) {
	xs := []float64{0, 1.5, 3, 4.5, 6, 7.5, 9, 12, 13.5, 16.5}
	ys := []float64{0, 1, 2, 2.5, 2, 1.5, 4, 10, -2, 2}
	pci := NewPiecewiseConstant1D(xs, ys, leftContinuous)
	for i := 0; i < b.N; i++ {
		pci.ValueAt(0)
		pci.ValueAt(16.5)
		pci.ValueAt(4)
		pci.ValueAt(7.32)
		pci.ValueAt(9.0001)
		pci.ValueAt(1.4)
		pci.ValueAt(1.6)
		pci.ValueAt(13.5)
		pci.ValueAt(4.5)
	}
}

func BenchmarkPiecewiseConstant1DLeftContinuousValueAt(b *testing.B) {
	benchmarkPiecewiseConstant1DValueAt(b, true)
}

func BenchmarkPiecewiseConstant1DRightContinuousValueAt(b *testing.B) {
	benchmarkPiecewiseConstant1DValueAt(b, false)
}

func TestPiecewiseConstant1DValueAt(t *testing.T) {
	t.Parallel()
	xs := []float64{0, 1, 2}
	ys := []float64{-0.5, 1.5, 1}
	pciLeft := NewPiecewiseConstant1D(xs, ys, true)
	pciRight := NewPiecewiseConstant1D(xs, ys, false)
	testInterpolator1DValueAt(t, pciLeft, xs, ys, 0)
	testInterpolator1DValueAt(t, pciRight, xs, ys, 0)
	testXs := []float64{0.1, 0.5, 0.8, 1.2}
	leftYs := []float64{1.5, 1.5, 1.5, 1}
	rightYs := []float64{-0.5, -0.5, -0.5, 1.5}
	testInterpolator1DValueAt(t, pciLeft, testXs, leftYs, 0)
	testInterpolator1DValueAt(t, pciRight, testXs, rightYs, 0)
}
