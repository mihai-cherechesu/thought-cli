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
	"github.com/MihaiCherechesu/thought-cli/pkg"
	"github.com/spf13/cobra"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lists the running services in a table format",
	Long: `Lists the running services in a table format that comprises metadata similar to the one below:

	┏━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━┳━━━━━┳━━━━━━━━┳━━━━━━━━━━━┓
	┃ IP          ┃ SERVICE            ┃ CPU ┃ MEMORY ┃ STATUS    ┃
	┣━━━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━╋━━━━━╋━━━━━━━━╋━━━━━━━━━━━┫
	┃ 10.58.1.68  ┃ GeoService         ┃ 51% ┃ 76%    ┃ Healthy   ┃
	┃ 10.58.1.144 ┃ GeoService         ┃ 81% ┃ 8%     ┃ Healthy   ┃
	┃ 10.58.1.20  ┃ GeoService         ┃ 43% ┃ 9%     ┃ Healthy   ┃
	┃ 10.58.1.94  ┃ GeoService         ┃ 65% ┃ 98%    ┃ Unhealthy ┃
	┃ 10.58.1.126 ┃ GeoService         ┃ 40% ┃ 36%    ┃ Healthy   ┃
	┃ 10.58.1.67  ┃ GeoService         ┃ 17% ┃ 3%     ┃ Healthy   ┃`,

	PreRun: func(cmd *cobra.Command, args []string) {
		follow, _ := cmd.Flags().GetBool("follow")
		if follow {
			cmd.MarkFlagRequired("service")
		}
	},

	Run: func(cmd *cobra.Command, args []string) {
		follow, _ := cmd.Flags().GetBool("follow")
		service, _ := cmd.Flags().GetString("service")
		merged, _ := cmd.Flags().GetBool("merged")
		pkg.RunLs(follow, service, merged)
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
	lsCmd.PersistentFlags().BoolP("follow", "f", false, "whether or not to follow the output for the service specified with --service.")
	lsCmd.PersistentFlags().StringP("service", "s", "", "service for which to list details. If the flag is not specified (default behaviour), all services are listed.")
	lsCmd.PersistentFlags().BoolP("merged", "m", false, "whether or not to have a merged output for the service(s). ")
}
