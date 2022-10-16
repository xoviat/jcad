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
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xoviat/jcad/lib"
	"github.com/xuri/excelize/v2"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export databases or libraries.",
	Long:  `Export an association database in the xlsx format.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		src := args[0]
		if !strings.HasSuffix(src, "xlsx") && !strings.HasSuffix(src, "xls") {
			fmt.Printf("export file name must be excel file\n")
			return
		}

		library, err := lib.NewDefaultLibrary()
		if err != nil {
			fmt.Printf("failed to open or create default library: %s\n", err)
			return
		}

		f := excelize.NewFile()
		f.NewSheet(string(lib.COMPONENTS_ASC_BKT))
		f.DeleteSheet("Sheet1")
		f.SaveAs(src)

		assocations := library.ExportAssociations()
		i := 1
		for asc := range assocations {
			fmt.Printf("%s: %s\n", asc[0], asc[1])
			f.SetSheetRow(
				string(lib.COMPONENTS_ASC_BKT),
				"A"+strconv.Itoa(i), &[]interface{}{asc[0], asc[1]},
			)

			i++
		}

		f.Save()
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// exportCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// exportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
