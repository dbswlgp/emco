// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "apply(Post) the resources from input file or url(without body) from command line",
	Run: func(cmd *cobra.Command, args []string) {
		var c RestyClient
		if len(token) > 0 {
			c = NewRestClientToken(token[0])
		} else {
			c = NewRestClient()
		}
		if len(inputFiles) > 0 {
			resources := readResources()
			for _, res := range resources {
				if res.file != "" {
					err := c.RestClientMultipartPost(res.anchor, res.body, res.file)
					if HandleError(err, "Apply: ", res.anchor) {
						return
					}
				} else if len(res.files) > 0 {
					err := c.RestClientMultipartPostMultipleFiles(res.anchor, res.body, res.files)
					if HandleError(err, "Apply: ", res.anchor) {
						return
					}
				} else {
					err := c.RestClientPost(res.anchor, res.body)
					if HandleError(err, "Apply: ", res.anchor) {
						return
					}
				}
			}
		} else if len(args) >= 1 {
			c.RestClientPost(args[0], []byte{})
		} else {
			fmt.Println("Use: 'emcoctl apply --help'")
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringSliceVarP(&inputFiles, "filename", "f", []string{}, "Filename of the input file")
	applyCmd.Flags().StringSliceVarP(&valuesFiles, "values", "v", []string{}, "Template Values to go with the input template file")
	applyCmd.Flags().StringSliceVarP(&token, "token", "t", []string{}, "Token for EMCO API")
	applyCmd.Flags().BoolVarP(&stopFlag, "stop", "s", false, "Stop on failure")
	applyCmd.Flags().IntVarP(&acceptWaitTime, "waitTime", "w", 0, "Wait for secs after a call is accepted")
}
