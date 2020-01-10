// +build !getall

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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/corneliusweig/ketall/internal/ketall/options"
)

func TestMainCompletionCommand(t *testing.T) {
	tests := [][]string{
		{"ketall", "completion", "zsh"},
		{"ketall", "completion", "bash"},
	}

	for _, testargs := range tests {
		t.Run(testargs[2], func(t *testing.T) {
			origOpts := ketallOptions
			newOpts, _, stdout, stderr := options.NewTestTestCmdOptions()

			defer func(args []string) {
				os.Args = args
				ketallOptions = origOpts
			}(os.Args)
			os.Args = testargs
			ketallOptions = newOpts

			err := Execute()

			assert.NoError(t, err)
			assert.NotEmpty(t, stdout.String())
			assert.Empty(t, stderr.String())
		})

	}
}
