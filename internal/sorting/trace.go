package sorting

import (
	"slices"
	"time"
)

type Step struct {
	Action  string `json:"action"`
	Values  []int  `json:"values"`
	Indices []int  `json:"indices"`
}

type Trace struct {
	Steps       []Step `json:"steps"`
	Comparisons int    `json:"comparisons"`
	Swaps       int    `json:"swaps"`
	Shifts      int    `json:"shifts"`
	DurationNS  int64  `json:"duration_ns"`
}

type recordFunc func(action string, indices ...int)

func traceSort(
	numbers []int,
	algorithm func([]int, recordFunc) []int,
) Trace {
	working := slices.Clone(numbers)
	trace := Trace{}
	// Measure execution time without recording overhead.
	timedInput := slices.Clone(numbers)

	started := time.Now()

	algorithm(timedInput, nil)

	trace.DurationNS = time.Since(started).Nanoseconds()

	record := func(action string, indices ...int) {
		if action == "compare" {
			trace.Comparisons++
		}

		if action == "swap" {
			trace.Swaps++
		}

		if action == "shift" {
			trace.Shifts++
		}

		step := Step{
			Action:  action,
			Values:  slices.Clone(working),
			Indices: append([]int{}, indices...),
		}

		trace.Steps = append(trace.Steps, step)
	}

	record("start")

	algorithm(working, record)

	record("done")

	return trace
}
