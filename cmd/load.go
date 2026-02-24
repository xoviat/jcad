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

	"github.com/spf13/cobra"
	"github.com/xoviat/jcad/lib"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load the basic parts list",
	Long:  `Load the basic parts list from the JLCBPCB website.`,
	Run: func(cmd *cobra.Command, args []string) {
		library, err := lib.NewDefaultLibrary(false)
		if err != nil {
			fmt.Printf("failed to open or create default library: %s\n", err)
			return
		}

		fmt.Println("loading basic components from JLCPCB")
		client := lib.NewJLC()

		components, errs := client.SelectBaseComponentList()

		err = library.ImportBasic(components, errs)
		if err != nil {
			fmt.Println("failed to load basic compnent list")
			return
		}

	},
}

func init() {
	rootCmd.AddCommand(loadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
