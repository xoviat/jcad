/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

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
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xoviat/jcad/lib"
	"github.com/xuri/excelize/v2"
)

var (
	ifile string
	efile string
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit global or specific component associations.",
	Long: `Edit allows modification, import, or export of existing component assocations.

	Example:
		- jcad edit                  : edit all component associations
		- jcad edit <file.kicad_pcb> : edit component associations for a pcb
		- jcad edit -export <file>   : export all component associations
		- jcad edit -import <file>   : erase and import all component associations
	`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		library, err := lib.NewDefaultLibrary()
		if err != nil {
			fmt.Printf("failed to obtain default library: %s\n", err)
			return
		}

		pcb := ""
		if len(args) > 0 {
			pcb, err = lib.NormalizePCB(args[0])
		}
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		lib.PrintHeader()
		fmt.Println("edit called")

		if efile != "" {
			f := excelize.NewFile()
			f.NewSheet(string(lib.COMPONENTS_ASC_BKT))
			f.DeleteSheet("Sheet1")
			f.SaveAs(efile)

			assocations := library.ExportAssociations()
			i := 1
			for asc := range assocations {
				// fmt.Printf("%s: %s\n", asc[0], asc[1])
				f.SetSheetRow(
					string(lib.COMPONENTS_ASC_BKT),
					"A"+strconv.Itoa(i), &[]interface{}{asc[0], asc[1]},
				)

				i++
			}

			f.Save()
		} else if ifile != "" {
			rows := make(chan []string, 100)
			if !strings.HasSuffix(strings.ToLower(ifile), ".xls") &&
				!strings.HasSuffix(strings.ToLower(ifile), ".xlsx") {

				fmt.Println("association file must be an excel spreadsheet")
				return
			}

			f, err := excelize.OpenFile(ifile)
			if err != nil {
				fmt.Printf("failed to open excel file: %s\n", ifile)
				return
			}

			if f.GetSheetName(0) != string(lib.COMPONENTS_ASC_BKT) {
				fmt.Println("component-associations sheet must be present")
				return
			}

			erows, err := f.Rows(f.GetSheetList()[0])
			if err != nil {
				fmt.Printf("failed to get sheet list: %s\n", ifile)
				return
			}

			go func() {
				for {
					if end := !erows.Next(); end {
						close(rows)
						return
					}

					row, err := erows.Columns()
					if err != nil || len(row) < 2 {
						continue
					}

					rows <- row
				}
			}()

			err = library.ImportAssocations(rows)
			if err != nil {
				fmt.Printf("failed to import library: %s\n", err)
				return
			}
		}

		_ = pcb
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringVarP(&ifile, "import", "i", "", "file to import")
	editCmd.Flags().StringVarP(&efile, "export", "e", "", "file to export")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// editCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// editCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
