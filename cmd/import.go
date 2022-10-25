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
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xoviat/jcad/lib"
	"github.com/xuri/excelize/v2"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a component or association database.",
	Long: `Import one of the following:
		- A component database, in an xlsx or csv format. 
		- An association database, in an xlsx format.`,
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

		isAssociations := false
		rows := make(chan []string, 100)
		if strings.HasSuffix(strings.ToLower(src), ".csv") {
			fp, err := os.Open(src)
			if err != nil {
				fmt.Printf("failed to open csv file: %s\n", src)
				return
			}

			reader := csv.NewReader(bufio.NewReader(fp))
			go func() {
				defer fp.Close()

				for row, _ := reader.Read(); len(row) > 0; row, _ = reader.Read() {
					if len(row) < 9 {
						continue
					}

					rows <- row
				}

				close(rows)
			}()
		} else if strings.HasSuffix(strings.ToLower(src), ".xls") ||
			strings.HasSuffix(strings.ToLower(src), ".xlsx") {

			f, err := excelize.OpenFile(src)
			if err != nil {
				fmt.Printf("failed to open excel file: %s\n", src)
				return
			}

			isAssociations = f.GetSheetName(0) == "component-associations"
			erows, err := f.Rows(f.GetSheetList()[0])
			if err != nil {
				fmt.Printf("failed to get sheet list: %s\n", src)
				return
			}

			go func() {
				for {
					if end := !erows.Next(); end {
						close(rows)
						return
					}

					row, err := erows.Columns()
					if err != nil || len(row) < 2 || (!isAssociations && len(row) < 9) {
						continue
					}

					rows <- row
				}
			}()
		} else {
			fmt.Printf("unknown file type: %s\n", src)
			return
		}

		// TODO: Import assocations if sheet name is component-assocations
		//  library.ImportAssocations

		if isAssociations {
			err = library.ImportAssocations(rows)
		} else {
			err = library.Import(rows)
		}
		if err != nil {
			fmt.Printf("failed to import library: %s\n", err)
			return
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
