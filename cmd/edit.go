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
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"

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
		- jcad edit --export <file>   : export all component associations
		- jcad edit --import <file>   : erase and import all component associations
	`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		export := func(library *lib.Library, path string) error {
			fmt.Println("preparing to export component associations to file...")

			f := excelize.NewFile()
			if _, err := f.NewSheet(string(lib.COMPONENTS_ASC_BKT)); err != nil {
				return fmt.Errorf("failed to create new sheet: %s", err)
			}

			if err := f.DeleteSheet("Sheet1"); err != nil {
				return fmt.Errorf("failed to delete sheet: %s", err)
			}

			if err := f.SaveAs(path); err != nil {
				return fmt.Errorf("failed to save as: %s", err)
			}

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

			return f.Save()
		}

		fimport := func(library *lib.Library, path string) error {
			fmt.Println("preparing to import component associations from file...")

			rows := make(chan []string, 100)
			if !strings.HasSuffix(strings.ToLower(path), ".xls") &&
				!strings.HasSuffix(strings.ToLower(path), ".xlsx") {

				return fmt.Errorf("association file must be an excel spreadsheet")
			}

			f, err := excelize.OpenFile(path)
			if err != nil {
				return fmt.Errorf("failed to open excel file: %s (%s)\n", path, err)
			}

			if f.GetSheetName(0) != string(lib.COMPONENTS_ASC_BKT) {
				return fmt.Errorf("component-associations sheet must be present")
			}

			erows, err := f.Rows(f.GetSheetList()[0])
			if err != nil {
				return fmt.Errorf("failed to get sheet list: %s (%s)\n", ifile, err)
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
				return fmt.Errorf("failed to import library: %s\n", err)
			}

			return nil
		}

		library, err := lib.NewDefaultLibrary(true)
		if err != nil {
			fmt.Printf("failed to obtain default library: %s\n", err)
			return
		}

		pcb := ""
		if len(args) > 0 && args[0] != "" {
			pcb, err = lib.NormalizePCB(args[0])
		}
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		lib.PrintHeader()
		if pcb != "" {
			fmt.Println("editing components for pcb is not implemented")
		}

		if efile != "" {
			if err := export(library, efile); err != nil {
				fmt.Printf("failed to export lib: %s\n", err)
				return
			}
		} else if ifile != "" {
			if err := fimport(library, efile); err != nil {
				fmt.Printf("failed to import lib: %s\n", err)
				return
			}
		} else {
			tempf, err := os.CreateTemp("", "associations-*.xlsx")
			if err != nil {
				fmt.Printf("Error creating temp file: %s\n", err)
				return
			}

			defer os.Remove(tempf.Name())

			if err = tempf.Close(); err != nil {
				fmt.Printf("failed to close file: %s\n", err)
				return
			}

			if err := export(library, tempf.Name()); err != nil {
				fmt.Printf("failed to export lib: %s\n", err)
				return
			}

			cmd, err := lib.OpenFile(tempf.Name())
			if err != nil {
				fmt.Printf("failed to launch editor: %s\n", err)
				return
			}

			if err := cmd.Wait(); err != nil {
				fmt.Printf("failed to wait for cmd: %s\n", err)
				return
			}

			for {
				time.Sleep(500 * time.Millisecond)
				if _, err := os.OpenFile(tempf.Name(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, fs.ModeExclusive); err != nil {
					continue
				}

				break
			}

			if err := fimport(library, tempf.Name()); err != nil {
				fmt.Printf("failed to import lib: %s\n", err)
				return
			}
		}
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
