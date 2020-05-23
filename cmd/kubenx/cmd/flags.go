package cmd

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"reflect"
)

// This part of code comes from skaffold opensource.
// https://github.com/GoogleContainerTools/skaffold
// Flag defines a Skaffold CLI flag which contains a list of
// subcommands the flag belongs to in `DefinedOn` field.
type Flag struct {
	Name               string
	Shorthand          string
	Usage              string
	Value              interface{}
	DefValue           interface{}
	DefValuePerCommand map[string]interface{}
	FlagAddMethod      string
	DefinedOn          []string
	Hidden             bool

	pflag *pflag.Flag
}

// FlagRegistry is a list of all Skaffold CLI flags.
// When adding a new flag to the registry, please specify the
// command/commands to which the flag belongs in `DefinedOn` field.
// If the flag is a global flag, or belongs to all the subcommands,
/// specify "all"
// FlagAddMethod is method which defines a flag value with specified
// name, default value, and usage string. e.g. `StringVar`, `BoolVar`
var FlagRegistry = []Flag{
	{
		Name:          "namespace",
		Shorthand:     "n",
		Usage:         "Run deployments in the specified namespace",
		Value:         aws.String(NO_STRING),
		DefValue:      "",
		FlagAddMethod: "StringVar",
		DefinedOn:     []string{"pod", "deployment", "service", "serviceaccount", "configmap", "ingress", "role", "rolebinding", "secret" },
	},
	{
		Name:          "region",
		Shorthand:     "r",
		Usage:         "Run command to specific region",
		Value:         aws.String(NO_STRING),
		DefValue:      "ap-northeast-2",
		FlagAddMethod: "StringVar",
		DefinedOn:     []string{"pod", "deployment", "service", "serviceaccount", "configmap", "ingress", "cluster", "init"},
	},
	{
		Name:          "all",
		Shorthand:     "A",
		Usage:         "All namespace flag",
		Value:         aws.Bool(false),
		DefValue:      false,
		FlagAddMethod: "BoolVar",
		DefinedOn:     []string{"pod", "deployment", "service", "serviceaccount", "configmap", "ingress", "role", "clusterrole", "rolebinding", "clusterrolebinding", "secret"},
	},
}

func (fl *Flag) flag() *pflag.Flag {
	if fl.pflag != nil {
		return fl.pflag
	}

	inputs := []interface{}{fl.Value, fl.Name}
	if fl.FlagAddMethod != "Var" {
		inputs = append(inputs, fl.DefValue)
	}
	inputs = append(inputs, fl.Usage)

	fs := pflag.NewFlagSet(fl.Name, pflag.ContinueOnError)
	reflect.ValueOf(fs).MethodByName(fl.FlagAddMethod).Call(reflectValueOf(inputs))
	f := fs.Lookup(fl.Name)
	f.Shorthand = fl.Shorthand
	f.Hidden = fl.Hidden

	fl.pflag = f
	return f
}

func reflectValueOf(values []interface{}) []reflect.Value {
	var results []reflect.Value
	for _, v := range values {
		results = append(results, reflect.ValueOf(v))
	}
	return results
}

//Add command flags
func SetCommandFlags(cmd *cobra.Command)  {
	for _, child := range cmd.Commands() {
		var flagsForCommand []*Flag
		for i := range FlagRegistry {
			fl := &FlagRegistry[i]

			if isStringInArr(child.Use, fl.DefinedOn){
				child.PersistentFlags().AddFlag(fl.flag())
				flagsForCommand = append(flagsForCommand, fl)
			}
		}

		// Apply command-specific default values to flags.
		child.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			// Update default values.
			for _, fl := range flagsForCommand {
				viper.BindPFlag(fl.Name, cmd.PersistentFlags().Lookup(fl.Name))
			}

			// Since PersistentPreRunE replaces the parent's PersistentPreRunE,
			// make sure we call it, if it is set.
			if parent := cmd.Parent(); parent != nil {
				if preRun := parent.PersistentPreRunE; preRun != nil {
					if err := preRun(cmd, args); err != nil {
						return err
					}
				} else if preRun := parent.PersistentPreRun; preRun != nil {
					preRun(cmd, args)
				}
			}

			return nil
		}
	}
}

