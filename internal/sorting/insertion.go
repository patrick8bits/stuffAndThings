package sorting

func InsertionSort(numbers []int) []int {
	return insertionSort(numbers, nil)
}

func TraceInsertionSort(numbers []int) Trace {
	return traceSort(numbers, insertionSort)
}

func insertionSort(numbers []int, record recordFunc) []int {
	for i := 1; i < len(numbers); i++ {
		key := numbers[i]
		j := i - 1

		for j >= 0 {
			if record != nil {
				record("compare", j)
			}

			if numbers[j] <= key {
				break
			}

			numbers[j+1] = numbers[j]

			if record != nil {
				record("shift", j, j+1)
			}

			j--
		}

		numbers[j+1] = key

		if record != nil {
			record("insert", j+1)
		}
	}

	return numbers
}
