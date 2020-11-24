package lib

type SVGFootprint struct {
	Title      string
	Rectangles []*SVGRectangle
	Grids      []*SVGGrid
}

type SVGRectangle struct{}

type SVGGrid struct{}

type SVGPolyline struct{}

type SVGCircle struct{}

type SVGText struct{}

type SVGPath struct{}
