package collections

// обобщенные методы (функции) дженерики.
func RangeChunks[T any](chunkSize int, collection []T) [][]T {
	res := [][]T{}
	currentChunk := []T{}

	for _, v := range collection {
		currentChunk = append(currentChunk, v)
		if len(currentChunk) == chunkSize {
			res = append(res, currentChunk)
			currentChunk = []T{}
		}
	}
	if len(currentChunk) > 0 {
		res = append(res, currentChunk)
	}
	return res
}
