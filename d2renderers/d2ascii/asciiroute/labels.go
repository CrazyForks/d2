package asciiroute

import (
	"math"
	"strings"

	"oss.terrastruct.com/d2/lib/geo"
)

// RouteLabelPosition holds calculated position for route label
type RouteLabelPosition struct {
	I       int     // Index of route segment
	X       int     // X coordinate for label
	Y       int     // Y coordinate offset
	MaxDiff float64 // Maximum difference for the segment
}

func (pos *RouteLabelPosition) ShouldDrawAt(currentIndex, x, y int, ax, ay, sx, sy float64) bool {
	if pos.I != currentIndex {
		return false
	}

	if sy != 0 {
		return int(math.Round(ay))+int(math.Round(pos.MaxDiff/2))*geo.Sign(sy) == y
	}

	if sx != 0 {
		return int(math.Round(ax))+int(math.Round(pos.MaxDiff/2))*geo.Sign(sx) == x
	}

	return false
}

func calculateBestLabelPosition(rd RouteDrawer, routes []*geo.Point, label string) *RouteLabelPosition {
	if len(routes) < 2 {
		return nil
	}

	fw := rd.GetFontWidth()
	fh := rd.GetFontHeight()

	maxDiff := 0.0
	bestIndex := -1
	bestX := 0.0
	scaleOld := 0.0

	for i := 0; i < len(routes)-1; i++ {
		diffY := math.Abs(routes[i].Y - routes[i+1].Y)
		diffX := math.Abs(routes[i].X - routes[i+1].X)
		diff := math.Max(diffY, diffX)
		scale := (math.Abs(float64(geo.Sign(diffX)))*fw + math.Abs(float64(geo.Sign(diffY)))*fh)

		if diff*scale > maxDiff*scaleOld {
			maxDiff = diff
			bestIndex = i
			bestX = routes[i].X

			// Center label on horizontal segments
			if diff == diffX && i+1 < len(routes) {
				direction := geo.Sign(routes[i+1].X - routes[i].X)
				bestX = routes[i].X + (float64(direction) * diff / 2)
			}
		}
		scaleOld = scale
	}

	if bestIndex == -1 {
		return nil
	}

	return &RouteLabelPosition{
		I:       bestIndex,
		X:       int(math.Round(bestX)) - len(label)/2,
		Y:       int(math.Round(maxDiff / 2)),
		MaxDiff: maxDiff,
	}
}

func drawConnectionLabel(rd RouteDrawer, labelPos *RouteLabelPosition, label, labelPosition string, x, y int, sx, sy float64, routes []*geo.Point, i int) {
	canvas := rd.GetCanvas()

	if sy != 0 {
		// Vertical segment - clear current position and draw label horizontally
		if isInBounds(rd, x, y) {
			canvas.Set(x, y, " ")
		}
		for j, ch := range label {
			if isInBounds(rd, labelPos.X+j, y) {
				canvas.Set(labelPos.X+j, y, string(ch))
			}
		}
	} else if sx != 0 {
		// Horizontal segment - draw label above or below
		yFactor := 0
		if strings.Contains(labelPosition, "TOP") {
			yFactor = -1
		} else if strings.Contains(labelPosition, "BOTTOM") {
			yFactor = 1
		}

		// Adjust X position based on LEFT/RIGHT preference
		xPos := labelPos.X
		if strings.Contains(labelPosition, "LEFT") {
			xPos = int(routes[labelPos.I+absInt((geo.Sign(sx)-1)/2)].X)
		} else if strings.Contains(labelPosition, "RIGHT") {
			xPos = int(routes[labelPos.I+((geo.Sign(sx)+1)/2)].X) - len(label)/2
		}

		for j, ch := range label {
			if isInBounds(rd, xPos+j, y+yFactor) {
				canvas.Set(xPos+j, y+yFactor, string(ch))
			}
		}
	}
}

func drawDestinationLabel(rd RouteDrawer, label string, cx, cy, sx, sy float64) {
	canvas := rd.GetCanvas()
	ly := 0
	lx := 0
	if math.Abs(sx) > 0 {
		ly = int(cy - 1)
		if sx > 0 {
			lx = int(cx) - 1 - len(label)
		} else {
			lx = int(cx)
		}
	} else if math.Abs(sy) > 0 {
		ly = int(cy - 1)
		lx = int(cx + 1)
	}
	for j, ch := range label {
		canvas.Set(lx+j+LabelOffsetX, ly, string(ch))
	}
}

func drawSourceLabel(rd RouteDrawer, label string, ax, cy, cx, sx, sy float64) {
	canvas := rd.GetCanvas()
	ly := 0
	lx := 0
	if math.Abs(sx) > 0 {
		ly = int(cy - 1)
		if sx > 0 {
			lx = int(ax)
		} else {
			lx = int(ax) - 1 - len(label)
		}
	} else if math.Abs(sy) > 0 {
		ly = int(cy - 1)
		lx = int(cx + 1)
	}
	for j, ch := range label {
		canvas.Set(lx+j, ly, string(ch))
	}
}
