/*
Copyright © 2019 Portworx

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
package context

import (
	pxcmd "github.com/portworx/pxc/cmd"
	"github.com/portworx/pxc/pkg/commander"
	"github.com/portworx/pxc/pkg/contextconfig"
	"github.com/spf13/cobra"
)

var contextUnsetCmd *cobra.Command

var _ = commander.RegisterCommandVar(func() {
	contextUnsetCmd = &cobra.Command{
		Use:   "unset",
		Short: "Unset the current context configuration",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return contextUnsetExec(cmd, args)
		},
	}
})

var _ = commander.RegisterCommandInit(func() {
	pxcmd.ContextAddCommand(contextUnsetCmd)
})

func contextUnsetExec(cmd *cobra.Command, args []string) error {
	contextManager, err := contextconfig.NewContextManager(pxcmd.GetConfigFile())
	if err != nil {
		return err
	}

	if err := contextManager.SetCurrent(""); err != nil {
		return err
	}
	return nil
}
