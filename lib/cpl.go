package lib

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"
)

/*
	Given a source cpl file, generate a BOM and a result cpl file
*/
func GenerateOutputs(src, bom, cpl string, library *Library) {
	fp, err := os.Open(src)
	if err != nil {
		return
	}
	defer fp.Close()

	fpb, err := os.Create(bom)
	if err != nil {
		return
	}
	defer fpb.Close()

	fpc, err := os.Create(cpl)
	if err != nil {
		return
	}
	defer fpc.Close()

	unpack := func(s []string) (string, string, string, string, string, string, string) {
		return s[0], s[1], s[2], s[3], s[4], s[5], s[6]
	}

	bwriter := csv.NewWriter(fpb)
	cwriter := csv.NewWriter(fpc)

	cwriter.Write([]string{"Designator", "Mid X", "Mid Y", "Layer", "Rotation"})
	bwriter.Write([]string{"Comment", "Designator", "Footprint", "LCSC Part"})
	/*
		Map component numbers to designators
	*/
	dmap := make(map[string][]string)
	cmap := make(map[string]*LibraryComponent)
	cmmap := make(map[string]string)
	//read data into multi-dimentional array of strings
	reader := csv.NewReader(fp)
	for line, _ := reader.Read(); len(line) > 0; line, _ = reader.Read() {
		// G***,LOGO,F4Silkscreen,57.2,-56.9,0.0,top
		// C1,10u,C_0805_2012Metric,30.0,-42.9,90.0,top
		designator, comment, footprint, x, y, rotation, layer := unpack(line)
		if designator == "G***" {
			continue
		}

		re1 := regexp.MustCompile("[^a-zA-Z]+")
		tdesignator := re1.ReplaceAllString(designator, "")

		component := library.FindMatching(tdesignator, comment, footprint)
		if component == nil {
			fmt.Printf("Enter component ID for %s, %s, %s\n:", designator, comment, footprint)

			id := ""
			fmt.Scanln(&id)

			if id == "" {
				continue
			}

			library.Associate(tdesignator, comment, footprint, id)
			component = library.FindMatching(tdesignator, comment, footprint)
		}

		/*
			Then, add it to the designator map
		*/
		if _, ok := dmap[component.ID]; !ok {
			dmap[component.ID] = []string{}
			cmap[component.ID] = component
		}
		dmap[component.ID] = append(dmap[component.ID], designator)
		cmmap[component.ID] = comment

		/*
			Write the component to the position file, since we're keeping it
		*/
		cwriter.Write([]string{designator, x, y, layer, rotation})
		line = []string{}
	}

	for ID, designators := range dmap {
		component := cmap[ID]
		designator := strings.Join(designators, ",")

		bwriter.Write([]string{cmmap[ID], designator, component.Package, ID})
	}

	cwriter.Flush()
	bwriter.Flush()
}
