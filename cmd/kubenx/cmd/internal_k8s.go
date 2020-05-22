package cmd

import (
	"k8s.io/client-go/tools/clientcmd/api"
	"os"
	"fmt"
	"time"
	"flag"
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"path/filepath"

	"k8s.io/client-go/rest"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/viper"
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	v1beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
)

//Port Forward Request
type PortForwardAPodRequest struct {
	// RestConfig is the kubernetes config
	RestConfig *rest.Config
	// Pod is the selected pod for this port forwarding
	Pod corev1.Pod
	// LocalPort is the local port that will be selected to expose the PodPort
	LocalPort int
	// PodPort is the target port for the pod
	PodPort int
	// Steams configures where to write or read input from
	Streams genericclioptions.IOStreams
	// StopCh is the channel used to manage the port forward lifecycle
	StopCh <-chan struct{}
	// ReadyCh communicates when the tunnel is ready to receive traffic
	ReadyCh chan struct{}
}

// Get Current Cluster
func GetCurrentCluster() (string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	rawConfig, err := clientcmd.ClientConfig.RawConfig(kubeConfig)
	if err != nil {
		return NO_STRING, err
	}

	return rawConfig.CurrentContext, nil
}

// Get api configuration
func getAPIConfig() (*api.Config, *string, error)  {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	configs, err := clientcmd.LoadFromFile(*kubeconfig)
	if err != nil {
		return nil, kubeconfig, err
	}

	return configs, kubeconfig, nil
}

//Get current configuration
func getCurrentConfig() (*api.Config, error) {
	configOverrides := &clientcmd.ConfigOverrides{}
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	currentConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return nil, err
	}

	return &currentConfig, nil
}

func getConfigFromFlag() (*rest.Config, error) {
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
		return nil, err
	}

	return config, nil
}




// Get current namespace
func _get_namespace() string {
	//Check all Namespace
	allNamespace := viper.GetBool("all")
	if allNamespace == true {
		return ""
	}

	//Check the flag
	ret := viper.GetString("namespace")

	// If no flag is given, then set current namespace
	if len(ret) <= 0 {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		namespace, _, err := clientcmd.ClientConfig.Namespace(kubeConfig)
		if err != nil {
			Red(err.Error())
			os.Exit(1)
		}

		ret = namespace
	}

	return ret
}

// Get Current Cluster
func _get_current_cluster() string {

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	rawConfig, err := clientcmd.ClientConfig.RawConfig(kubeConfig)
	if err != nil {
		Red(err.Error())
		os.Exit(1)
	}

	return rawConfig.CurrentContext
}

// Get K8s Client (Version 1)
func _get_k8s_client() *kubernetes.Clientset {
	//get Configuration
	config := _get_configuration()

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		Red(err.Error())
		os.Exit(1)
	}

	return clientset
}

// Get K8s Client (Version 1)
func _get_k8s_client_with_configuration(config *rest.Config) *kubernetes.Clientset {
	//get Configuration
	if config == nil {
		config = _get_configuration()
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		Red(err.Error())
		os.Exit(1)
	}

	return clientset
}

// Get v1beta Client
func _get_v1beta_client() *v1beta1.ExtensionsV1beta1Client {
	//get Configuration
	config := _get_configuration()

	// create the clientset
	clientset, err := v1beta1.NewForConfig(config)
	if err != nil {
		Red(err.Error())
		os.Exit(1)
	}

	return clientset
}

// Get v1beta Client with configuration
func _get_v1beta_client_with_configuration(config *rest.Config) *v1beta1.ExtensionsV1beta1Client {
	//get Configuration
	if config == nil {
		config = _get_configuration()
	}

	// create the clientset
	clientset, err := v1beta1.NewForConfig(config)
	if err != nil {
		Red(err.Error())
		os.Exit(1)
	}

	return clientset
}

// Get Kubernetes Configuration
func _get_configuration() *rest.Config {
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

	return config
}

