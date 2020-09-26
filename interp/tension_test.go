// Copyright ©2020 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interp

import "testing"

func TestPredictErrors(t *testing.T) {
	t.Parallel()
	for i, test := range []struct {
		sigma           float64
		xs, ys, d2ydx2s []float64
	}{
		{
			sigma:   -0.5,
			xs:      []float64{0, 1},
			ys:      []float64{4, 1},
			d2ydx2s: []float64{0, 1},
		},
		{
			sigma:   0,
			xs:      []float64{0, 1},
			ys:      []float64{4, 1},
			d2ydx2s: []float64{0, 1},
		},
	} {
		ets := ExponentialTensionSpline{test.sigma, test.xs, test.ys, test.d2ydx2s}
		if !panics(func() { ets.Predict(0.2) }) {
			t.Errorf("Expected panic in test case %d", i)
		}
	}
}

func TestFitErrors(t *testing.T) {
	t.Parallel()
	for i, test := range []struct {
		sigma  float64
		xs, ys []float64
	}{
		{
			sigma: -0.5,
			xs:    []float64{0, 1},
			ys:    []float64{4, 1},
		},
		{
			sigma: 0,
			xs:    []float64{0, 1},
			ys:    []float64{4, 1},
		},
		{
			sigma: 1,
			xs:    []float64{0},
			ys:    []float64{4},
		},
		{
			sigma: 1,
			xs:    []float64{0, 1, 2},
			ys:    []float64{4, 3},
		},
		{
			sigma: 1,
			xs:    []float64{0, 1, 1},
			ys:    []float64{4, 3, 5},
		},
	} {
		ets := ExponentialTensionSpline{sigma: test.sigma}
		var err error
		if !panics(func() { err = ets.Fit(test.xs, test.ys) }) {
			t.Errorf("Expected panic in test case %d, got %v", i, err)
		}
	}
}
