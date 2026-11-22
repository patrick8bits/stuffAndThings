package sorting

import (
	"slices"
	"testing"
)

func TestSortingAlgorithms(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "Random numbers",
			input:    []int{42, 17, 95, 8, 63},
			expected: []int{8, 17, 42, 63, 95},
		},
		{
			name:     "Already sorted",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Reverse order",
			input:    []int{5, 4, 3, 2, 1},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Duplicate values",
			input:    []int{3, 1, 3, 2, 1},
			expected: []int{1, 1, 2, 3, 3},
		},
		{
			name:     "Negative numbers",
			input:    []int{-5, 3, -2, 0},
			expected: []int{-5, -2, 0, 3},
		},
		{
			name:     "Single element",
			input:    []int{42},
			expected: []int{42},
		},
		{
			name:     "Empty slice",
			input:    []int{},
			expected: []int{},
		},
	}

	algorithms := []struct {
		name string
		sort func([]int) []int
	}{
		{"BubbleSort", BubbleSort},
		{"InsertionSort", InsertionSort},
		{"QuickSort", QuickSort},
	}

	for _, algorithm := range algorithms {
		t.Run(algorithm.name, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					input := slices.Clone(tt.input)

					result := algorithm.sort(input)

					if !slices.Equal(result, tt.expected) {
						t.Errorf("Expected %v, got %v",
							tt.expected, result)
					}
				})
			}
		})
	}
}