// Change current Context
func _change_current_context() {
	var kubeconfig *string
	var contextList []string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	configs, err := clientcmd.LoadFromFile(*kubeconfig)
	if err != nil {
		Red(err.Error())
		os.Exit(1)
	}

	// get list of context
	for context, _ := range configs.Contexts {
		contextList = append(contextList, context)
	}

	// Get Client Configuration
	configOverrides := &clientcmd.ConfigOverrides{}
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	currentConfig, err := kubeConfig.RawConfig()
	if err != nil {
		Red(err.Error())
		os.Exit(1)
	}

	//getting current Context
	currentContext := currentConfig.CurrentContext

	// Get New Context
	newContext := ""
	Red("Current Context: " + currentContext)
	prompt := &survey.Select{
		Message: "Choose Context:",
		Options: contextList,
	}
	survey.AskOne(prompt, &newContext)

	if newContext == "" {
		Red("Changing Context has been canceled")
		os.Exit(1)
	}

	//Change To New Context
	currentConfig.CurrentContext = newContext
	configAccess := clientcmd.NewDefaultClientConfig(*configs, configOverrides).ConfigAccess()

	clientcmd.ModifyConfig(configAccess, currentConfig, false)
	Yellow("Context is changed to \"" + newContext + "\"")
}

// Get ingress List
func _get_ingress_list() {
	//Get kubernetes Client
	clientset := _get_v1beta_client()

	//Get namespace
	namespace := _get_namespace()

	//Get all ingress list in the namespace
	ingresses, err := clientset.Ingresses(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		Red(err.Error())
		os.Exit(1)
	}

	//Tables to show information
	table := _get_table_object()
	table.SetHeader([]string{"Name", "HOST", "ADDRESS", "PATH", "PORTS", "TARGET SERVICE", "AGE"})

	now := time.Now()
	for _, ingress := range ingresses.Items {
		objectMeta := ingress.ObjectMeta
		ingressSpec := ingress.Spec

		var address string
		if len(ingress.Status.LoadBalancer.Ingress) == 0 {
			address = ""
		} else {
			address = ingress.Status.LoadBalancer.Ingress[0].Hostname
		}
		duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

		host := ingressSpec.Rules[0].Host
		if len(host) <= 0 {
			host = "*"
		}

		port := []string{}
		service := []string{}
		paths := []string{}
		for _, path := range ingressSpec.Rules[0].IngressRuleValue.HTTP.Paths {
			if path.Backend.ServicePort.IntVal != 0 {
				port = append(port, _int32_to_string(path.Backend.ServicePort.IntVal))
			}
			service = append(service, path.Backend.ServiceName)
			paths = append(paths, path.Path)
		}
		table.Append([]string{objectMeta.Name, host, address, strings.Join(paths, ","), strings.Join(port, ","), strings.Join(service, ","), duration})
	}
	table.Render()
}

