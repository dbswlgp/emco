// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the resources from input file or url(without body) from command line",
	Run: func(cmd *cobra.Command, args []string) {
		var c RestyClient
		if len(token) > 0 {
			c = NewRestClientToken(token[0])
		} else {
			c = NewRestClient()
		}
		if len(inputFiles) > 0 {
			resources := readResources()
			for i := len(resources) - 1; i >= 0; i-- {
				res := resources[i]
				err := c.RestClientDelete(res.anchor, res.body)
				if HandleError(err, "Delete: ", res.anchor) {
					return
				}
			}
		} else if len(args) >= 1 {
			c.RestClientDeleteAnchor(args[0])
		} else {
			fmt.Println("Use: 'emcoctl delete --help'")
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringSliceVarP(&inputFiles, "filename", "f", []string{}, "Filename of the input file")
	deleteCmd.Flags().StringSliceVarP(&valuesFiles, "values", "v", []string{}, "Template Values to go with the input template file")
	deleteCmd.Flags().StringSliceVarP(&token, "token", "t", []string{}, "Token for EMCO API")
	deleteCmd.Flags().BoolVarP(&stopFlag, "stop", "s", false, "Stop on failure")
	deleteCmd.Flags().IntVarP(&acceptWaitTime, "waitTime", "w", 0, "Wait for secs after a call is accepted")
}
