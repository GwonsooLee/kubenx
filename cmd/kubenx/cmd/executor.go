package cmd

import (
	"context"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"flag"
)
type Executor struct {
	Client *kubernetes.Clientset
}

// Run executor for command line
func runExecutor(ctx context.Context, action func(Executor) error) error {
	executor, err:= createNewExecutor()
	if err != nil {
		return err
	}

	err = action(executor)

	return alwaysSucceedWhenCancelled(ctx, err)
}

// Create new executor
func createNewExecutor() (Executor, error) {
	//Get kubernetes Client
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		Red(err.Error())
		os.Exit(1)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		Red(err.Error())
		os.Exit(1)
	}

	return Executor{
		Client: clientset,
	}, err

}
