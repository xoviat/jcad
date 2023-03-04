/*
Copyright © 2020 Mars Galactic <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/mholt/archiver"
	"github.com/spf13/cobra"
	"github.com/xoviat/jcad/lib"
)

var (
	lredesignate string
	lrotate      string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate manufacturing outputs from a KiCAD board file.",
	Long: `Generate manufacturing outputs for JLCPCB from a KiCAD board file:
		- A ZIP file containing the gerber files used to create the PCB.
		- CPL and BOM files used to place the SMT components.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pcb, err := lib.Normalize(args[0])
		if err != nil {
			fmt.Printf("failed to normalize path: %s\n", err)
			return
		}

		if !strings.HasSuffix(pcb, ".kicad_pcb") {
			fmt.Println("pcb does not exist or is not KiCad PCB")
			return
		}

		library, err := lib.NewDefaultLibrary()
		if err != nil {
			fmt.Printf("failed to open or create default library: %s\n", err)
			return
		}

		rname := strings.TrimSuffix(filepath.Base(pcb), path.Ext(pcb))
		filenames := struct {
			Name    string
			POS     string
			BOM     string
			CPL     string
			Gerbers string
			ZIP     string
		}{
			Name:    rname,
			POS:     filepath.Join(filepath.Dir(pcb), rname+"-data.pos"),
			BOM:     filepath.Join(filepath.Dir(pcb), rname+"-BOM.csv"),
			CPL:     filepath.Join(filepath.Dir(pcb), rname+"-all-pos.csv"),
			Gerbers: filepath.Join(filepath.Dir(pcb), rname+"-gerber"),
			ZIP:     filepath.Join(filepath.Dir(pcb), rname+"-gerber.zip"),
		}

		fmt.Println("JCAD: KiCAD -> JLCPCB PCB assembly output generator")
		fmt.Printf("Processing %s\n", pcb)

		os.RemoveAll(filenames.Gerbers)
		os.MkdirAll(filenames.Gerbers, 0777)

		lib.ExecuteKiCadCommand(
			[]string{
				"pcb", "export", "gerbers", filepath.Join("..", filepath.Base(pcb)),
			}, filenames.Gerbers,
		)
		lib.ExecuteKiCadCommand(
			[]string{
				"pcb", "export", "drill", filepath.Join("..", filepath.Base(pcb)),
			}, filenames.Gerbers,
		)
		lib.ExecuteKiCadCommand(
			[]string{
				"pcb", "export", "pos", filepath.Base(pcb),
				"--output", filenames.POS,
				"--units", "mm",
				"--format", "csv",
			}, filepath.Dir(pcb),
		)

		mredesignations := make(map[string]bool)
		mrotations := make(map[string]float64)

		if lredesignate != "" {
			redesignations := strings.Split(lredesignate, ",")
			for _, redesignation := range redesignations {
				mredesignations[redesignation] = true
			}
		}

		if lrotate != "" {
			rotations := strings.Split(lrotate, ",")
			for i := 0; i < len(rotations); i += 2 {
				rotation, err := strconv.ParseFloat(rotations[i+1], 64)
				if err != nil {
					fmt.Printf("failed to parse board component rotation: %1.2f\n", rotation)
					continue
				}

				mrotations[rotations[i]] = rotation
			}
		}

		bom := make(lib.BOM)
		components := lib.ReadPOS(filenames.POS)
		assocations := lib.NewAssociationMap(library)

		/*
			filter components that we may possibly assemble
			and retreive component associations from the database
		*/
		i := 0
		for _, component := range components {
			if _, ok := mredesignations[component.Designator]; ok {
				assocations.Associate(component, nil)
			}

			components[i] = component
			i++

			//			if component.LibraryComponent != nil {
			//				fmt.Printf(
			//					"key %s associated with %s\n",
			//					component.StringKey(), component.LibraryComponent.CID(),
			//				)
			//			}

			if rotation, ok := mrotations[component.Designator]; ok {
				library.SetRotation(assocations.FindAssociated(component), rotation)
			}
		}
		components = components[:i]

		/*
			retreive associations that we haven't from the user
		*/
		i = 0
		for _, component := range components {
			/*
				if we've marked this as a component to skip
			*/
			if lc := assocations.FindAssociated(component); lc != nil && lc.ID == 0 {
				continue
			}

			if lc := assocations.FindAssociated(component); lc == nil {
				fmt.Printf("Enter component ID for %s, %s, %s\n:", component.Designator, component.Comment, component.Package)
				id := prompt.Input("> ", func(d prompt.Document) []prompt.Suggest {
					suggestions := []prompt.Suggest{}
					for _, lcomponent := range library.FindMatching(component) {
						suggestions = append(suggestions, prompt.Suggest{
							Text: lcomponent.CID(), Description: lcomponent.Part + " " + lcomponent.Package,
						})
					}

					return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
				})

				if id == "" {
					assocations.Associate(component, &lib.LibraryComponent{})
				} else {
					assocations.Associate(component, library.Exact(id))
				}
			}

			/*
				if we've marked this as a component to skip
			*/
			if lc := assocations.FindAssociated(component); lc != nil && lc.ID == 0 {
				continue
			}

			components[i] = component
			i++

			lc := assocations.FindAssociated(component)
			/*
				add the component to the BOM
			*/
			bom.AddComponent(component, lc)

			/*
				increase the rotation of the component by the preset amount

				rotate is always called in order to normalize within 360 degrees
			*/
			component.Rotate(lc.Rotation)
		}
		components = components[:i]

		lib.WriteBOM(filenames.BOM, bom)
		lib.WriteCPL(filenames.CPL, components)

		os.Remove(filenames.ZIP)
		archiver.Archive([]string{filenames.Gerbers}, filenames.ZIP)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.PersistentFlags().StringVarP(&lredesignate, "redesignate", "d", "", "components to redesignate")
	generateCmd.PersistentFlags().StringVarP(&lrotate, "rotate", "r", "", "components to rotate")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
