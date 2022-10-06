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
	"text/template"

	"github.com/corneliusweig/ketall/pkg/version"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

const (
	versionTemplate = `{{.Version}}
`
	fullInfoTemplate = `ketall:     {{.Version}}
platform:   {{.Platform}}
git commit: {{.GitCommit}}
build date: {{.BuildDate}}
go version: {{.GoVersion}}
compiler:   {{.Compiler}}
`
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Args:  cobra.NoArgs,
	Run:   runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolP("full", "f", false, "print extended version information")
}

func runVersion(cmd *cobra.Command, _ []string) {
	var tpl string

	if cmd.Flag("full").Changed {
		tpl = fullInfoTemplate
	} else {
		tpl = versionTemplate
	}

	var t = template.Must(template.New("info").Parse(tpl))

	if err := t.Execute(ketallOptions.Streams.Out, version.GetBuildInfo()); err != nil {
		klog.Warning("Could not print version info")
	}
}
