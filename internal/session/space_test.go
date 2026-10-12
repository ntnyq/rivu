package session

import "testing"

func TestRequiredSpaceIncludesOnlySelectedPiecePayload(t *testing.T) {
	for _, tt := range []struct {
		name        string
		pieceLength int64
		files       []File
		want        int64
	}{
		{"shared piece", 16, []File{{Size: 5, Priority: 1}, {Offset: 5, Size: 50}}, 16},
		{"leading overlap", 16, []File{{Size: 50}, {Offset: 50, Size: 5, Priority: 1}}, 7},
		{"deduplicate shared piece", 16, []File{{Size: 5, Priority: 1}, {Offset: 5, Size: 5, Priority: 1}, {Offset: 10, Size: 50}}, 16},
		{"v2 padding", 16, []File{{Size: 5, Priority: 1}, {Offset: 16, Size: 50}}, 5},
		{"nonadjacent pieces", 16, []File{{Size: 1, Priority: 1}, {Offset: 1, Size: 63}, {Offset: 64, Size: 1, Priority: 2}}, 17},
		{"empty selection", 16, []File{{Size: 50}}, 0},
		{"zero byte file", 16, []File{{Size: 0, Priority: 1}}, 0},
		{"legacy geometry", 0, []File{{Size: 5, Priority: 1}, {Size: 50}}, 55},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := Source{Files: tt.files, PieceLength: tt.pieceLength}
			if got := s.requiredSpace(); got != tt.want {
				t.Fatalf("required %d bytes, want %d", got, tt.want)
			}
		})
	}
}
