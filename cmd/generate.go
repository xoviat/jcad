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
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/mholt/archiver"
	"github.com/spf13/cobra"
	"github.com/xoviat/jcad/lib"
)

var (
	lclear []string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate manufacturing outputs from a KiCAD board file.",
	Long: `Generate manufacturing outputs for JLCPCB from a KiCAD board file:
		- A ZIP file containing the gerber files used to create the PCB.
		- CPL and BOM files used to place the SMT components.
		
	Example:
		- jcad generate <file.kicad_pcb>`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		library, err := lib.NewDefaultLibrary()
		if err != nil {
			fmt.Printf("failed to obtain default library: %s\n", err)
			return
		}

		kicad, err := lib.NewKicadInterface()
		if err != nil {
			fmt.Printf("failed to obtain kicad instance: %s\n", err)
			return
		}

		pcb, err := lib.NormalizePCB(args[0])
		if err != nil {
			fmt.Println(err.Error())
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

		lib.PrintHeader()
		fmt.Printf("Using KiCad bin path: %s\n", kicad.GetBinPath())
		fmt.Printf("Processing %s\n", pcb)

		os.RemoveAll(filenames.Gerbers)
		os.MkdirAll(filenames.Gerbers, 0777)

		kicad.ExecuteCommand(
			[]string{
				"pcb", "export", "gerbers", filepath.Join("..", filepath.Base(pcb)),
			}, filenames.Gerbers,
		)
		kicad.ExecuteCommand(
			[]string{
				"pcb", "export", "drill", filepath.Join("..", filepath.Base(pcb)),
			}, filenames.Gerbers,
		)
		kicad.ExecuteCommand(
			[]string{
				"pcb", "export", "pos", filepath.Base(pcb),
				"--output", filenames.POS,
				"--units", "mm",
				"--format", "csv",
			}, filepath.Dir(pcb),
		)

		/* map of components to clear */
		mclear := make(map[string]struct{})
		for _, designator := range lclear {
			mclear[designator] = struct{}{}
		}

		client := lib.NewJLC()
		bom := make(lib.BOM)
		components := lib.ReadPOS(filenames.POS)
		assocations := lib.NewAssociationMap(library)

		/*
			filter components that we may possibly assemble
			and retreive component associations from the database
		*/
		for _, component := range components {
			if _, ok := mclear[component.Designator]; ok {
				assocations.Associate(component, nil)
			}
		}

		/*
			retreive associations that we haven't from the user
		*/
		i := 0
		for _, component := range components {
			/*
				if we've marked this as a component to skip
			*/
			if lc := assocations.FindAssociated(component); lc != nil && lc.ID == 0 {
				continue
			}

			if lc := assocations.FindAssociated(component); lc == nil {
				fmt.Printf("Enter component ID for %s, %s, %s\n:", component.Designator, component.Comment, component.Package)
				results, _ := client.SelectComponentList(component.Comment)

				cid := prompt.Input("> ", func(d prompt.Document) []prompt.Suggest {
					suggestions := make([]prompt.Suggest, len(results))

					i := 0
					for _, result := range results {
						suggestions[i] = prompt.Suggest{
							Text: result.CID(), Description: result.Part + " : " + result.Description,
						}
						i++
					}

					return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
				})

				if cid == "" {
					assocations.Associate(component, &lib.LibraryComponent{})
				} else if result, ok := results[lib.FromCID(cid)]; ok {
					assocations.Associate(component, result)
				} else {
					assocations.Associate(component, client.Exact(cid))
				}
			}

			/*
				If the associated part is not in the library, then load it
			*/
			if lc := assocations.FindAssociated(component); lc != nil && lc.ID != 0 && lc.Description == "" {
				fmt.Printf("Loading data from JLCPCB for %s\n", component.Designator)
				assocations.Associate(component, client.Exact(lc.CID()))
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

	generateCmd.PersistentFlags().StringSliceVarP(
		&lclear, "clear", "c", []string{}, "list of component associations to clear",
	)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
