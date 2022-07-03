/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	Short: "Generate JLCPCB input files, given a KiCAD board file.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pcb, err := lib.Normalize(args[0])
		if err != nil {
			fmt.Printf("failed to normalize path: %s\n", err)
			return
		}

		if !lib.Exists(pcb) || !strings.HasSuffix(pcb, ".kicad_pcb") {
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
			KCPL    string
			BOM     string
			CPL     string
			Gerbers string
			ZIP     string
		}{
			Name:    rname,
			KCPL:    filepath.Join(filepath.Dir(pcb), rname+"-data.cpl"),
			BOM:     filepath.Join(filepath.Dir(pcb), rname+"-BOM.csv"),
			CPL:     filepath.Join(filepath.Dir(pcb), rname+"-all-pos.csv"),
			Gerbers: filepath.Join(filepath.Dir(pcb), rname+"-gerber"),
			ZIP:     filepath.Join(filepath.Dir(pcb), rname+"-gerber.zip"),
		}

		fmt.Println("JCAD: KiCAD -> JLCPCB PCB assembly output generator")
		fmt.Printf("Processing %s\n", pcb)

		os.RemoveAll(filenames.Gerbers)
		os.MkdirAll(filenames.Gerbers, 0777)

		lib.ExecuteScript("generate_cpl.py", []string{pcb, filenames.KCPL})
		lib.ExecuteScript("generate_gerbers.py", []string{pcb, filenames.Gerbers})

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

		/*
			Includes ONLY SMT components
		*/
		components := []*lib.BoardComponent{}
		bom := make(lib.BOM)
		clist := lib.ReadKCPL(filenames.KCPL)

		for _, component := range clist {
			if !library.CanAssemble(component) {
				continue
			}

			if rotation, ok := mrotations[component.Designator]; ok {
				library.SetRotation(library.FindAssociated(component), rotation)
			}

			if _, ok := mredesignations[component.Designator]; ok {
				library.Associate(component, nil)
			}
		}

		/*
			TODO: Read ahead the board components and reorganize the redesignations and rotations
		*/
		for _, component := range clist {
			if !library.CanAssemble(component) {
				continue
			}

			lcomponent := library.FindAssociated(component)
			/*
				If we have marked this as a component to skip
			*/
			if lcomponent != nil && lcomponent.ID == 0 {
				continue
			}

			if lcomponent == nil {
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
					lcomponent = &lib.LibraryComponent{}
				} else {
					lcomponent = library.Exact(id)
				}

				library.Associate(component, lcomponent)
			}

			if lcomponent == nil {
				fmt.Println("notice: unexpected condition")
				continue
			}

			components = append(components, component)

			/*
				Then, add it to the BOM
			*/
			bom.AddComponent(&lib.LinkedComponent{
				BoardComponent:   component,
				LibraryComponent: lcomponent,
			})

			if lcomponent.Rotation == 0 {
				continue
			}

			component.Rotate(lcomponent.Rotation)
		}

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
