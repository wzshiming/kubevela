package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/cloud-native-application/rudrx/api/types"

	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/cloud-native-application/rudrx/pkg/utils/system"

	"github.com/crossplane/oam-kubernetes-runtime/apis/core"
	"github.com/spf13/cobra"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	"github.com/cloud-native-application/rudrx/pkg/cmd"
	cmdutil "github.com/cloud-native-application/rudrx/pkg/cmd/util"
	"github.com/cloud-native-application/rudrx/pkg/utils/logs"
)

// noUsageError suppresses usage printing when it occurs
// (since cobra doesn't provide a good way to avoid printing
// out usage in only certain situations).
type noUsageError struct{ error }

var (
	scheme = k8sruntime.NewScheme()

	// VelaVersion is the version of cli.
	VelaVersion = "UNKNOWN"

	// GitRevision is the commit of repo
	GitRevision = "UNKNOWN"
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = core.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	rand.Seed(time.Now().UnixNano())

	command := newCommand()

	logs.InitLogs()
	defer logs.FlushLogs()

	command.Execute()
}

func newCommand() *cobra.Command {
	ioStream := cmdutil.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}

	cmds := &cobra.Command{
		Use:          "vela",
		Short:        "✈️  A Micro App Plafrom for Kubernetes.",
		Long:         "✈️  A Micro App Plafrom for Kubernetes.",
		Run:          runHelp,
		SilenceUsage: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
	restConf, err := config.GetConfig()
	if err != nil {
		fmt.Println("get kubeconfig err", err)
		os.Exit(1)
	}

	commandArgs := types.Args{
		Config: restConf,
		Schema: scheme,
	}

	if err := system.InitApplicationDir(); err != nil {
		fmt.Println("InitApplicationDir err", err)
		os.Exit(1)
	}
	if err := system.InitDefinitionDir(); err != nil {
		fmt.Println("InitDefinitionDir err", err)
		os.Exit(1)
	}

	cmds.AddCommand(
		cmd.NewAdminInitCommand(commandArgs, ioStream),
		cmd.NewAdminInfoCommand(VelaVersion, ioStream),

		cmd.NewTraitsCommand(ioStream),
		cmd.NewWorkloadsCommand(ioStream),
		cmd.NewRefreshCommand(commandArgs, ioStream),

		cmd.NewDeleteCommand(commandArgs, ioStream, os.Args[1:]),
		cmd.NewAppsCommand(commandArgs, ioStream),
		cmd.NewAppStatusCommand(commandArgs, ioStream),

		cmd.NewEnvInitCommand(commandArgs, ioStream),
		cmd.NewEnvSwitchCommand(ioStream),
		cmd.NewEnvDeleteCommand(ioStream),
		cmd.NewEnvCommand(ioStream),

		cmd.NewAddonConfigCommand(ioStream),
		cmd.NewAddonListCommand(commandArgs, ioStream),

		cmd.NewCompletionCommand(),
		NewVersionCommand(),
	)
	if err = cmd.AddWorkloadPlugins(cmds, commandArgs, ioStream); err != nil {
		fmt.Println("Add plugins from workloadDefinition err", err)
		os.Exit(1)
	}
	if err = cmd.AddTraitPlugins(cmds, commandArgs, ioStream); err != nil {
		fmt.Println("Add plugins from traitDefinition err", err)
		os.Exit(1)
	}
	if err = cmd.DetachTraitPlugins(cmds, commandArgs, ioStream); err != nil {
		fmt.Println("Add plugins from traitDefinition err", err)
		os.Exit(1)
	}
	return cmds
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints out build version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf(`Version: %v
GitRevision: %v
GolangVersion: %v
`,
				VelaVersion,
				GitRevision,
				runtime.Version())
		},
	}
}