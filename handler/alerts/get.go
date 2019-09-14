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

package alerts

import (
	"bytes"
	"unsafe"

	"github.com/cheynewallace/tabby"
	"github.com/portworx/pxc/cmd"
	"github.com/portworx/pxc/pkg/cliops"
	"github.com/portworx/pxc/pkg/commander"
	"github.com/portworx/pxc/pkg/portworx"
	"github.com/portworx/pxc/pkg/util"
	"github.com/spf13/cobra"
	"text/tabwriter"

	api "github.com/libopenstorage/openstorage-sdk-clients/sdk/golang"
	prototime "github.com/portworx/pxc/pkg/openstorage/proto/time"
)

var getAlertsCmd *cobra.Command

var _ = commander.RegisterCommandVar(func() {
	getAlertsCmd = &cobra.Command{
		Use:     "alerts",
		Aliases: []string{"alerts"},
		Short:   "Get information about Portworx alerts",
		RunE:    getAlertsExec,
	}
})

var _ = commander.RegisterCommandInit(func() {
	cmd.GetAddCommand(getAlertsCmd)
	getAlertsCmd.Flags().StringP("type", "t", "all", "Fetch alerts from Portworx cluster")
	getAlertsCmd.Flags().StringP("output", "o", "", "Output in yaml|json|wide")
	//TODO: Need to support more flags
})

func GetAddCommand(cmd *cobra.Command) {
	getAlertsCmd.AddCommand(cmd)
}

func getAlertsExec(cmd *cobra.Command, args []string) error {
	ctx, conn, err := portworx.PxConnectDefault()
	_ = ctx
	if err != nil {
		return err
	}
	defer conn.Close()
	// Parse out all of the common cli volume flags
	cai := cliops.GetCliAlertInputs(cmd, args)

	// Create a cliVolumeOps object
	alertOps := cliops.NewCliAlertOps(cai)

	// initialize alertOP interface
	alertOps.PxAlertOps = portworx.NewPxAlertOps()

	// Create the parser object
	alertgf := NewAlertGetFormatter(alertOps)
	return util.PrintFormatted(alertgf)
}

type alertGetFormatter struct {
	cliops.CliAlertOps
}

func NewAlertGetFormatter(cvOps *cliops.CliAlertOps) *alertGetFormatter {
	return &alertGetFormatter{
		CliAlertOps: *cvOps,
	}
}

// YamlFormat returns the yaml representation of the object
func (p *alertGetFormatter) YamlFormat() (string, error) {
	alerts, err := p.PxAlertOps.GetPxAlerts(p.CliAlertInputs.AlertType)
	if err != nil {
		return "", err
	}
	return util.ToYaml(alerts.AlertResp)
}

// JsonFormat returns the json representation of the object
func (p *alertGetFormatter) JsonFormat() (string, error) {
	alerts, err := p.PxAlertOps.GetPxAlerts(p.CliAlertInputs.AlertType)
	if err != nil {
		return "", err
	}
	return util.ToJson(alerts.AlertResp)
}

// WideFormat returns the wide string representation of the object
func (p *alertGetFormatter) WideFormat() (string, error) {
	p.Wide = true
	return p.toTabbed()
}

// DefaultFormat returns the default string representation of the object
func (p *alertGetFormatter) DefaultFormat() (string, error) {
	return p.toTabbed()
}

func (p *alertGetFormatter) toTabbed() (string, error) {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	t := tabby.NewCustom(writer)

	alerts, err := p.PxAlertOps.GetPxAlerts(p.CliAlertInputs.AlertType)
	if err != nil {
		return "", err
	}

	if unsafe.Sizeof(alerts) == 0 {
		util.Printf("No alerts found\n")
		return "", nil
	}

	// Start the columns
	t.AddHeader(p.getHeader()...)
	for _, n := range alerts.AlertResp {
		l, err := p.getLine(n, alerts.AlertIdToName[n.GetAlertType()])
		if err != nil {
			return "", nil
		}
		t.AddLine(l...)
	}
	t.Print()
	return b.String(), nil
}

func (p *alertGetFormatter) getHeader() []interface{} {
	var header []interface{}
	if p.Wide {
		header = []interface{}{"Type", "Id", "Resource", "Severity", "Count", "LastSeen", "FirstSeen", "Description"}
	} else {
		header = []interface{}{"Id", "Severity", "Count", "LastSeen", "FirstSeen", "Description"}
	}

	return header
}

func (p *alertGetFormatter) getLine(resp *api.Alert, name string) ([]interface{}, error) {
	var line []interface{}

	if p.Wide {
		line = []interface{}{
			portworx.GetResourceTypeString(resp.GetResource()), name, resp.GetResourceId(),
			portworx.SeverityString(resp.GetSeverity()), resp.GetCount(),
			prototime.TimestampToTime(resp.GetTimestamp()).Format(util.TimeFormat),
			prototime.TimestampToTime(resp.GetFirstSeen()).Format(util.TimeFormat), resp.GetMessage(),
		}
	} else {
		line = []interface{}{
			name, portworx.SeverityString(resp.GetSeverity()), resp.GetCount(),
			prototime.TimestampToTime(resp.GetTimestamp()).Format(util.TimeFormat),
			prototime.TimestampToTime(resp.GetFirstSeen()).Format(util.TimeFormat), resp.GetMessage(),
		}
	}

	return line, nil
}
