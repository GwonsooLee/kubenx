package cmd

import (
	"context"
	"github.com/GwonsooLee/kubenx/pkg/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/iam"
	"k8s.io/client-go/kubernetes"
	v1beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	rbacv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
	"k8s.io/client-go/rest"
)

type Executor struct {
	Client 			*kubernetes.Clientset
	BetaV1Client 	*v1beta1.ExtensionsV1beta1Client
	RbacV1Client 	*rbacv1.RbacV1Client
	EKS 			*eks.EKS
	EC2 			*ec2.EC2
	IAM 			*iam.IAM
	Config 			*rest.Config
	Namespace 		string
	Context   		context.Context
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

	config, err := getConfigFromFlag()

	executor.Config = config
	if err != nil {
		return executor, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return executor, err
	}

	executor.Client = clientset

	// create the v1beta1 clientset
	betav1clientset, err := v1beta1.NewForConfig(config)
	if err != nil {
		return executor, err
	}

	executor.BetaV1Client = betav1clientset


	// create the rbac client
	rbacv1clientset, err := rbacv1.NewForConfig(config)
	if err != nil {
		return executor, err
	}

	executor.RbacV1Client = rbacv1clientset

	//Get Namespace
	namespace, err := getNamespace()
	if err != nil {
		return executor, err
	}

	executor.Namespace = namespace

	return executor, err
}

// Run function without executor
func runWithoutExecutor(ctx context.Context, action func() error) error {
	err := action()

	return alwaysSucceedWhenCancelled(ctx, err)
}
