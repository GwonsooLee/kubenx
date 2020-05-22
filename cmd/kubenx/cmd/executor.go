package cmd

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"os"
	"flag"
	"context"
	"path/filepath"
	"k8s.io/client-go/rest"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/GwonsooLee/kubenx/pkg/aws"
)
type Executor struct {
	Client 		*kubernetes.Clientset
	EKS 		*eks.EKS
	EC2 		*ec2.EC2
	IAM 		*iam.IAM
	Config 		*rest.Config
	Namespace 	string
	Context   	context.Context
}

// Run executor for command line
func runExecutor(ctx context.Context, action func(Executor) error) error {
	executor, err:= createNewExecutor()
	if err != nil {
		return err
	}

	//Run function with executor
	err = action(executor)

	return alwaysSucceedWhenCancelled(ctx, err)
}

// Run executor for command line
func runExecutorWithAWS(ctx context.Context, action func(Executor) error) error {
	executor, err:= createNewExecutor()
	if err != nil {
		return err
	}

	//Set AWS sessions
	executor.EKS = aws.GetEksSession()
	executor.EC2 = aws.GetEC2Session()
	executor.IAM = aws.GetIAMSession()

	//Run function with executor
	err = action(executor)

	return alwaysSucceedWhenCancelled(ctx, err)
}

// Create new executor
func createNewExecutor() (Executor, error) {
	executor := Executor{}
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

	executor.Config = config

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		Red(err.Error())
		os.Exit(1)
	}

	executor.Client = clientset

	//Check the flag
	setAll := viper.GetBool("all")
	namespace := viper.GetString("namespace")

	// If no flag is given, then set current namespace
	if len(namespace) <= 0 {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		namespace, _, err = clientcmd.ClientConfig.Namespace(kubeConfig)
		if err != nil {
			Red(err.Error())
			os.Exit(1)
		}
	}

	if setAll && namespace != NO_STRING {
		namespace = NO_STRING
	}

	executor.Namespace = namespace

	return executor, err
}


// Run function without executor
func runWithoutExecutor(ctx context.Context, action func() error) error {
	err := action()

	return alwaysSucceedWhenCancelled(ctx, err)
}
