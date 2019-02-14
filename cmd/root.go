// Copyright © 2019 Cornelius Weig <cornelius.weig@tngtech.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/corneliusweig/ketall/pkg/options"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"

	"github.com/corneliusweig/ketall/pkg"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ketallOptions = options.NewCmdOptions()
	v             string
)

const (
	ketallLongDescription = `
ketall retrieves all resources which allow to be fetched. This complements the
usual "kubectl get all" command, which does not list cluster-level resources.
`
	ketallExamples = `
  Get all resources
  $ ketall

  Get all resources and use list of cached server resources
  $ ketall --cache
`
)

var rootCmd = &cobra.Command{
	Use:     "ketall",
	Short:   "Get all resources",
	Long:    ketallLongDescription,
	Args:    cobra.NoArgs,
	Example: ketallExamples,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.KetAll(ketallOptions)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal("Ececution failed:", err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&ketallOptions.CfgFile, "config", "", "config file (default is $HOME/.kube/ketall.yaml)")
	rootCmd.PersistentFlags().StringVarP(&v, "verbosity", "v", pkg.DefaultLogLevel.String(), "Log level (debug, info, warn, error, fatal, panic)")

	rootCmd.Flags().BoolVar(&ketallOptions.UseCache, "cache", false, "use cached list of server resources")

	ketallOptions.GenericCliFlags.AddFlags(rootCmd.Flags())
	ketallOptions.PrintFlags.AddFlags(rootCmd)

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := SetUpLogs(os.Stderr, v); err != nil {
			return err
		}
		return nil
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if ketallOptions.CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(ketallOptions.CfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			logrus.Warn("Could not read home dir: %s", err)
			return
		}

		// Search config in "~/.kube/ketall" (without extension).
		viper.AddConfigPath(filepath.Join(home, ".kube"))
		viper.SetConfigName("ketall")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logrus.Debug("Using config file:", viper.ConfigFileUsed())
	}
}

func SetUpLogs(out io.Writer, level string) error {
	logrus.SetOutput(out)
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return errors.Wrap(err, "parsing log level")
	}
	logrus.SetLevel(lvl)
	logrus.Debugf("Set log-level to %s", level)
	return nil
}
