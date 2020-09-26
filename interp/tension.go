// Copyright ©2020 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interp

import (
	"fmt"
	"math"
)

// ExponentialTensionSpline is a piecewise exponential spline
// with continuous value, first and second derivative.
// Based on "Interest Rate Modelling. Volume I: Foundations and Vanilla Models",
// Leif B.G. Andersen and Vladimir V. Piterbarg (2010), Sec. 6.2.4.
type ExponentialTensionSpline struct {
	// Tension parameter, sigma > 0. sigma ~= 0 corresponds to a natural C^2
	// spline, while sigma >> 0 results in piecewise-linear interpolation.
	sigma float64

	// Interpolated X values.
	xs []float64

	// Interpolated Y values.
	ys []float64

	// Interpolated d^2Y / dX^2 values.
	d2ydx2s []float64
}

// Predict returns the interpolation value at x.
func (ets *ExponentialTensionSpline) Predict(x float64) float64 {
	if ets.sigma <= 0 {
		panic(fmt.Sprintf("Sigma must be positive, got %g", ets.sigma))
	}
	i := findSegment(ets.xs, x)
	if i < 0 {
		return ets.ys[0]
	}
	if x == ets.xs[i] {
		return ets.ys[i]
	}
	m := len(ets.xs) - 1
	if i == m {
		return ets.ys[m]
	}
	dx := x - ets.xs[i]
	h := ets.xs[i+1] - ets.xs[i]
	s := math.Sinh(ets.sigma * h)
	sl := math.Sinh(ets.sigma * dx)
	sr := math.Sinh(ets.sigma * (h - dx))
	rl := dx / h
	rr := 1 - rl
	return ((sr/s-rr)*ets.d2ydx2s[i]-(sl/s-rl)*ets.d2ydx2s[i+1])/ets.sigma/ets.sigma + ets.ys[i]*rr + ets.ys[i+1]*rl
}

// Fit fits a predictor to (X, Y) value pairs provided as two slices.
// It panics if ets.sigma <= 0, len(xs) < 2, elements of xs are not
// strictly increasing or len(xs) != len(ys). Always returns nil.
func (ets *ExponentialTensionSpline) Fit(xs, ys []float64) error {
	if ets.sigma <= 0 {
		panic(fmt.Sprintf("Sigma must be positive, got %g", ets.sigma))
	}
	n := len(xs)
	if len(ys) != n {
		panic(differentLengths)
	}
	if n < 2 {
		panic(tooFewPoints)
	}
	for i := 1; i < n; i++ {
		if xs[i] <= xs[i-1] {
			panic(xsNotStrictlyIncreasing)
		}
	}
	ets.xs = make([]float64, n)
	ets.ys = make([]float64, n)
	copy(ets.xs, xs)
	copy(ets.ys, ys)
	return nil
}
