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
	"strconv"

	"github.com/spf13/cobra"
	"github.com/xoviat/JCAD/lib"
)

// setRotationCmd represents the setRotation command
var setRotationCmd = &cobra.Command{
	Use:   "set-rotation",
	Short: "Set the rotation of a component.",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("set-rotation requires two args")
			return
		}

		id := args[0]
		rotation, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			fmt.Println("failed to parse rotation")
		}

		library, _ := lib.NewDefaultLibrary()
		library.SetRotation(id, rotation)
	},
}

func init() {
	rootCmd.AddCommand(setRotationCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setRotationCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setRotationCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
