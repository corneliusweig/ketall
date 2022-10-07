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

package options

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type KetallOptions struct {
	GenericCliFlags *genericclioptions.ConfigFlags
	UseCache        bool     `json:"useCache,omitempty"`
	AllowIncomplete bool     `json:"allowIncomplete,omitempty"`
	Scope           string   `json:"scope,omitempty"`
	Since           string   `json:"since,omitempty"`
	Selector        string   `json:"selector,omitempty"`
	FieldSelector   string   `json:"fieldSelector,omitempty"`
	MaxInflight     int64    `json:"maxInflight,omitempty"`
	Namespace       string   `json:"namespace.omitempty"`
	Exclusions      []string `json:"exclusions,omitempty"`
}

func NewDefaultCmdOptions() *KetallOptions {
	return &KetallOptions{
		GenericCliFlags: genericclioptions.NewConfigFlags(true),
	}
}

func GetGenricCliFlags() *genericclioptions.ConfigFlags {
	return genericclioptions.NewConfigFlags(true)
}
