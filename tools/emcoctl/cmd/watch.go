// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch for resource status notifications from input file or url from command line",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) >= 1 {
			WatchGrpcEndpoint(args...)
		} else {
			fmt.Println("Use: 'emcoctl watch --help'")
		}
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().StringSliceVarP(&inputFiles, "filename", "f", []string{}, "Filename of the input file")
	watchCmd.Flags().StringSliceVarP(&valuesFiles, "values", "v", []string{}, "Template Values to go with the input template file")
	watchCmd.Flags().StringSliceVarP(&token, "token", "t", []string{}, "Token for EMCO API")
}
