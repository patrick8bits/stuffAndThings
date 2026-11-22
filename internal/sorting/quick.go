package sorting

func QuickSort(numbers []int) []int {
	return quickSort(numbers, nil)
}

func TraceQuickSort(numbers []int) Trace {
	return traceSort(numbers, quickSort)
}

func quickSort(numbers []int, record recordFunc) []int {
	quickSortRange(numbers, 0, len(numbers)-1, record)
	return numbers
}

func quickSortRange(
	numbers []int,
	low int,
	high int,
	record recordFunc,
) {
	if low >= high {
		return
	}

	pivot := numbers[high]
	i := low

	for j := low; j < high; j++ {
		if record != nil {
			record("compare", j, high)
		}

		if numbers[j] <= pivot {
			if i != j {
				numbers[i], numbers[j] =
					numbers[j], numbers[i]

				if record != nil {
					record("swap", i, j)
				}
			}

			i++
		}
	}

	if i != high {
		numbers[i], numbers[high] =
			numbers[high], numbers[i]

		if record != nil {
			record("swap", i, high)
		}
	}

	quickSortRange(numbers, low, i-1, record)
	quickSortRange(numbers, i+1, high, record)
}