// Get Deployment List
func _get_deployment_list() {
	now := time.Now()
	var depList []string

	// Get Configuration
	config := _get_configuration()

	//Get Clientset
	clientset := _get_v1beta_client_with_configuration(config)
	clientset_v1 := _get_k8s_client_with_configuration(config)

	//Get namespace
	namespace := _get_namespace()

	//Get all deployments list in the namespace
	errorCount := 0

	//Tables to show information
	table := _get_table_object()
	table.SetHeader([]string{"Name", "READY", "UP-TO-DATE", "AVAILABLE", "Strategy Type", "MaxUnavailable", "MaxSurge", "CONTAINERS", "IMAGE", "AGE"})

	// Search Deployment with v1beta1
	deployments, err := clientset.Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		errorCount += 1
	} else {
		//Start Searching
		for _, deployment := range deployments.Items {
			//Get Object Meta Data
			objectMeta := deployment.ObjectMeta
			spec := deployment.Spec

			duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

			// Check if it is using RollingUpdate strategy in order to get MaxUnAvailable, MaxSurge
			maxUnavailable := ""
			maxSurge := ""
			if spec.Strategy.RollingUpdate != nil {
				maxUnavailable = spec.Strategy.RollingUpdate.MaxUnavailable.StrVal
				maxSurge = spec.Strategy.RollingUpdate.MaxSurge.StrVal
			}

			// Get container spec in pod
			podSpec := spec.Template.Spec
			nameString := ""
			imageString := ""
			for _, container := range podSpec.Containers {
				nameString += container.Name + "\n"
				imageString += container.Image + "\n"
			}

			table.Append([]string{objectMeta.Name, _int32_to_string(*spec.Replicas), _int32_to_string(deployment.Status.UpdatedReplicas), _int32_to_string(deployment.Status.AvailableReplicas), string(spec.Strategy.Type), maxUnavailable, maxSurge, nameString, imageString, duration})
			depList = append(depList, objectMeta.Name)
		}
	}

	// Get Deployment with core v1 version
	deploymentsV1, err := clientset_v1.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		errorCount += 1
	} else {
		//Start Searching
		for _, deployment := range deploymentsV1.Items {
			//Get Object Meta Data
			objectMeta := deployment.ObjectMeta

			hasValue := false
			for _, element := range depList {
				if element == objectMeta.Name {
					hasValue = true
					break
				}
			}

			if hasValue { continue}

			spec := deployment.Spec
			duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

			// Check if it is using RollingUpdate strategy in order to get MaxUnAvailable, MaxSurge
			maxUnavailable := ""
			maxSurge := ""
			if spec.Strategy.RollingUpdate != nil {
				maxUnavailable = spec.Strategy.RollingUpdate.MaxUnavailable.StrVal
				maxSurge = spec.Strategy.RollingUpdate.MaxSurge.StrVal
			}

			// Get container spec in pod
			podSpec := spec.Template.Spec
			nameString := ""
			imageString := ""
			for _, container := range podSpec.Containers {
				nameString += container.Name + "\n"
				imageString += container.Image + "\n"
			}

			table.Append([]string{objectMeta.Name, _int32_to_string(*spec.Replicas), _int32_to_string(deployment.Status.UpdatedReplicas), _int32_to_string(deployment.Status.AvailableReplicas), string(spec.Strategy.Type), maxUnavailable, maxSurge, nameString, imageString, duration})
		}
	}

	// If Error count is larger than 1, which means no deployment int the cluster
	if errorCount > 1 {
		Red("The server could not find the requested resource")
		os.Exit(1)
	}
	table.Render()
}

// Get All Raw Pod list
func _get_all_raw_pods(clientset *kubernetes.Clientset, namespace string, labelSelector string) []corev1.Pod {
	listOpt := metav1.ListOptions{}
	if len(labelSelector) > 0 {
		listOpt = metav1.ListOptions{LabelSelector: labelSelector}
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), listOpt)
	if err != nil {
		Red(err.Error())
		os.Exit(1)
	}

	return pods.Items
}

//Get All raw node list
func _get_all_raw_node(clientset *kubernetes.Clientset, labelSelector string) []corev1.Node {
	listOpt := metav1.ListOptions{}
	if len(labelSelector) > 0 {
		listOpt = metav1.ListOptions{LabelSelector: labelSelector}
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), listOpt)
	if err != nil {
		Red(err.Error())
		os.Exit(1)
	}

	return nodes.Items
}



//Get All node list
func _get_node_list() {
	//Get kubernetes Client
	clientset := _get_k8s_client()

	nodes := _get_all_raw_node(clientset, NO_STRING)
	_render_node_list_info(nodes)
}

//Retrive only node List for ssh
func _get_node_list_for_option(clientset *kubernetes.Clientset) []string {
	//Get kubernetes Client
	if clientset == nil {
		clientset = _get_k8s_client()
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		Red(err.Error())
		os.Exit(1)
	}

	nodeList := []string{}
	for _, node := range nodes.Items {
		nodeStatus := node.Status
		objectMeta := node.ObjectMeta
		for _, nodeAddr := range nodeStatus.Addresses {
			if nodeAddr.Type == "Hostname" {
				labels := _create_label_for_option(objectMeta.Labels)
				nodeList = append(nodeList, fmt.Sprintf("%s (%s)", nodeAddr.Address,labels))
				break
			}

		}
	}

	return nodeList
}

func _create_label_for_option(labels map[string]string) string  {
	ret := []string{}
	LabelFilters := DEFAULT_NODE_LABEL_FILTERS


	for _, key := range LabelFilters {
		if len(labels[key]) > 0 {
			ret = append(ret, key+"="+labels[key])
		}
	}

	if len(ret) == 0 {
		return "No Labels for filtering"
	}

	return strings.Join(ret, ",")
}

