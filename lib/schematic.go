package lib

import "io/ioutil"

type Schematic struct {
	version string
	blocks  []SchematicBlock
}

func (s *Schematic) Compoonents() []*SchematicComponent {
	compoents := make([]*SchematicComponent, 10)
	for _, block := range s.blocks {
		if component, ok := block.(*SchematicComponent); ok {
			compoents = append(compoents, component)
		}
	}

	return compoents
}

func (s *Schematic) Text() string {
	text := ""
	for _, block := range s.blocks {
		text += block.Text() + "\n"
	}

	return text
}

type SchematicBlock interface {
	Text() string
}

/*
	Represents other data present in the sch file that is of no concern to us
*/
type SchematicText struct {
	/*
		EELAYER 30 0
		EELAYER END
		$Descr A4 11693 8268
		encoding utf-8
		Sheet 1 1
		Title ""
		Date ""
		Rev ""
		Comp ""
		Comment1 ""
		Comment2 ""
		Comment3 ""
		Comment4 ""
		$EndDescr
	*/
	text string
}

func (st *SchematicText) Text() string {
	return st.text
}

/*
	Represents a schematic component
*/
type SchematicComponent struct {
	lines [][]string
	/*
		$Comp
		L Regulator_Linear:AMS1117-3.3 U1
		U 1 1 5E7A1557
		P 2650 1150
		F 0 "U1" H 2650 1392 50  0000 C CNN
		F 1 "AMS1117-3.3" H 2650 1301 50  0000 C CNN
		F 2 "Package_TO_SOT_SMD:SOT-223-3_TabPin2" H 2650 1350 50  0001 C CNN
		F 3 "http://www.advanced-monolithic.com/pdf/ds1117.pdf" H 2750 900 50  0001 C CNN
			1    2650 1150
			1    0    0    -1
		$EndComp
	*/

}

func (st *SchematicComponent) Designator() string {
	return ""
}

func (st *SchematicComponent) Value() string {
	return ""
}

func (st *SchematicComponent) Footprint() string {
	return ""
}

func (st *SchematicComponent) Part() string {
	return ""
}

func (st *SchematicComponent) Text() string {
	text := "$Comp"
	for _, line := range st.lines {
		for _, part := range line {
			text += " " + part
		}
		text += "\n"
	}
	text += "$EndComp"

	return text
}

func LoadSchematic(src string) *Schematic {
	return nil
}

func SaveSchematic(src string, sch *Schematic) {
	ioutil.WriteFile(src, []byte(sch.Text()), 0777)
}
