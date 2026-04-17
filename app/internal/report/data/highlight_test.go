package reportdata

import (
	"testing"

	"github.com/Kryvea/Kryvea/internal/mongo"
)

func TestHighlight(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		highlights []mongo.HighlightedText
		expected   []mongo.Highlighted
	}{
		{
			name: "single highlight",
			text: "This is a sample text for testing highlights.",
			highlights: []mongo.HighlightedText{
				{
					Start: mongo.LineCol{Line: 1, Col: 11},
					End:   mongo.LineCol{Line: 1, Col: 17},
					Color: "FF0000",
				},
			},
			expected: []mongo.Highlighted{
				{Text: "This is a "},
				{Text: "sample", Color: "FF0000"},
				{Text: " text for testing highlights."},
			},
		},
		{
			name: "multiple highlights",
			text: "Highlighting multiple sections in this text.",
			highlights: []mongo.HighlightedText{
				{
					Start: mongo.LineCol{Line: 1, Col: 1},
					End:   mongo.LineCol{Line: 1, Col: 13},
					Color: "00FF00",
				},
				{
					Start: mongo.LineCol{Line: 1, Col: 22},
					End:   mongo.LineCol{Line: 1, Col: 31},
					Color: "0000FF",
				},
			},
			expected: []mongo.Highlighted{
				{Text: "Highlighting", Color: "00FF00"},
				{Text: " multiple"},
				{Text: " sections", Color: "0000FF"},
				{Text: " in this text."},
			},
		},
		{
			name: "overlapping highlights",
			text: "Overlapping highlights can be tricky.",
			highlights: []mongo.HighlightedText{
				{
					Start: mongo.LineCol{Line: 1, Col: 1},
					End:   mongo.LineCol{Line: 1, Col: 12},
					Color: "FF00FF",
				},
				{
					Start: mongo.LineCol{Line: 1, Col: 5},
					End:   mongo.LineCol{Line: 1, Col: 23},
					Color: "00FFFF",
				},
			},
			expected: []mongo.Highlighted{
				{Text: "Over", Color: "FF00FF"},
				{Text: "lapping highlights", Color: "00FFFF"},
				{Text: " can be tricky."},
			},
		},
		{
			name: "highlight at text boundaries",
			text: "Boundary highlights.",
			highlights: []mongo.HighlightedText{
				{
					Start: mongo.LineCol{Line: 1, Col: 1},
					End:   mongo.LineCol{Line: 1, Col: 9},
					Color: "123456",
				},
				{
					Start: mongo.LineCol{Line: 1, Col: 10},
					End:   mongo.LineCol{Line: 1, Col: 21},
					Color: "654321",
				},
			},
			expected: []mongo.Highlighted{
				{Text: "Boundary", Color: "123456"},
				{Text: " "},
				{Text: "highlights.", Color: "654321"},
			},
		},
		{
			name: "highlight at text boundaries multiline",
			text: "Boundary highlights.\nThis is a new line\nThird line.\nVery long fourth line here.",
			highlights: []mongo.HighlightedText{
				{
					Start: mongo.LineCol{Line: 1, Col: 10},
					End:   mongo.LineCol{Line: 4, Col: 5},
					Color: "123456",
				},
			},
			expected: []mongo.Highlighted{
				{Text: "Boundary "},
				{Text: "highlights.\nThis is a new line\nThird line.\nVery", Color: "123456"},
				{Text: " long fourth line here."},
			},
		},
		{
			name:       "no highlights",
			text:       "No highlights in this text.",
			highlights: []mongo.HighlightedText{},
			expected: []mongo.Highlighted{
				{Text: "No highlights in this text."},
			},
		},
		{
			name:       "nil highlights",
			text:       "nil highlights in this text.",
			highlights: nil,
			expected: []mongo.Highlighted{
				{Text: "nil highlights in this text."},
			},
		},
		{
			name: "multiline single highlight",
			text: "This is line one.\nThis is line two.\nThis is line three.",
			highlights: []mongo.HighlightedText{
				{
					Start: mongo.LineCol{Line: 2, Col: 6},
					End:   mongo.LineCol{Line: 2, Col: 11},
					Color: "FF5733",
				},
			},
			expected: []mongo.Highlighted{
				{Text: "This is line one.\nThis "},
				{Text: "is li", Color: "FF5733"},
				{Text: "ne two.\nThis is line three."},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			highlighted := splitText(tt.text, tt.highlights)

			if len(highlighted) != len(tt.expected) {
				t.Errorf("Expected %d segments, got %d", len(tt.expected), len(highlighted))
			}

			for i, exp := range tt.expected {
				if i >= len(highlighted) {
					break
				}
				got := highlighted[i]
				if got.Text != exp.Text {
					t.Errorf("Segment %d: expected text '%s', got '%s'", i, exp.Text, got.Text)
				}
				if got.Color != exp.Color {
					t.Errorf("Segment %d: expected color '%s', got '%s'", i, exp.Color, got.Color)
				}
			}
		})
	}
}
