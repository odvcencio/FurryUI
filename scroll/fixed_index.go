package scroll

// FixedHeightIndex provides fast indexing for fixed-height items.
type FixedHeightIndex struct {
	Height int
	Count  func() int
}

// TotalHeight returns the total height for all items.
func (f FixedHeightIndex) TotalHeight() int {
	height := f.Height
	if height <= 0 {
		return 0
	}
	count := f.count()
	if count < 0 {
		count = 0
	}
	return height * count
}

// IndexForOffset returns the item index for a given offset.
func (f FixedHeightIndex) IndexForOffset(offset int) int {
	height := f.Height
	if height <= 0 {
		return 0
	}
	if offset <= 0 {
		return 0
	}
	index := offset / height
	maxIndex := f.count() - 1
	if maxIndex < 0 {
		maxIndex = 0
	}
	if index > maxIndex {
		index = maxIndex
	}
	return index
}

// OffsetForIndex returns the offset for the given item index.
func (f FixedHeightIndex) OffsetForIndex(index int) int {
	height := f.Height
	if height <= 0 || index <= 0 {
		return 0
	}
	count := f.count()
	if count <= 0 {
		return 0
	}
	if index >= count {
		index = count - 1
	}
	return index * height
}

func (f FixedHeightIndex) count() int {
	if f.Count == nil {
		return 0
	}
	return f.Count()
}
