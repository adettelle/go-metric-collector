// Package collections provides utility functions for working with collections of data.
// It includes generic methods such as chunking a collection into smaller pieces.
package collections

// RangeChunks is a generic function. It divides a collection into chunks of the specified size.
// The result is a 2D slice where each inner slice has a length of up to chunkSize elements.
// Example:
//
//	collection := []int{1, 2, 3, 4, 5, 6, 7}
//	chunks := RangeChunks(3, collection) // chunks will be [[1, 2, 3], [4, 5, 6], [7]]
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
	// If there are remaining elements in currentChunk, add them as the last chunk
	if len(currentChunk) > 0 {
		res = append(res, currentChunk)
	}
	return res
}
