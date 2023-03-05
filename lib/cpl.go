package lib

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

/*
Represents a Board Component

May contain a Library Component, if linked
*/
type BoardComponent struct {
	Designator string
	Comment    string
	Package    string
	X          float64
	Y          float64
	Rotation   float64
	Layer      string
}

func (bc *BoardComponent) Prefix() string {
	return re1.ReplaceAllString(bc.Designator, "")
}

func (bc *BoardComponent) Value() string {
	return NormalizeValue(bc.Comment)
}

func (bc *BoardComponent) Key() []byte {
	return []byte(
		bc.Prefix() + ":" +
			strings.ReplaceAll(bc.Comment, ":", "_") + ":" +
			strings.ReplaceAll(bc.Package, ":", "_"),
	)
}

func (bc *BoardComponent) StringKey() string {
	return string(bc.Key())
}

func (bc *BoardComponent) Rotate(drotation float64) error {
	bc.Rotation += drotation
	for bc.Rotation < 0 {
		bc.Rotation += 360
	}
	for bc.Rotation > 360 {
		bc.Rotation -= 360
	}

	return nil
}

type AssocationMap struct {
	library     *Library
	assocations map[string]*LibraryComponent
}

func NewAssociationMap(library *Library) *AssocationMap {
	return &AssocationMap{library, make(map[string]*LibraryComponent)}
}

func (am *AssocationMap) FindAssociated(bcomponent *BoardComponent) *LibraryComponent {
	key := bcomponent.StringKey()
	if lcomponent, ok := am.assocations[key]; ok {
		return lcomponent
	}

	lcomponent := am.library.FindAssociated(bcomponent)
	am.assocations[key] = lcomponent

	return lcomponent
}

func (am *AssocationMap) Associate(bcomponent *BoardComponent, lcomponent *LibraryComponent) {
	key := bcomponent.StringKey()
	if lcomponent == nil {
		delete(am.assocations, key)
	}

	am.library.Associate(bcomponent, lcomponent)
	if lcomponent != nil {
		am.assocations[key] = lcomponent
	}
}

type BOMEntry struct {
	Comment     string
	Package     string
	Designators []string
	Component   *LibraryComponent
}

type BOM map[int64]*BOMEntry

func (bom BOM) AddComponent(component *BoardComponent, lc *LibraryComponent) {
	if _, ok := bom[lc.ID]; !ok {
		bom[lc.ID] = &BOMEntry{
			Comment:   component.Comment,
			Package:   component.Package,
			Component: lc,
		}
	}

	bom[lc.ID].Designators = append(
		bom[lc.ID].Designators, component.Designator,
	)
}

/*
Read a KiCAD POS file produced by kicad-cli export pcb pos

Return a list of Board Components
*/
func ReadPOS(src string) []*BoardComponent {
	fp, err := os.Open(src)
	if err != nil {
		return []*BoardComponent{}
	}
	defer fp.Close()

	components := []*BoardComponent{}
	reader := csv.NewReader(bufio.NewReader(fp))
	for line, err := reader.Read(); err != io.EOF; line, err = reader.Read() {
		if strings.TrimSpace(line[0]) == "Ref" {
			continue
		}

		rotation, err := strconv.ParseFloat(line[5], 32)
		if err != nil {
			rotation = 0
		}

		x, err := strconv.ParseFloat(line[3], 32)
		if err != nil {
			rotation = 0
		}

		y, err := strconv.ParseFloat(line[4], 32)
		if err != nil {
			rotation = 0
		}

		components = append(components, &BoardComponent{
			Designator: strings.TrimSpace(line[0]),
			Comment:    strings.TrimSpace(line[1]),
			Package:    strings.TrimSpace(line[2]),
			X:          x,
			Y:          y,
			Rotation:   rotation,
			Layer:      strings.TrimSpace(line[6]),
		})
	}

	return components
}

func WriteCPL(dst string, components []*BoardComponent) {
	fp, err := os.Create(dst)
	if err != nil {
		return
	}
	defer fp.Close()

	writer := csv.NewWriter(fp)
	writer.Write([]string{"Designator", "Mid X", "Mid Y", "Layer", "Rotation"})
	for _, component := range components {
		writer.Write([]string{
			component.Designator,
			fmt.Sprintf("%1.4f", component.X),
			fmt.Sprintf("%1.4f", component.Y),
			component.Layer,
			fmt.Sprintf("%1.0f", component.Rotation),
		})
	}

	writer.Flush()
}

func WriteBOM(dst string, bom BOM) {
	fp, err := os.Create(dst)
	if err != nil {
		return
	}
	defer fp.Close()

	writer := csv.NewWriter(fp)
	writer.Write([]string{"Comment", "Designator", "Footprint", "LCSC Part #"})
	for id, entry := range bom {
		writer.Write([]string{
			entry.Comment,
			strings.Join(entry.Designators, ","),
			entry.Package,
			LibraryComponent{ID: id}.CID(),
		})
	}

	writer.Flush()
}
