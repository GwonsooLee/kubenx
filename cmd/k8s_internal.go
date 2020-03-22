package cmd

import (
	"flag"
	"os"
	"strings"
	"path/filepath"
	"encoding/json"
	"strconv"

	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/olekukonko/tablewriter"
)

// Get All pod list
func _get_pod_list()  {
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
		panic(err.Error())
		os.Exit(1)
	}


	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
		os.Exit(1)
	}

	//Get All Pods
	namespace := _get_current_namespace()
	pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
		os.Exit(1)
	}

	//Variable for all pods
	var objectMeta metav1.ObjectMeta
	//var podStatus metav1.PodStatus

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name",  "READY", "STATUS", "Hostname", "Pod IP", "HostIP", "Node", "CreationTimestamp"})

	for _, pod := range pods.Items {
		objectMeta = pod.ObjectMeta
		podStatus := pod.Status
		podSpec := pod.Spec
		byteStatus, _ := json.Marshal(podStatus.Phase)
		status := strings.Replace(string(byteStatus), "\"", "", 2)
		byteCT, _ := json.Marshal(objectMeta.CreationTimestamp)
		creationTimestamp :=strings.Replace(string(byteCT), "\"", "", 2)

		readyCount := 0
		totalCount := 0
		for _, containerStatus := range podStatus.ContainerStatuses {
			totalCount += 1
			if(containerStatus.Ready) {
				readyCount += 1
			}
		}

		table.Append([]string{objectMeta.Name, strconv.Itoa(readyCount)+"/"+strconv.Itoa(totalCount), status, podSpec.Hostname, podStatus.PodIP, podStatus.HostIP , podSpec.NodeName, creationTimestamp })
	}
	table.Render()
}

// Get current namespace
func _get_current_namespace() string {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	namespace, _, _ := clientcmd.ClientConfig.Namespace(kubeConfig)
	return namespace
}