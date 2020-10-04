package lib

import (
	"bufio"
	"os"
	"strings"
)

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

/*
	Return a list of components, given a schematic
*/
func ParseSchematic(src string) []*SchematicComponent {
	fp, err := os.Open(src)
	if err != nil {
		return []*SchematicComponent{}
	}

	components := []*SchematicComponent{}
	scanner := bufio.NewScanner(fp)
	scanning := true
	for scanning {
		text := ""
		for scanning = scanner.Scan(); scanning && text != "$Comp"; scanning = scanner.Scan() {
			text = strings.TrimSpace(scanner.Text())
		}

		component := SchematicComponent{}
		for scanning = scanner.Scan(); scanning && text != "$EndComp"; scanning = scanner.Scan() {
			text = strings.TrimSpace(scanner.Text())
			component.lines = append(component.lines, strings.Split(text, " "))
		}

		components = append(components, &component)
	}

	return components
}