// Get All Service list
func _get_service_list() {
	//Get kubernetes Client
	clientset := _get_k8s_client()

	//Get namespace
	namespace := _get_namespace()

	//Get All Pods
	services, err := clientset.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		Red(err.Error())
		os.Exit(1)
	}

	//Variable for all pods
	var objectMeta metav1.ObjectMeta

	table := _get_table_object()
	table.SetHeader([]string{"NAME", "TYPE", "CLUSTER-IP", "EXTERNAL-IP", "PORT(S)", "ENDPOINT(S)", "AGE"})

	//Get detailed information about Service
	now := time.Now()
	for _, service := range services.Items {
		objectMeta = service.ObjectMeta

		duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

		serviceSpec := service.Spec
		serviceType := string(serviceSpec.Type)
		ExternalIPs := serviceSpec.ExternalIPs

		// Get Load Balancer Name for this
		externalIP := ""
		if serviceType == "LoadBalancer" {
			externalIP = service.Status.LoadBalancer.Ingress[0].Hostname
		} else if len(ExternalIPs) == 0 {
			externalIP = "<None>"
		}

		// Get Listening Ports of Service
		portString := ""
		for _, port := range serviceSpec.Ports {
			appPort := strconv.FormatInt(int64(port.Port), 10)
			portString += appPort

			if port.NodePort != 0 {
				portString += (":" + _int32_to_string(port.NodePort) + "/TCP, ")
			} else {
				portString += "/TCP, "
			}
		}
		portString = portString[:len(portString)-2]

		// Convert Label Selector to string set
		labelSelector := []string{}
		if len(serviceSpec.Selector) > 0 {
			for key, value := range serviceSpec.Selector {
				labelSelector = append(labelSelector, key + "=" + value)
			}
		}

		//Get All Pods with endpoint
		pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: strings.Join(labelSelector,",")})
		if err != nil {
			Red(err.Error())
			os.Exit(1)
		}

		//Get Endpoints (Pods)
		endpoints := []string{}
		if len(pods.Items) > 0 {
			for _, pod := range pods.Items {
				podIP := pod.Status.PodIP
				endpoints = append(endpoints, podIP)
			}
		}

		table.Append([]string{objectMeta.Name, serviceType, serviceSpec.ClusterIP, externalIP, portString, strings.Join(endpoints, ","), duration})
	}
	table.Render()
}

// Select Pod and port before port forward
func selectPodPortNS(options []string) (string, int, int) {
	// Choose Pod from the list
	var pod, local_port, pod_port string

	// Choose Pod from the list
	prompt := &survey.Select{
		Message: "Choose a pod:",
		Options: options,
	}
	survey.AskOne(prompt, &pod)

	if pod == "" {
		Red("You canceled the choice")
		os.Exit(1)
	}

	// Choose local port first
	prompt_input := &survey.Input{
		Message: "Local port to use:",
	}
	survey.AskOne(prompt_input, &local_port)

	if local_port == "" {
		Red("You canceled the choice")
		os.Exit(1)
	}

	// Choose local port first
	pod_port = local_port
	prompt_input2 := &survey.Input{
		Message: "Pod port[ Default: " + local_port + "]:",
		Default: local_port,
	}
	err := survey.AskOne(prompt_input2, &pod_port)
	if err == terminal.InterruptErr {
		Red("interrupted")
		os.Exit(1)
	}

	return pod, _string_to_int(local_port), _string_to_int(pod_port)
}

