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
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xoviat/jcad/lib"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import databases or libraries.",
	Long: `Import databases or libraries.

		- A JLCPCB component library, in the xlsx format.
		- An Eagle library, in the .lbr format.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		src, err := lib.Normalize(args[0])
		if err != nil {
			fmt.Printf("failed to normalize path: %s\n", src)
			return
		}

		library, err := lib.NewDefaultLibrary()
		if err != nil {
			fmt.Printf("failed to open or create default library: %s\n", err)
			return
		}

		if !lib.Exists(src) {
			fmt.Printf("failed to stat file: %s\n", src)
			return
		}

		if !strings.HasSuffix(src, ".lbr") {
			err := library.Import(src)
			if err != nil {
				fmt.Printf("failed to import library: %s\n", err)
				return
			}
		} else {
			fpsrc, err := os.Open(src)
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

			for _, pkg := range elibrary.Packages {
				fmt.Println("importing package: " + pkg.Name)
			}

			for _, symbol := range elibrary.Symbols {
				fmt.Println("importing symbol: " + symbol.Name)
			}

			library.AddPackages(elibrary.Packages)
			library.AddSymbols(elibrary.Symbols)
		}
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
