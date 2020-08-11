package cmd

import (
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/GwonsooLee/kubenx/pkg/color"
	"github.com/GwonsooLee/kubenx/pkg/runner"
	"github.com/GwonsooLee/kubenx/pkg/utils"
	"github.com/spf13/cobra"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

//Create Command for get pod
func NewCmdNamespace() *cobra.Command {
	return NewCmd("namespace").
		WithDescription("Change namespace").
		SetAliases([]string{"ns"}).
		RunWithArgs(execNamespace)
}

// Function for get command
func execNamespace(ctx context.Context, out io.Writer, args []string) error {
	return runWithoutExecutor(ctx, func() error {
		var namespaceList []string

		//Get API configuration
		configs, kubeconfig, err := runner.GetAPIConfig()
		if err != nil {
			color.Red.Fprintln(out, err.Error())
			return err
		}

		// Get Client Configuration
		currentConfig, err := runner.GetCurrentConfig()
		if err != nil {
			color.Red.Fprintln(out, err.Error())
			return err
		}

		// use the current context in kubeconfig
		config, err := clientcmd.BuildConfigFromFlags(utils.NO_STRING, *kubeconfig)
		if err != nil {
			color.Red.Fprintln(out, err.Error())
			return err
		}

		// create the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			color.Red.Fprintln(out, err.Error())
			return err
		}

		// Get All Namespace in current context
		namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			color.Red.Fprintln(out, err.Error())
			return err
		}

		for _, obj := range namespaces.Items {
			namespaceList = append(namespaceList, obj.ObjectMeta.Name)
		}

		newNamespace := ""
		if len(args) == 0 {
			//getting current Context
			currentContext := currentConfig.CurrentContext
			currentNamespace := currentConfig.Contexts[currentConfig.CurrentContext].Namespace

			// Get New Context
			utils.Red("[ " + currentContext + " ] Current Namespace: " + currentNamespace)
			prompt := &survey.Select{
				Message: "Choose Context:",
				Options: namespaceList,
			}
			survey.AskOne(prompt, &newNamespace)

			if newNamespace == "" {
				color.Red.Fprintln(out, fmt.Errorf("No namespace is selected..."))
				return err
			}

		} else {
			target := args[0]

			// Check whether the target exists in current context
			containsTarget := false
			for _, namespace := range namespaceList {
				if target == namespace {
					containsTarget = true
					break
				}
			}

			// If the target is not in the context
			if !containsTarget {
				color.Yellow.Fprintf(out, "[ %s ] namespace doesn't exist. Please check the namespaces.", target)
				return nil
			}

			newNamespace = target
		}

		//Change To New Namespace
		currentConfig.Contexts[currentConfig.CurrentContext].Namespace = newNamespace
		configAccess := clientcmd.NewDefaultClientConfig(*configs, &clientcmd.ConfigOverrides{}).ConfigAccess()

		clientcmd.ModifyConfig(configAccess, *currentConfig, false)
		color.Yellow.Fprintf(out, "Namespace is changed to %s", newNamespace)
		return nil
	})
}
