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
		for j := range len(rows[i]) {
			colors[i][j] = ""
		}
	}

	modified := true
	for modified {
		for i := 0; i < len(coordinates); i++ {
			modified = false
			if coordinates[i].Start.Line > len(rows) {
				copy(coordinates[i:], coordinates[i+1:])
				continue
			}
			if coordinates[i].End.Line > len(rows) {
				coordinates[i].End.Line = len(rows)
				coordinates[i].End.Col = len(rows[coordinates[i].End.Line-1])
			}
			if coordinates[i].Start.Line != coordinates[i].End.Line {
				coordinates = append(coordinates, model.HighlightedText{})
				first, second := coordinates[i], coordinates[i]

				first.End.Line = first.Start.Line
				first.End.Col = len(rows[first.End.Line-1]) + 1

				second.Start.Line++
				second.Start.Col = 1

				copy(coordinates[i+2:], coordinates[i+1:])
				coordinates[i] = first
				coordinates[i+1] = second
				modified = true

				continue
			}
			if coordinates[i].Start.Col > len(rows[coordinates[i].Start.Line-1]) {
				coordinates[i].Start.Col = len(rows[coordinates[i].Start.Line-1])
			}
			if coordinates[i].Start.Col < 0 {
				coordinates[i].Start.Col = 1
			}
			if coordinates[i].End.Col > len(rows[coordinates[i].End.Line-1]) {
				coordinates[i].End.Col = len(rows[coordinates[i].End.Line-1])
				if !strings.HasSuffix(rows[coordinates[i].End.Line-1], "\n") {
					coordinates[i].End.Col++
				}
			}
			if coordinates[i].End.Col < 0 {
				coordinates[i].End.Col = 1
			}
		}
	}

	for _, coordinate := range coordinates {
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
