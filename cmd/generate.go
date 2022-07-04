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

		bom := make(lib.BOM)
		components := lib.ReadKCPL(filenames.KCPL)
		assocations := make(map[string]*lib.LibraryComponent)

		/*
			filter components that we may possibly assemble
			and retreive component associations from the database
		*/
		i := 0
		for _, component := range components {
			if !library.CanAssemble(component) {
				i++
				continue
			}

			if _, ok := mredesignations[component.Designator]; ok {
				library.Associate(component, nil)
				delete(assocations, string(component.Key()))
			}

			if _, ok := assocations[string(component.Key())]; !ok {
				assocations[string(component.Key())] = library.FindAssociated(component)
			}

			component.LibraryComponent = assocations[string(component.Key())]
			components[i] = component

			if rotation, ok := mrotations[component.Designator]; ok {
				library.SetRotation(component.LibraryComponent, rotation)
			}
			i++
		}
		components = components[:i]

		/*
			retreive associations that we haven't from the user
		*/
		i = 0
		for _, component := range components {
			/*
				if we already retreived the assocation from the user
			*/
			if _, ok := assocations[string(component.Key())]; ok && component.LibraryComponent == nil {
				component.LibraryComponent = assocations[string(component.Key())]
			}

			/*
				if we have marked this as a component to skip
			*/
			if component.LibraryComponent != nil && component.LibraryComponent.ID == 0 {
				i++
				continue
			}

			if component.LibraryComponent == nil {
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
					component.LibraryComponent = &lib.LibraryComponent{}
				} else {
					component.LibraryComponent = library.Exact(id)
				}

				assocations[string(component.Key())] = component.LibraryComponent
				library.Associate(component, component.LibraryComponent)
			}

			if component.LibraryComponent == nil {
				panic("unexpected condition")
			}

			/*
				if we have marked this as a component to skip
			*/
			if component.LibraryComponent.ID == 0 {
				i++
				continue
			}

			components[i] = component
			/*
				add the component to the BOM
			*/
			bom.AddComponent(component)

			/*
				increase the rotation of the component by the preset amount
			*/
			if component.LibraryComponent.Rotation != 0 {
				component.Rotate(component.LibraryComponent.Rotation)
			}

			i++
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
