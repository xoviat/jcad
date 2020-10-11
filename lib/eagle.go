package lib

import "encoding/xml"

/*
	Package generates an eagle library from the database
*/

type EagleLibrary struct {
	XMLName     *xml.Name                `xml:"eagle"`
	Settings    []*EagleLibrarySetting   `xml:"drawing>settings>setting"`
	Grid        *EagleLibraryGrid        `xml:"drawing>grid"`
	Layers      []*EagleLibraryLayer     `xml:"drawing>layers>layer"`
	Description string                   `xml:"drawing>library>description"`
	Packages    []*EagleLibraryPackage   `xml:"drawing>library>packages>package"`
	Symbols     []*EagleLibrarySymbol    `xml:"drawing>library>symbols>symbol"`
	DevicesSets []*EagleLibraryDeviceSet `xml:"drawing>library>devicesets>deviceset"`
}

type EagleLibrarySetting struct {
	Attr xml.Attr `xml:",any,attr"`
}

type EagleLibraryGrid struct {
	Distance    string `xml:"distance,attr"`
	Unitdist    string `xml:"unitdist,attr"`
	Unit        string `xml:"unit,attr"`
	Style       string `xml:"style,attr"`
	Multiple    string `xml:"multiple,attr"`
	Display     string `xml:"display,attr"`
	Altdistance string `xml:"altdistance,attr"`
	Altunitdist string `xml:"altunitdist,attr"`
	Altunit     string `xml:"altunit,attr"`
}

type EagleLibraryLayer struct {
	Number  string `xml:"number,attr"`
	Name    string `xml:"name,attr"`
	Color   string `xml:"color,attr"`
	Fill    string `xml:"fill,attr"`
	Visible string `xml:"visible,attr"`
	Active  string `xml:"active,attr"`
}

type EagleLibraryPackage struct {
	Name        string                    `xml:"name,attr"`
	Description string                    `xml:"description"`
	SMDs        []*EagleLibrarySMD        `xml:"smd"`
	Texts       []*EagleLibraryText       `xml:"text"`
	Wires       []*EagleLibraryWire       `xml:"wire"`
	Rectangles  []*EagleLibraryRectangle  `xml:"rectangle"`
	Pads        []*EagleLibraryPackagePad `xml:"pad"`
	Polygons    []*EagleLibraryPolygon    `xml:"polygon"`
	Circles     []*EagleLibraryCircle     `xml:"circle"`
}

type EagleLibrarySMD struct {
	Name  string `xml:"name,attr"`
	X     string `xml:"x,attr"`
	Y     string `xml:"y,attr"`
	Dx    string `xml:"dx,attr"`
	Dy    string `xml:"dy,attr"`
	Layer string `xml:"layer,attr"`
}

type EagleLibraryText struct {
	X     string `xml:"x,attr"`
	Y     string `xml:"y,attr"`
	Size  string `xml:"size,attr"`
	Layer string `xml:"layer,attr"`
	Font  string `xml:"font,attr"`
	Ratio string `xml:"ratio,attr"`
	Align string `xml:"align,attr"`
	Text  string `xml:",chardata"`
}

type EagleLibraryWire struct {
	X1    string `xml:"x1,attr"`
	Y1    string `xml:"y1,attr"`
	X2    string `xml:"x2,attr"`
	Y2    string `xml:"y2,attr"`
	Width string `xml:"width,attr"`
	Layer string `xml:"layer,attr"`
}

type EagleLibraryRectangle struct {
	X1    string `xml:"x1,attr"`
	Y1    string `xml:"y1,attr"`
	X2    string `xml:"x2,attr"`
	Y2    string `xml:"y2,attr"`
	Layer string `xml:"layer,attr"`
}

type EagleLibraryPackagePad struct {
	Name     string `xml:"name,attr"`
	X        string `xml:"x,attr"`
	Y        string `xml:"y,attr"`
	Drill    string `xml:"drill,attr"`
	Diameter string `xml:"diameter,attr"`
	Stop     string `xml:"stop,attr"`
}

type EagleLibraryPolygon struct {
	Width    string                `xml:"width,attr"`
	Layer    string                `xml:"layer,attr"`
	Vertices []*EagleLibraryVertex `xml:"vertex"`
}

type EagleLibraryVertex struct {
	X     string `xml:"x,attr"`
	Y     string `xml:"y,attr"`
	Curve string `xml:"curve,attr"`
}

type EagleLibraryCircle struct {
	X      string `xml:"x,attr"`
	Y      string `xml:"y,attr"`
	Radius string `xml:"radius,attr"`
	Width  string `xml:"width,attr"`
	Layer  string `xml:"layer,attr"`
}

type EagleLibraryPin struct {
	Name      string `xml:"name,attr"`
	X         string `xml:"x,attr"`
	Y         string `xml:"y,attr"`
	Visible   string `xml:"visible,attr"`
	Length    string `xml:"length,attr"`
	Direction string `xml:"direction,attr"`
	Swaplevel string `xml:"swaplevel,attr"`
	Rot       string `xml:"rot,attr"`
}

type EagleLibrarySymbol struct {
	Name        string                   `xml:"name,attr"`
	Description string                   `xml:"description"`
	Wires       []*EagleLibraryWire      `xml:"wire"`
	Polygons    []*EagleLibraryPolygon   `xml:"polygon"`
	Rectangles  []*EagleLibraryRectangle `xml:"rectangle"`
	Texts       []*EagleLibraryText      `xml:"text"`
	Pins        []*EagleLibraryPin       `xml:"pin"`
}

type EagleLibraryDeviceSet struct {
	Name        string                `xml:"name,attr"`
	Prefix      string                `xml:"prefix,attr"`
	Description string                `xml:"description"`
	Devices     []*EagleLibraryDevice `xml:"devices>device"`
	Gates       []*EagleLibraryGate   `xml:"gates>gate"`
}

type EagleLibraryGate struct {
	Name   string `xml:"name,attr"`
	Symbol string `xml:"symbol,attr"`
	X      string `xml:"x,attr"`
	Y      string `xml:"y,attr"`
}

type EagleLibraryDevice struct {
	Name         string                    `xml:"name,attr"`
	Package      string                    `xml:"package,attr"`
	Connects     []*EagleLibraryConnect    `xml:"connects>connect"`
	Technologies []*EagleLibraryTechnology `xml:"technologies>technology"`
}

type EagleLibraryConnect struct {
	Gate string `xml:"gate,attr"`
	Pin  string `xml:"pin,attr"`
	Pad  string `xml:"pad,attr"`
}

type EagleLibraryTechnology struct {
	Name       string                   `xml:"name,attr"`
	Attributes []*EagleLibraryAttribute `xml:"attribute"`
}

type EagleLibraryAttribute struct {
	Name     string `xml:"name,attr"`
	Value    string `xml:"value,attr"`
	Constant string `xml:"contstant,attr"`
}
