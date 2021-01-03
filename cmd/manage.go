package cmd

import (
	"github.com/nicjohnson145/godot/internal/builder"
	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(manageCmd)
	manageCmd.Flags().StringVar(&as, "as", "", "Override the generated template name, useful for resolving collisions")
	manageCmd.Flags().BoolVar(&add, "add", false, "Add the imported file to the current target")
}

var (
	as  string
	add bool

	manageCmd = &cobra.Command{
		Use:   "manage <file>",
		Short: "Add file to be managed by godot",
		Long:  "Import/create a file to be managed by godot at the specified location",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			getter := &util.OSHomeDir{}
			return importFile(getter, args[0], as, add)
		},
	}
)

func importFile(getter util.HomeDirGetter, file string, as string, add bool) error {
	conf := config.NewConfig(getter)
	builder := builder.Builder{Config: conf, Getter: getter}

	err := builder.Import(file, as)
	if err != nil {
		return err
	}
	var name string
	if as != "" {
		name, err = conf.AddFile(as, file)
	} else {
		name, err = conf.ManageFile(file)
	}

	if add {
		err = conf.AddToTarget(conf.Target, name)
	}

	if err == nil {
		err = conf.Write()
	}

	return err
}
