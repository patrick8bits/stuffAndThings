package sorting

import (
	"slices"
	"testing"
)

func TestTraceBubbleSort(t *testing.T) {
	numbers := []int{3, 1, 2}

	trace := TraceBubbleSort(numbers)
	t.Logf("Execution time: %d ns", trace.DurationNS)

	if trace.Comparisons != 3 {
		t.Errorf("Expected 3 comparisons, got %d",
			trace.Comparisons)
	}

	if trace.Swaps != 2 {
		t.Errorf("Expected 2 swaps, got %d",
			trace.Swaps)
	}

	if len(trace.Steps) != 7 {
		t.Fatalf("Expected 7 steps, got %d",
			len(trace.Steps))
	}

	first := trace.Steps[0].Values
	last := trace.Steps[len(trace.Steps)-1].Values

	if !slices.Equal(first, []int{3, 1, 2}) {
		t.Errorf("Incorrect initial frame: %v", first)
	}

	if !slices.Equal(last, []int{1, 2, 3}) {
		t.Errorf("Incorrect final frame: %v", last)
	}

	if !slices.Equal(numbers, []int{3, 1, 2}) {
		t.Error("Original input was modified")
	}
}

func TestRemainingTraces(t *testing.T) {
	tests := []struct {
		name        string
		sort        func([]int) Trace
		comparisons int
		swaps       int
		shifts      int
	}{
		{"InsertionSort", TraceInsertionSort, 3, 0, 2},
		{"QuickSort", TraceQuickSort, 2, 2, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			numbers := []int{3, 1, 2}
			trace := tt.sort(numbers)

			if trace.Comparisons != tt.comparisons {
				t.Errorf("Expected %d comparisons, got %d",
					tt.comparisons, trace.Comparisons)
			}

			if trace.Swaps != tt.swaps {
				t.Errorf("Expected %d swaps, got %d",
					tt.swaps, trace.Swaps)
			}

			if trace.Shifts != tt.shifts {
				t.Errorf("Expected %d shifts, got %d",
					tt.shifts, trace.Shifts)
			}

			last := trace.Steps[len(trace.Steps)-1].Values

			if !slices.Equal(last, []int{1, 2, 3}) {
				t.Errorf("Incorrect final frame: %v", last)
			}

			if !slices.Equal(numbers, []int{3, 1, 2}) {
				t.Error("Original input was modified")
			}
		})
	}
}
