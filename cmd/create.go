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
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/xoviat/jcad/lib"
)

var (
	rename bool
)

// createCmd represents the generateLibrary command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an eagle library from the database.",
	Long: `Create an eagle library from the database of symbols.

		Arguments are:
			- category: the category to use
			- output file: the file to output
	`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cname := args[0]
		dst := args[1]
		if !strings.HasSuffix(dst, ".lbr") {
			fmt.Println("dst does not end with lbr")
		}

		/*
			First, get a list of all capacitors in the library
		*/
		library, err := lib.NewDefaultLibrary()
		if err != nil {
			fmt.Printf("failed to encode library: %s\n", err)
			return
		}

		for _, extended := range []bool{false, true} {
			if extended {
				dst = strings.TrimSuffix(dst, ".lbr") + "-Extended.lbr"
			}

			fpdst, err := os.Create(dst)
			if err != nil {
				fmt.Printf("failed to open file: %s\n", err)
				return
			}

			elibrary := lib.NewEagleLibrary()
			/*
				Goal: load an lbr file, and modify the devicesets
			*/

			symbols := make(map[string]*lib.EagleLibrarySymbol)
			packages := make(map[string]*lib.EagleLibraryPackage)

			symbol := library.GetSymbol(cname)
			if symbol.Name == "" || rename {
				fmt.Printf("Enter symbol for %s\n:", cname)
				sname := prompt.Input("> ", func(d prompt.Document) []prompt.Suggest {
					suggestions := []prompt.Suggest{}

					return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
				})

				rename = false
				library.AssociateSymbol(cname, sname)
				symbol = library.GetSymbol(cname)
			}

			symbols[symbol.Name] = symbol

			sets := make(map[string][]*lib.LibraryComponent)
			for _, component := range library.FindInCategory(cname) {
				if !extended && component.LibraryType != "Basic" {
					continue
				}

				value := component.Value()
				if _, ok := sets[value]; !ok {
					sets[value] = []*lib.LibraryComponent{}
				}
				sets[value] = append(sets[value], component)
			}

			devicesets := []*lib.EagleLibraryDeviceSet{}
			for value, set := range sets {
				/*
					<gates>
						<gate name="G$1" symbol="CAP" x="0" y="0"/>
					</gates>
				*/

				deviceset := &lib.EagleLibraryDeviceSet{
					Description: value,
					Name:        value,
					Prefix:      set[0].Prefix(),
					Gates: []*lib.EagleLibraryGate{
						{
							Name:   "G$1",
							Symbol: symbol.Name,
							X:      "0",
							Y:      "0",
						},
					},
				}

				for _, component := range set {
					/*
						<device name="-0603-50V-10%" package="0603">
							<connects>
								<connect gate="G$1" pin="1" pad="1"/>
								<connect gate="G$1" pin="2" pad="2"/>
							</connects>
							<technologies>
								<technology name="">
									<attribute name="PROD_ID" value="CAP-00867"/>
									<attribute name="VALUE" value="10nF"/>
								</technology>
							</technologies>
						</device>
					*/

					if _, ok := packages[component.Package]; !ok {
						packages[component.Package] = library.GetPackage(component.Package)
					}

					deviceset.Devices = append(deviceset.Devices, &lib.EagleLibraryDevice{
						Name:    component.Description,
						Package: component.Package,
						Connects: []*lib.EagleLibraryConnect{
							{Gate: "G$1", Pin: "1", Pad: "1"},
							{Gate: "G$1", Pin: "2", Pad: "2"},
						},
						Technologies: []*lib.EagleLibraryTechnology{
							{
								Attributes: []*lib.EagleLibraryAttribute{
									{Name: "PROD_ID", Value: component.CID()},
									{Name: "VALUE", Value: value},
								},
							},
						},
					})
				}

				devicesets = append(devicesets, deviceset)
			}

			elibrary.DevicesSets = devicesets
			for _, symbol := range symbols {
				elibrary.Symbols = append(elibrary.Symbols, symbol)
			}

			for _, pkg := range packages {
				if pkg.Name == "" {
					continue
				}

				elibrary.Packages = append(elibrary.Packages, pkg)
			}

			enc := xml.NewEncoder(fpdst)
			err = enc.Encode(elibrary)
			if err != nil {
				fmt.Printf("failed to encode library: %s\n", err)
				return
			}

			fpdst.Close()
		}

	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	createCmd.Flags().BoolVarP(&rename, "rename", "r", false, "Whether to rename the references symobol.")
}
