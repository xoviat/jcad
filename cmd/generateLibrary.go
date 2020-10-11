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

	"github.com/spf13/cobra"
	"github.com/xoviat/JCAD/lib"
)

// generateLibraryCmd represents the generateLibrary command
var generateLibraryCmd = &cobra.Command{
	Use:   "generate-library",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Println("invalid number of args")
			return
		}

		src := args[0]
		fpsrc, err := os.Open(src)
		if err != nil {
			fmt.Printf("failed to open file: %s\n", err)
			return
		}

		dst := args[1]
		fpdst, err := os.Create(dst)
		if err != nil {
			fmt.Printf("failed to open file: %s\n", err)
			return
		}

		elibrary := lib.EagleLibrary{}
		dec := xml.NewDecoder(fpsrc)
		err = dec.Decode(&elibrary)
		if err != nil {
			fmt.Printf("failed to decode library: %s\n", err)
			return
		}

		/*
			Goal: load an lbr file, and modify the devicesets
		*/

		/*
			First, get a list of all capacitors in the library
		*/
		library, err := lib.NewDefaultLibrary()
		if err != nil {
			fmt.Printf("failed to encode library: %s\n", err)
			return
		}

		sets := make(map[string][]*lib.LibraryComponent)
		for _, capacitor := range library.FindInCategory("Capacitors") {
			value := capacitor.Value()
			if _, ok := sets[value]; !ok {
				sets[value] = []*lib.LibraryComponent{}
			}
			sets[value] = append(sets[value], capacitor)
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
				Prefix:      "C",
				Gates: []*lib.EagleLibraryGate{
					{
						Name:   "G$1",
						Symbol: "CAP",
						X:      "0",
						Y:      "0",
					},
				},
			}

			for _, capacitor := range set {
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

				deviceset.Devices = append(deviceset.Devices, &lib.EagleLibraryDevice{
					Name:    capacitor.Description,
					Package: capacitor.Package,
					Connects: []*lib.EagleLibraryConnect{
						{Gate: "G$1", Pin: "1", Pad: "1"},
						{Gate: "G$1", Pin: "2", Pad: "2"},
					},
					Technologies: []*lib.EagleLibraryTechnology{
						{
							Attributes: []*lib.EagleLibraryAttribute{
								{Name: "PROD_ID", Value: capacitor.CID()},
								{Name: "VALUE", Value: value},
							},
						},
					},
				})
			}

			devicesets = append(devicesets, deviceset)
		}

		elibrary.DevicesSets = devicesets

		enc := xml.NewEncoder(fpdst)
		err = enc.Encode(elibrary)
		if err != nil {
			fmt.Printf("failed to encode library: %s\n", err)
			return
		}

		fpsrc.Close()
		fpdst.Close()
	},
}

func init() {
	rootCmd.AddCommand(generateLibraryCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateLibraryCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateLibraryCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
