package reportdata

import (
	"strings"

	"github.com/Kryvea/Kryvea/internal/model"
)

func splitText(s string, coordinates []model.HighlightedText) []model.Highlighted {
	if len(coordinates) == 0 {
		return []model.Highlighted{
			{
				Text:  escapeXMLString(s),
				Color: "",
			},
		}
	}

	rows := strings.SplitAfter(s, "\n")
	colors := make([][]string, len(rows))
	for i := range colors {
		colors[i] = make([]string, len(rows[i]))
	}

	// Split multi-line highlights into one single-line segment per row, guarding
	// every line/column bound so out-of-range coordinates never index colors out of bounds.
	normalized := make([]model.HighlightedText, 0, len(coordinates))
	for _, coordinate := range coordinates {
		if coordinate.Start.Line < 1 || coordinate.Start.Line > len(rows) {
			continue
		}
		if coordinate.End.Line > len(rows) {
			coordinate.End.Line = len(rows)
			coordinate.End.Col = len(rows[coordinate.End.Line-1])
		}

		for coordinate.Start.Line != coordinate.End.Line {
			if coordinate.Start.Line > len(rows) {
				break
			}
			segment := coordinate
			segment.End.Line = segment.Start.Line
			if segment.Start.Col < 1 {
				segment.Start.Col = 1
			}
			segment.End.Col = len(rows[segment.Start.Line-1]) + 1
			normalized = append(normalized, segment)

			coordinate.Start.Line++
			coordinate.Start.Col = 1
		}
		if coordinate.Start.Line > len(rows) {
			continue
		}

		// upper bound first: on an empty row len(row) is 0 and the lower
		// bound must win, or the paint loop would index colors[row][-1]
		row := rows[coordinate.Start.Line-1]
		if coordinate.Start.Col > len(row) {
			coordinate.Start.Col = len(row)
		}
		if coordinate.Start.Col < 1 {
			coordinate.Start.Col = 1
		}
		if coordinate.End.Col > len(row) {
			coordinate.End.Col = len(row)
			if !strings.HasSuffix(row, "\n") {
				coordinate.End.Col++
			}
		}
		if coordinate.End.Col < 1 {
			coordinate.End.Col = 1
		}

		normalized = append(normalized, coordinate)
	}

	for _, coordinate := range normalized {
		for i := coordinate.Start.Col; i < coordinate.End.Col; i++ {
			colors[coordinate.Start.Line-1][i-1] = coordinate.Color
		}
	}

	splitted := []model.Highlighted{}
	splitColor := model.Highlighted{
		Text:  "",
		Color: "",
	}

	builder := strings.Builder{}
	for i, colorRow := range colors {
		for j, color := range colorRow {
			if color != splitColor.Color {
				splitColor.Text = escapeXMLString(builder.String())
				if splitColor.Text != "" {
					splitted = append(splitted, splitColor)
				}
				splitColor = model.Highlighted{}
				builder = strings.Builder{}
			}
			builder.WriteByte(rows[i][j])
			splitColor.Color = color
		}
	}
	if builder.Len() > 0 {
		splitColor.Text = escapeXMLString(builder.String())
		splitted = append(splitted, splitColor)
	}

	return splitted
}
