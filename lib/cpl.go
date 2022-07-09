package lib

import (
	"encoding/csv"
	"fmt"
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
	X          string
	Y          string
	Rotation   float64
	Layer      string
}

func (bc *BoardComponent) Key() []byte {
	return []byte(
		re1.ReplaceAllString(bc.Designator, "") + ":" +
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
	if lcomponent == nil {
		delete(am.assocations, bcomponent.StringKey())
	}

	am.library.Associate(bcomponent, lcomponent)
}

type BOMEntry struct {
	Comment     string
	Designators []string
	Component   *LibraryComponent
}

type BOM map[string]*BOMEntry

func (bom BOM) AddComponent(bcomponent *BoardComponent, lcomponent *LibraryComponent) {
	if _, ok := bom[lcomponent.CID()]; !ok {
		bom[lcomponent.CID()] = &BOMEntry{
			Comment:   bcomponent.Comment,
			Component: lcomponent,
		}
	}

	bom[lcomponent.CID()].Designators = append(
		bom[lcomponent.CID()].Designators,
		bcomponent.Designator,
	)
}

/*
	Read a KiCAD CPL file produced by generate_cpl.py

	Return a list of Board Components
*/
func ReadKCPL(src string) []*BoardComponent {
	fp, err := os.Open(src)
	if err != nil {
		return []*BoardComponent{}
	}
	defer fp.Close()

	components := []*BoardComponent{}
	reader := csv.NewReader(fp)
	for line, _ := reader.Read(); len(line) > 0; line, _ = reader.Read() {
		rotation, err := strconv.ParseFloat(line[5], 32)
		if err != nil {
			rotation = 0
		}

		components = append(components, &BoardComponent{
			Designator: line[0],
			Comment:    line[1],
			Package:    line[2],
			X:          line[3],
			Y:          line[4],
			Rotation:   rotation,
			Layer:      line[6],
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
			component.X,
			component.Y,
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
	for cid, entry := range bom {
		writer.Write([]string{
			entry.Comment,
			strings.Join(entry.Designators, ","),
			entry.Component.Package,
			cid,
		})
	}

	writer.Flush()
}
