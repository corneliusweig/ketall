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

	"github.com/corneliusweig/ketall/internal/ketall/options"
	"github.com/stretchr/testify/assert"
)

func TestMainHelp(t *testing.T) {
	origOpts := ketallOptions
	newOpts, _, stdout, stderr := options.NewTestTestCmdOptions()

	defer func(args []string) {
		os.Args = args
		ketallOptions = origOpts
	}(os.Args)
	os.Args = []string{"ketall", "help"}
	ketallOptions = newOpts

	err := Execute()

	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "Available Commands:")
	assert.Empty(t, stderr.String())
}

func TestMainUnknownCommand(t *testing.T) {
	origOpts := ketallOptions
	newOpts, _, _, _ := options.NewTestTestCmdOptions()

	defer func(args []string) {
		os.Args = args
		ketallOptions = origOpts
	}(os.Args)
	os.Args = []string{"ketall", "unknown"}
	ketallOptions = newOpts

	err := Execute()

	assert.Error(t, err)
}

func TestMainVersionCommand(t *testing.T) {
	origOpts := ketallOptions
	newOpts, _, stdout, stderr := options.NewTestTestCmdOptions()

	defer func(args []string) {
		os.Args = args
		ketallOptions = origOpts
	}(os.Args)
	os.Args = []string{"ketall", "version", "--full"}
	ketallOptions = newOpts

	err := Execute()

	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "ketall:")
	assert.Empty(t, stderr.String())
}
