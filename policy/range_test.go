package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRangeContains(t *testing.T) {
	cases := []struct {
		desc     string
		min      int
		max      int
		value    int
		expected bool
	}{
		{
			desc:     "Above min",
			min:      10,
			value:    20,
			expected: true,
		},
		{
			desc:     "Below min",
			min:      10,
			value:    2,
			expected: false,
		},
		{
			desc:     "Below max",
			max:      10,
			value:    5,
			expected: true,
		},
		{
			desc:     "Above max",
			max:      10,
			value:    20,
			expected: false,
		},
		{
			desc:     "Between min and max",
			min:      10,
			max:      20,
			value:    15,
			expected: true,
		},
		{
			desc:     "Outside min and max",
			min:      10,
			max:      20,
			value:    25,
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			rng := Range[int]{
				Min: &tc.min,
				Max: &tc.max,
			}
			if tc.min == 0 {
				rng.Min = nil
			}
			if tc.max == 0 {
				rng.Max = nil
			}

			actual := rng.Contains(tc.value)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestRangeReason(t *testing.T) {
	cases := []struct {
		desc     string
		expected string
		min      int
		max      int
	}{
		{
			desc:     "Min",
			min:      10,
			expected: "is lower than 10",
		},
		{
			desc:     "Max",
			max:      10,
			expected: "is higher than 10",
		},
		{
			desc:     "Min and max",
			min:      10,
			max:      20,
			expected: "is not between 10 and 20",
		},
		{
			desc:     "Empty",
			expected: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			rng := Range[int]{
				Min: &tc.min,
				Max: &tc.max,
			}
			if tc.min == 0 {
				rng.Min = nil
			}
			if tc.max == 0 {
				rng.Max = nil
			}

			actual := rng.Reason()
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestCheck(t *testing.T) {
	cases := []struct {
		desc     string
		min      int
		max      int
		value    int
		expected bool
	}{
		{
			desc:     "Contains",
			min:      1,
			max:      5,
			value:    3,
			expected: true,
		},
		{
			desc:     "Equal to min",
			min:      1,
			max:      5,
			value:    1,
			expected: true,
		},
		{
			desc:     "Equal to max",
			min:      1,
			max:      5,
			value:    5,
			expected: true,
		},
		{
			desc:     "Lower than min",
			min:      1,
			max:      5,
			value:    0,
			expected: false,
		},
		{
			desc:     "Higher than max",
			min:      1,
			max:      5,
			value:    6,
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			rng := &Range[int]{
				Min: &tc.min,
				Max: &tc.max,
			}

			actual := check[int](rng, tc.value)
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("Nil", func(t *testing.T) {
		actual := check[int](nil, 0)
		assert.True(t, actual)
	})
}

func TestStatRangeContains(t *testing.T) {
	cases := []struct {
		desc      string
		operation Operation
		values    []int
		min       int
		max       int
		expected  bool
	}{
		{
			desc:      "Median",
			operation: Median,
			min:       1,
			max:       9,
			values:    []int{0, 4, 5, 6, 8},
			expected:  true,
		},
		{
			desc:      "Median min",
			operation: Median,
			min:       2,
			values:    []int{0, 4, 5, 6, 8},
			expected:  true,
		},
		{
			desc:      "Median min out",
			operation: Median,
			min:       10,
			values:    []int{0, 4, 5, 6, 8},
			expected:  false,
		},
		{
			desc:      "Median max",
			operation: Median,
			max:       9,
			values:    []int{0, 4, 5, 6, 8},
			expected:  true,
		},
		{
			desc:      "Median max out",
			operation: Median,
			max:       4,
			values:    []int{0, 4, 5, 6, 8},
			expected:  false,
		},
		{
			desc:      "Mean",
			operation: Mean,
			min:       1,
			max:       8,
			values:    []int{0, 4, 5, 6, 8},
			expected:  true,
		},
		{
			desc:      "Mean min",
			operation: Mean,
			min:       1,
			values:    []int{0, 4, 5, 6, 8},
			expected:  true,
		},
		{
			desc:      "Mean min out",
			operation: Mean,
			min:       10,
			values:    []int{0, 4, 5, 6, 8},
			expected:  false,
		},
		{
			desc:      "Mean max",
			operation: Mean,
			max:       9,
			values:    []int{0, 4, 5, 6, 8},
			expected:  true,
		},
		{
			desc:      "Mean max out",
			operation: Mean,
			max:       3,
			values:    []int{0, 4, 5, 6, 8},
			expected:  false,
		},
		{
			desc:      "Mode",
			operation: Mode,
			min:       3,
			max:       6,
			values:    []int{2, 4, 5, 5, 25, 26},
			expected:  true,
		},
		{
			desc:      "Mode min",
			operation: Mode,
			min:       11,
			values:    []int{11, 11, 13},
			expected:  true,
		},
		{
			desc:      "Mode min out",
			operation: Mode,
			min:       12,
			values:    []int{11, 11, 13},
			expected:  false,
		},
		{
			desc:      "Mode max",
			operation: Mode,
			max:       6,
			values:    []int{0, 4, 6, 6, 8},
			expected:  true,
		},
		{
			desc:      "Mode max out",
			operation: Mode,
			max:       3,
			values:    []int{0, 7, 8, 8},
			expected:  false,
		},
		{
			desc:      "Range",
			operation: RangeOp,
			min:       1,
			max:       10,
			values:    []int{0, 4, 5, 6, 8},
			expected:  true,
		},
		{
			desc:      "Range min",
			operation: RangeOp,
			min:       5,
			values:    []int{0, 11, 15},
			expected:  true,
		},
		{
			desc:      "Range min out",
			operation: RangeOp,
			min:       12,
			values:    []int{6, 11, 13},
			expected:  false,
		},
		{
			desc:      "Range max",
			operation: RangeOp,
			max:       6,
			values:    []int{1, 4, 5},
			expected:  true,
		},
		{
			desc:      "Range max out",
			operation: RangeOp,
			max:       3,
			values:    []int{0, 4},
			expected:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			statRange := StatRange[int]{
				Min:       &tc.min,
				Max:       &tc.max,
				Operation: tc.operation,
			}
			if tc.min == 0 {
				statRange.Min = nil
			}
			if tc.max == 0 {
				statRange.Max = nil
			}

			actual := statRange.Contains(tc.values)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestStatRangeReason(t *testing.T) {
	cases := []struct {
		desc      string
		expected  string
		operation Operation
		min       int
		max       int
	}{
		{
			desc:      "Min",
			operation: Mean,
			min:       10,
			expected:  "mean value is lower than 10",
		},
		{
			desc:      "Max",
			operation: Median,
			max:       10,
			expected:  "median value is higher than 10",
		},
		{
			desc:      "Min and max (mode)",
			operation: Mode,
			min:       10,
			max:       20,
			expected:  "mode value is not between 10 and 20",
		},
		{
			desc:      "Min and max (range)",
			operation: RangeOp,
			min:       5,
			max:       8,
			expected:  "range value is not between 5 and 8",
		},
		{
			desc:     "Default operation",
			expected: "mean value ",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			rng := StatRange[int]{
				Operation: tc.operation,
				Min:       &tc.min,
				Max:       &tc.max,
			}
			if tc.min == 0 {
				rng.Min = nil
			}
			if tc.max == 0 {
				rng.Max = nil
			}

			actual := rng.Reason()
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestMedian(t *testing.T) {
	cases := []struct {
		desc     string
		values   []int
		expected int
	}{
		{
			desc:     "Even number of values",
			values:   []int{1, 4, 5, 7, 8, 12},
			expected: 6,
		},
		{
			desc:     "Odd number of values",
			values:   []int{1, 4, 5, 7, 8, 12, 13},
			expected: 7,
		},
		{
			desc:     "No values",
			values:   []int{},
			expected: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := median(tc.values)
			assert.Exactly(t, tc.expected, actual)
		})
	}
}

func TestMean(t *testing.T) {
	cases := []struct {
		desc     string
		values   []int
		expected int
	}{
		{
			desc:     "Round result",
			values:   []int{4, 6, 11},
			expected: 7,
		},
		{
			desc:     "Approximate result",
			values:   []int{4, 6, 10},
			expected: 6,
		},
		{
			desc:     "No values",
			values:   []int{},
			expected: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := mean(tc.values)
			assert.Exactly(t, tc.expected, actual)
		})
	}
}

func TestMode(t *testing.T) {
	cases := []struct {
		desc     string
		values   []int
		expected int
	}{
		{
			desc:     "Mode",
			values:   []int{1, 1, 2, 5, 7, 4, 6, 1},
			expected: 1,
		},
		{
			desc:     "No values",
			values:   []int{},
			expected: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := mode(tc.values)
			assert.Exactly(t, tc.expected, actual)
		})
	}
}

func TestRangeOp(t *testing.T) {
	cases := []struct {
		desc     string
		values   []int
		expected int
	}{
		{
			desc:     "Range",
			values:   []int{2, 23},
			expected: 21,
		},
		{
			desc:     "No values",
			values:   []int{},
			expected: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := rangeOp(tc.values)
			assert.Exactly(t, tc.expected, actual)
		})
	}
}
