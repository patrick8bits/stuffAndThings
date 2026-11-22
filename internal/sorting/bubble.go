package sorting

func BubbleSort(numbers []int) []int {
	return bubbleSort(numbers, nil)
}

func TraceBubbleSort(numbers []int) Trace {
	return traceSort(numbers, bubbleSort)
}

func bubbleSort(numbers []int, record recordFunc) []int {
	n := len(numbers)

	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {

			if record != nil {
				record("compare", j, j+1)
			}

			if numbers[j] > numbers[j+1] {
				numbers[j], numbers[j+1] =
					numbers[j+1], numbers[j]

				if record != nil {
					record("swap", j, j+1)
				}
			}
		}
	}

	return numbers
}
