package policy

import (
	"fmt"
	"slices"
	"strings"

	"github.com/lightningnetwork/lnd/lnrpc"
	"golang.org/x/exp/constraints"
)

// Operations to measure the central tendency of a data set.
const (
	// Middle value in a list ordered from smallest to largest.
	Median Operation = "median"
	// Average of a list of numbers.
	Mean Operation = "mean"
	// Most frequently occurring value on a list.
	Mode Operation = "mode"
	// Difference between the biggest and the smallest number.
	RangeOp Operation = "range"
)

// Operation is a mathematical operation applied to a set of values.
type Operation string

// Number is an integer or float.
type Number interface {
	constraints.Integer | constraints.Float
}

// Range represents the limits of a series.
type Range[T Number] struct {
	Min *T `yaml:"min,omitempty"`
	Max *T `yaml:"max,omitempty"`
}

// Contains returns whether the received value is within the range.
func (r Range[T]) Contains(v T) bool {
	if r.Min != nil && v < *r.Min {
		return false
	}
	if r.Max != nil && v > *r.Max {
		return false
	}
	return true
}

// Reason returns the reason why a number was not in the range.
func (r Range[T]) Reason() string {
	if r.Min != nil && r.Max != nil {
		return fmt.Sprintf("is not between %v and %v", *r.Min, *r.Max)
	}
	if r.Min != nil {
		return fmt.Sprintf("is lower than %v", *r.Min)
	}
	if r.Max != nil {
		return fmt.Sprintf("is higher than %v", *r.Max)
	}

	return ""
}

func check[T Number](r *Range[T], v T) bool {
	if r == nil {
		return true
	}

	return r.Contains(v)
}

// StatRange is like a range but received multiple values and applies an operation to them.
type StatRange[T Number] struct {
	Min       *T        `yaml:"min,omitempty"`
	Max       *T        `yaml:"max,omitempty"`
	Operation Operation `yaml:"operation,omitempty"`
}

// Contains returns whether the aggregated value is within the range.
func (a StatRange[T]) Contains(values []T) bool {
	var v T
	switch a.Operation {
	case Median:
		v = median(values)
	case Mode:
		v = mode(values)
	case RangeOp:
		v = rangeOp(values)
	default:
		v = mean(values)
	}

	// Range is not used as a property to have a cleaner configuration and avoid declaring min
	// and max inside "range"
	r := &Range[T]{
		Min: a.Min,
		Max: a.Max,
	}
	return r.Contains(v)
}

// Reason returns the reason why a number was not in the range.
func (a StatRange[T]) Reason() string {
	r := &Range[T]{
		Min: a.Min,
		Max: a.Max,
	}

	var sb strings.Builder
	if a.Operation == "" {
		a.Operation = Mean
	}
	sb.WriteString(string(a.Operation))
	sb.WriteString(" value ")
	sb.WriteString(r.Reason())
	return sb.String()
}

type channelFunc[T Number] func(peer *lnrpc.NodeInfo, channel *lnrpc.ChannelEdge) T

func checkStat[T Number](
	sr *StatRange[T],
	peer *lnrpc.NodeInfo,
	f channelFunc[T],
) bool {
	if sr == nil {
		return true
	}

	values := make([]T, 0, len(peer.Channels))
	for _, channel := range peer.Channels {
		value := f(peer, channel)
		values = append(values, value)
	}

	return sr.Contains(values)
}

func median[T Number](values []T) T {
	if len(values) == 0 {
		return 0
	}
	slices.Sort(values)

	l := len(values)
	if l%2 == 0 {
		return (values[l/2-1] + values[l/2]) / 2.0
	}

	return values[l/2]
}

func mean[T Number](values []T) T {
	if len(values) == 0 {
		return 0
	}

	var sum T
	for _, v := range values {
		sum += v
	}

	return sum / T(len(values))
}

func mode[T Number](values []T) T {
	if len(values) == 0 {
		return 0
	}

	occurences := make(map[T]T)
	for _, v := range values {
		occurences[v]++
	}

	var highest T
	for value, count := range occurences {
		if count > occurences[highest] {
			highest = value
		}
	}

	return highest
}

func rangeOp[T Number](values []T) T {
	if len(values) < 2 {
		return 0
	}

	slices.Sort(values)

	return values[len(values)-1] - values[0]
}
