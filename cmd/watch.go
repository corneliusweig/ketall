/*
Copyright 2019 Cornelius Weig

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
	"context"
	"github.com/corneliusweig/ketall/pkg/ketall"
	"github.com/corneliusweig/ketall/pkg/ketall/constants"
	"github.com/spf13/cobra"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		catchCtrlC(cancel)
		ketall.WatchAll(ctx, ketallOptions)
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().BoolVar(&ketallOptions.UseCache, constants.FlagUseCache, false, "use cached list of server resources")
	watchCmd.Flags().StringVar(&ketallOptions.Scope, constants.FlagScope, "", "only resources with scope cluster|namespace")
	watchCmd.Flags().StringVarP(&ketallOptions.Selector, constants.FlagSelector, "l", "", "selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	watchCmd.Flags().StringVar(&ketallOptions.FieldSelector, constants.FlagFieldSelector, "", "Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2). The common field queries for all types are metadata.name and metadata.namespace.")
	watchCmd.Flags().StringSliceVar(&ketallOptions.Exclusions, constants.FlagExclude, []string{"events"}, "filter by resource name (plural form or short name)")

	ketallOptions.GenericCliFlags.AddFlags(watchCmd.Flags())
	ketallOptions.PrintFlags.AddFlags(watchCmd)
}