// Get PortForward Dialer
func _port_forward_to_pod(req PortForwardAPodRequest) error {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		req.Pod.Namespace, req.Pod.Name)
	hostIP := strings.TrimLeft(req.RestConfig.Host, "htps:/")

	transport, upgrader, err := spdy.RoundTripperFor(req.RestConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", req.LocalPort, req.PodPort)}, req.StopCh, req.ReadyCh, req.Streams.Out, req.Streams.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

// Get Node for inspect
func _get_target_node(clientset *kubernetes.Clientset, args []string) string {
	// Pass from command
	if len(args) == 1 {
		return args[0]
	}

	options := _get_node_list_for_option(clientset)

	if len(options) == 0 {
		Red("No available node exists")
		os.Exit(1)
	}

	var node string
	if len(options) > 0 {
		prompt := &survey.Select{
			Message: "Choose a node:",
			Options: options,
		}
		survey.AskOne(prompt, &node)
	}

	return node
}

//Inspect node in detail
func _inspect_node(args []string)  {
	//Get kubernetes Client
	clientset := _get_k8s_client()

	//namespace
	namespace := viper.GetString("namespace")

	//get target node
	target := _get_target_node(clientset, args)

	// Get node information
	detail, err := clientset.CoreV1().Nodes().Get(context.Background(), target, metav1.GetOptions{})
	if err != nil {
		Red(err.Error())
		os.Exit(1)
	}

	taints := detail.Spec.Taints

	Yellow("========Taint INFO=======")
	for _, taint := range taints {
		txt := fmt.Sprintf("%s=%s:%s", taint.Key, taint.Value, taint.Effect)
		Blue(txt)
	}

	if len(taints) == 0 {
		Red("There is no taints applied")
	}

	//Get all pods
	pods := _get_all_raw_pods(clientset, namespace, NO_STRING)

	filtered := []corev1.Pod{}
	for _, pod := range pods {
		if pod.Spec.NodeName == target {
			filtered = append(filtered, pod)
		}
	}

	fmt.Println()
	Yellow("========POD INFO=======")
	_render_pod_list_info(filtered)
}

// Render Pod list
func _render_pod_list_info(pods []corev1.Pod)  {
	if len(pods) <= 0 {
		Red("No pod exists in the namespace")
		return
	}
	// Table setup
	table := _get_table_object()
	table.SetHeader([]string{"Name", "READY", "STATUS", "Hostname", "Pod IP", "Host IP", "Node", "Age"})

	now := time.Now()
	for _, pod := range pods {
		objectMeta := pod.ObjectMeta
		podStatus := pod.Status
		podSpec := pod.Spec
		duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

		readyCount := 0
		totalCount := 0
		var status string

		status = fmt.Sprintf("%v", podStatus.Phase)
		for _, containerStatus := range podStatus.ContainerStatuses {
			totalCount += 1
			if containerStatus.Ready {
				readyCount += 1
			}

			if containerStatus.State.Waiting != nil {
				status = containerStatus.State.Waiting.Reason
				break
			}

			if containerStatus.State.Running != nil {
				status = "Running"
			}

			if containerStatus.State.Terminated != nil {
				status = fmt.Sprintf("%s (Exit Code: %d", containerStatus.State.Terminated.Reason, containerStatus.State.Terminated.ExitCode)
			}
		}

		table.Append([]string{objectMeta.Name, strconv.Itoa(readyCount) + "/" + strconv.Itoa(totalCount), status, podSpec.Hostname, podStatus.PodIP, podStatus.HostIP, podSpec.NodeName, duration})
	}
	table.Render()
}


// Render Pod list
func _render_node_list_info(nodes []corev1.Node)  {
	if len(nodes) <= 0 {
		Red("No node exists in the namespace")
		return
	}
	//Variable for all pods
	var objectMeta metav1.ObjectMeta
	var externalIp, internalIp string
	now := time.Now()

	// Table setup
	table := _get_table_object()
	table.SetHeader([]string{"NAME", "STATUS", "INTERNAL-IP", "EXTERNAL-IP", "LABEL", "OS-IMAGE", "AGE"})

	//Get detailed information about Service
	labelFilters := DEFAULT_NODE_LABEL_FILTERS
	for _, node := range nodes {
		objectMeta = node.ObjectMeta

		duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

		labels := []string{}
		for _, key := range labelFilters {
			if len(objectMeta.Labels[key]) > 0 {
				labels = append(labels, key+"="+objectMeta.Labels[key])
			}
		}

		nodeStatus := node.Status
		for _, nodeAddr := range nodeStatus.Addresses {
			if nodeAddr.Type == "InternalIP" {
				internalIp = nodeAddr.Address
			}

			if nodeAddr.Type == "ExternalIP" {
				externalIp = nodeAddr.Address
			}
		}

		status := ""
		for _, condition := range nodeStatus.Conditions {
			if condition.Status == "True" {
				status = fmt.Sprint(condition.Type)
			}
		}

		table.Append([]string{objectMeta.Name, status, internalIp, externalIp, strings.Join(labels,","), nodeStatus.NodeInfo.OSImage, duration})
	}
	table.Render()
}

func _make_label_selector() string {
	key := _get_single_string_input("Key", "Input for key has been cancelled")
	value := _get_single_string_input("Value", "Input for value has been cancelled")

	return fmt.Sprintf("%s=%s", key, value)
}
