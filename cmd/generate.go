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
	"regexp"

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

		lib.ExecuteScript("generate_cpl.py", []string{pcb, "test-data/temp/data.cpl"})
		// lib.GenerateOutputs("test-data/temp/data.cpl", "test-data/temp/bom.csv", "test-data/temp/cpl.csv", library)

		/*
			Map component numbers to designators
		*/

		/*
			Includes ONLY SMT components
		*/
		components := []*lib.BoardComponent{}
		entries := map[string]*lib.BOMEntry{}

		re1 := regexp.MustCompile("[^a-zA-Z]+")
		for _, component := range lib.ReadCPL("test-data/temp/data.cpl") {
			tdesignator := re1.ReplaceAllString(component.Designator, "")

			lcomponent := library.FindMatching(tdesignator, component.Comment, component.Footprint)
			if lcomponent == nil {
				fmt.Printf("Enter component ID for %s, %s, %s\n:", component.Designator, component.Comment, component.Footprint)

				id := ""
				fmt.Scanln(&id)

				if id == "" {
					continue
				}

				library.Associate(tdesignator, component.Comment, component.Footprint, id)
				lcomponent = library.FindMatching(tdesignator, component.Comment, component.Footprint)
			}

			components = append(components, component)

			/*
				Then, add it to the designator map
			*/
			if _, ok := entries[lcomponent.ID]; !ok {
				entries[lcomponent.ID] = &lib.BOMEntry{}
				entries[lcomponent.ID].Comment = component.Comment
				entries[lcomponent.ID].Component = lcomponent
			}
			entries[lcomponent.ID].Designators = append(entries[lcomponent.ID].Designators, component.Designator)
		}

		sentries := []*lib.BOMEntry{}
		for _, entry := range entries {
			sentries = append(sentries, entry)
		}

		lib.WriteBOM("test-data/temp/bom.csv", sentries)
		lib.WriteCPL("test-data/temp/cpl.csv", components)

		lib.ExecuteScript("generate_gerbers.py", []string{pcb, "test-data/temp/gerbers"})
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
