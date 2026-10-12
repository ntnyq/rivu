package session

import "sort"

// requiredSpace counts payload bytes touched by selected pieces. Unselected
// neighbors can share those pieces; v2 alignment gaps do not occupy files.
func (s Source) requiredSpace() int64 {
	if s.PieceLength <= 0 {
		// Older saved sources lack geometry. Keep their estimate conservative.
		var total int64
		for _, f := range s.Files {
			total += f.Size
		}
		return total
	}
	type span struct{ start, end int64 }
	var selected []span
	for _, f := range s.Files {
		if f.Priority > 0 && f.Size > 0 {
			selected = append(selected, span{
				start: f.Offset / s.PieceLength * s.PieceLength,
				end:   (f.Offset + f.Size + s.PieceLength - 1) / s.PieceLength * s.PieceLength,
			})
		}
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].start < selected[j].start })
	merged := selected[:0]
	for _, interval := range selected {
		if len(merged) > 0 && interval.start <= merged[len(merged)-1].end {
			last := &merged[len(merged)-1]
			last.end = max(last.end, interval.end)
		} else {
			merged = append(merged, interval)
		}
	}
	var total int64
	for _, f := range s.Files {
		first := sort.Search(len(merged), func(i int) bool { return merged[i].end > f.Offset })
		for _, interval := range merged[first:] {
			if interval.start >= f.Offset+f.Size {
				break
			}
			total += max(0, min(interval.end, f.Offset+f.Size)-max(interval.start, f.Offset))
		}
	}
	return total
}
