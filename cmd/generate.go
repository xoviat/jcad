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
	"github.com/xoviat/JCAD/lib"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("len(args) < 1")
			return
		}

		pcb := args[0]

		library, _ := lib.NewDefaultLibrary()

		rname := strings.TrimSuffix(filepath.Base(pcb), path.Ext(pcb))
		scpl := filepath.Join(filepath.Dir(pcb), rname+"-data.cpl")
		bom := filepath.Join(filepath.Dir(pcb), rname+"-BOM.csv")
		cpl := filepath.Join(filepath.Dir(pcb), rname+"-all-pos.csv")
		gerbers := filepath.Join(filepath.Dir(pcb), rname+"-gerber")
		zip := filepath.Join(filepath.Dir(pcb), rname+"-gerber.zip")

		lib.ExecuteScript("generate_cpl.py", []string{pcb, scpl})

		smap := make(map[string]bool)

		/*
			Includes ONLY SMT components
		*/
		components := []*lib.BoardComponent{}
		entries := map[string]*lib.BOMEntry{}
		for _, component := range lib.ReadCPL(scpl) {
			if !library.CanAssemble(component) {
				continue
			}

			lcomponent := library.FindAssociated(component)
			if lcomponent == nil {
				sKey := string(component.Key())
				if _, ok := smap[sKey]; ok {
					continue
				}

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
					smap[sKey] = true
					continue
				}

				lcomponent = library.Exact(id)
				library.Associate(component, lcomponent)
			}

			if lcomponent == nil {
				fmt.Println("notice: unexpected condition")
				continue
			}

			components = append(components, component)

			/*
				Then, add it to the designator map
			*/
			if _, ok := entries[lcomponent.CID()]; !ok {
				entries[lcomponent.CID()] = &lib.BOMEntry{}
				entries[lcomponent.CID()].Comment = component.Comment
				entries[lcomponent.CID()].Component = lcomponent
			}
			entries[lcomponent.CID()].Designators = append(entries[lcomponent.CID()].Designators, component.Designator)

			rotation, err := strconv.ParseFloat(component.Rotation, 64)
			if err != nil {
				fmt.Printf("failed to parse board component rotation: %s\n", component.Rotation)
				continue
			}

			rotation += lcomponent.Rotation
			component.Rotation = fmt.Sprintf("%.1f", rotation)
		}

		sentries := []*lib.BOMEntry{}
		for _, entry := range entries {
			sentries = append(sentries, entry)
		}

		lib.WriteBOM(bom, sentries)
		lib.WriteCPL(cpl, components)

		lib.ExecuteScript("generate_gerbers.py", []string{pcb, gerbers})

		os.Remove(zip)
		archiver.Archive([]string{gerbers}, zip)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
