// Copyright © 2019 Portworx
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

package logs

import (
	"fmt"

	pxcmd "github.com/portworx/pxc/cmd"
	"github.com/portworx/pxc/pkg/cliops"
	"github.com/portworx/pxc/pkg/commander"
	"github.com/portworx/pxc/pkg/kubernetes"
	"github.com/portworx/pxc/pkg/util"
	"github.com/spf13/cobra"
)

var logsNodeCmd *cobra.Command

var _ = commander.RegisterCommandVar(func() {
	logsNodeCmd = &cobra.Command{
		Use:   "node",
		Short: "Print Portworx logs for specified nodes",
		Example: `
        $ px logs node --all-nodes
        Return Portworx logs from all nodes

        $ px logs node abc
        Return Portworx logs from  node abc

        $ px logs node -f  abc
        Begin streaming the Portworx logs from  node abc

        $ px logs node --tail=20 abc
        Apply filters to only the most recent 20 log lines and display the matched lines

        $ px logs node abc --filter "error,warning"
        Display all log lines that has either error or warning on node abc

        $ px logs node --since=1h node
        Show all Portworx logs from node abc written in the last hour`,
		RunE: logsNodesExec,
	}
})

// logsCmd represents the logs command
var _ = commander.RegisterCommandInit(func() {
	pxcmd.LogsAddCommand(logsNodeCmd)
	cliops.AddCommonLogOptions(logsNodeCmd)
	logsNodeCmd.Flags().Bool("all-nodes", false, "If specified, logs from all nodes will be displayed")
})

func NodeAddCommand(cmd *cobra.Command) {
	logsNodeCmd.AddCommand(cmd)
}

func getNodeLogOptions(
	cmd *cobra.Command,
	args []string,
	cvOps *cliops.CliVolumeOps,
) (*kubernetes.COpsLogOptions, error) {
	allNodes, _ := cmd.Flags().GetBool("all-nodes")
	if (allNodes == false && len(args) == 0) ||
		(allNodes == true && len(args) > 0) {
		return nil, fmt.Errorf("Either specify the nodes or --all-nodes")
	}

	lo, err := cliops.GetCommonLogOptions(cmd)
	if err != nil {
		return nil, err
	}
	p, err := cliops.GetRequiredPortworxPods(cvOps, args, lo.PortworxNamespace)
	if err != nil {
		return nil, err
	}
	lo.CInfo = p
	return lo, nil
}

func logsNodesExec(cmd *cobra.Command, args []string) error {
	cvi := &cliops.CliVolumeInputs{
		ShowK8s: true,
	}
	cvi.GetNamespace(cmd)

	// Create a cliVolumeOps object
	cvOps := cliops.NewCliVolumeOps(cvi)

	// Connect to px and k8s (if needed)
	err := cvOps.Connect()
	if err != nil {
		return err
	}
	defer cvOps.Close()

	lo, err := getNodeLogOptions(cmd, args, cvOps)
	if err != nil {
		return err
	}

	return cvOps.PxVolumeOps.GetCOps().GetLogs(lo, util.Stdout)
}
