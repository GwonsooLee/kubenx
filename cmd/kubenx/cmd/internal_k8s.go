package cmd

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	typedRbacv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/GwonsooLee/kubenx/pkg/table"
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
// Get All Raw Pod list
func getAllRawPods(ctx context.Context, clientset *kubernetes.Clientset, namespace string, labelSelector string) ([]corev1.Pod, error) {
	listOpt := metav1.ListOptions{}
	if len(labelSelector) > 0 {
		listOpt = metav1.ListOptions{LabelSelector: labelSelector}
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, listOpt)
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

// Get All Raw configmap list
func getAllRawConfigMaps(ctx context.Context, clientset *kubernetes.Clientset, namespace string, labelSelector string) ([]corev1.ConfigMap, error) {
	listOpt := metav1.ListOptions{}
	if len(labelSelector) > 0 {
		listOpt = metav1.ListOptions{LabelSelector: labelSelector}
	}

	configMaps, err := clientset.CoreV1().ConfigMaps(namespace).List(ctx, listOpt)
	if err != nil {
		return nil, err
	}

	return configMaps.Items, nil
}

// Get All Raw secret list
func getAllRawSecrets(ctx context.Context, clientset *kubernetes.Clientset, namespace string, labelSelector string) ([]corev1.Secret, error) {
	listOpt := metav1.ListOptions{}
	if len(labelSelector) > 0 {
		listOpt = metav1.ListOptions{LabelSelector: labelSelector}
	}

	secrets, err := clientset.CoreV1().Secrets(namespace).List(ctx, listOpt)
	if err != nil {
		return nil, err
	}

	return secrets.Items, nil
}


// Get All Raw clusterrole list
func getAllRawClusterRoles(ctx context.Context, clientset *typedRbacv1.RbacV1Client, labelSelector string) ([]rbacv1.ClusterRole, error) {
	listOpt := metav1.ListOptions{}
	if len(labelSelector) > 0 {
		listOpt = metav1.ListOptions{LabelSelector: labelSelector}
	}

	clusterRoles, err := clientset.ClusterRoles().List(ctx, listOpt)
	if err != nil {
		return nil, err
	}

	return clusterRoles.Items, nil
}

// Get All Raw cluster role binding list
func getAllRawClusterRoleBindings(ctx context.Context, clientset *typedRbacv1.RbacV1Client, labelSelector string) ([]rbacv1.ClusterRoleBinding, error) {
	listOpt := metav1.ListOptions{}
	if len(labelSelector) > 0 {
		listOpt = metav1.ListOptions{LabelSelector: labelSelector}
	}

	clusterRoleBindings, err := clientset.ClusterRoleBindings().List(ctx, listOpt)
	if err != nil {
		return nil, err
	}

	return clusterRoleBindings.Items, nil
}


// Get All Raw role list
func getAllRawRoles(ctx context.Context, clientset *typedRbacv1.RbacV1Client, namespace string, labelSelector string) ([]rbacv1.Role, error) {
	listOpt := metav1.ListOptions{}
	if len(labelSelector) > 0 {
		listOpt = metav1.ListOptions{LabelSelector: labelSelector}
	}

	roles, err := clientset.Roles(namespace).List(ctx, listOpt)
	if err != nil {
		return nil, err
	}

	return roles.Items, nil
}

// Get All Raw rolebindings list
func getAllRawRoleBindings(ctx context.Context, clientset *typedRbacv1.RbacV1Client, namespace string, labelSelector string) ([]rbacv1.RoleBinding, error) {
	listOpt := metav1.ListOptions{}
	if len(labelSelector) > 0 {
		listOpt = metav1.ListOptions{LabelSelector: labelSelector}
	}

	roleBindings, err := clientset.RoleBindings(namespace).List(ctx, listOpt)
	if err != nil {
		return nil, err
	}

	return roleBindings.Items, nil
}

// Get All Raw serviceaccount list
func getAllRawServiceAccount(ctx context.Context, clientset *kubernetes.Clientset, namespace string, labelSelector string) ([]corev1.ServiceAccount, error) {
	listOpt := metav1.ListOptions{}
	if len(labelSelector) > 0 {
		listOpt = metav1.ListOptions{LabelSelector: labelSelector}
	}

	serviceaccounts, err := clientset.CoreV1().ServiceAccounts(namespace).List(ctx, listOpt)
	if err != nil {
		return nil, err
	}

	return serviceaccounts.Items, nil
}

//Retrive only node List for ssh
func getNodeListForOption(clientset *kubernetes.Clientset) []string {
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
				labels := createLabelForOption(objectMeta.Labels)
				nodeList = append(nodeList, fmt.Sprintf("%s (%s)", nodeAddr.Address,labels))
				break
			}

		}
	}

	return nodeList
}

//Create Label to display for options
func createLabelForOption(labels map[string]string) string  {
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
func portForwardToPod(req PortForwardAPodRequest) error {
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
func getTargetNode(clientset *kubernetes.Clientset, args []string) (string,error) {
	// Pass from command
	if len(args) == 1 {
		return args[0], nil
	}

	options := getNodeListForOption(clientset)

	if len(options) == 0 {
		return NO_STRING, fmt.Errorf("No node list")
	}

	var node string
	if len(options) > 0 {
		prompt := &survey.Select{
			Message: "Choose a node:",
			Options: options,
		}
		survey.AskOne(prompt, &node)
	}

	return strings.Split(node, " ")[0], nil
}

// Render ServiceAccount list
func renderServiceAccountsListInfo(serviceaccounts []corev1.ServiceAccount) bool {
	if len(serviceaccounts) <= 0 {
		return false
	}
	// Table setup
	table := table.GetTableObject()
	table.SetHeader([]string{"NAME", "SECRET COUNT", "KEYS", "IAM ROLE", "AGE"})

	now := time.Now()
	for _, serviceaccount := range serviceaccounts {
		var iamRole string
		objectMeta := serviceaccount.ObjectMeta
		duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

		count := len(serviceaccount.Secrets)
		keyGroups := []string{}
		for _, secret := range serviceaccount.Secrets {
			keyGroups = append(keyGroups, secret.Name)

			// Only shows first five secret keys
			if len(keyGroups) == 5 {
				break
			}
		}

		for key,value := range objectMeta.Annotations {
			if key == AWS_IAM_ANNOTATION {
				iamRole = strings.Split(value, "/")[1]
				break
			}
		}

		table.Append([]string{objectMeta.Name, strconv.Itoa(count), strings.Join(keyGroups, ","), iamRole, duration})
	}
	table.Render()

	return true
}

// Render Secret list
func renderSecretsListInfo(secrets []corev1.Secret) bool {
	if len(secrets) <= 0 {
		return false
	}
	// Table setup
	table := table.GetTableObject()
	table.SetHeader([]string{"Name", "TYPE", "DATA COUNT", "FIRST FIVE KEYS", "AGE"})

	now := time.Now()
	for _, secret := range secrets {
		objectMeta := secret.ObjectMeta
		duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

		count := len(secret.Data)
		keyGroups := []string{}
		for key, _ := range secret.Data {
			keyGroups = append(keyGroups, key)

			// Only shows first five secret keys
			if len(keyGroups) == 5 {
				break
			}
		}

		table.Append([]string{objectMeta.Name, string(secret.Type), strconv.Itoa(count), strings.Join(keyGroups, ","), duration})
	}
	table.Render()

	return true
}

// Render Role list
func renderRolesListInfo(roles []rbacv1.Role) bool {
	if len(roles) <= 0 {
		return false
	}
	// Table setup
	table := table.GetTableObject()
	table.SetHeader([]string{"Name", "AGE"})

	now := time.Now()
	for _, role := range roles {
		objectMeta := role.ObjectMeta
		duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

		table.Append([]string{objectMeta.Name, duration})
	}
	table.Render()

	return true
}

// Render Role Binding list
func renderRoleBindingsListInfo(roleBindings []rbacv1.RoleBinding) bool {
	if len(roleBindings) <= 0 {
		return false
	}
	// Table setup
	table := table.GetTableObject()
	table.SetHeader([]string{"Name", "AGE"})

	now := time.Now()
	for _, roleBinding := range roleBindings {
		objectMeta := roleBinding.ObjectMeta
		duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

		table.Append([]string{objectMeta.Name, duration})
	}
	table.Render()

	return true
}

// Render Cluster Role list
func renderClusterRolesListInfo(clusterRoles []rbacv1.ClusterRole) bool {
	if len(clusterRoles) <= 0 {
		return false
	}
	// Table setup
	table := table.GetTableObject()
	table.SetHeader([]string{"Name", "AGE"})

	now := time.Now()
	for _, clusterRole := range clusterRoles {
		objectMeta := clusterRole.ObjectMeta
		duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

		table.Append([]string{objectMeta.Name, duration})
	}
	table.Render()

	return true
}

// Render Cluster Role Binding list
func renderClusterRoleBindingsListInfo(clusterRoleBindings []rbacv1.ClusterRoleBinding) bool {
	if len(clusterRoleBindings) <= 0 {
		return false
	}
	// Table setup
	table := table.GetTableObject()
	table.SetHeader([]string{"Name", "AGE"})

	now := time.Now()
	for _, clusterRoleBinding := range clusterRoleBindings {
		objectMeta := clusterRoleBinding.ObjectMeta
		duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

		table.Append([]string{objectMeta.Name, duration})
	}
	table.Render()

	return true
}

// Render ConfigMap list
func renderConfigMapsListInfo(configmaps []corev1.ConfigMap) bool {
	if len(configmaps) <= 0 {
		return false
	}
	// Table setup
	table := table.GetTableObject()
	table.SetHeader([]string{"Name", "DATA COUNT", "FIRST FIVE KEYS", "AGE"})

	now := time.Now()
	for _, configmap := range configmaps {
		objectMeta := configmap.ObjectMeta
		duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

		count := len(configmap.Data)
		keyGroups := []string{}
		for key, _ := range configmap.Data {
			keyGroups = append(keyGroups, key)

			// Only shows first five configmap keys
			if len(keyGroups) == 5 {
				break
			}
		}

		table.Append([]string{objectMeta.Name, strconv.Itoa(count), strings.Join(keyGroups, ","), duration})
	}
	table.Render()

	return true
}

// Render Pod list
func renderPodListInfo(pods []corev1.Pod) bool {
	if len(pods) <= 0 {
		return false
	}
	// Table setup
	table := table.GetTableObject()
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
				status = fmt.Sprintf("%s", containerStatus.State.Terminated.Reason)
			}
		}

		table.Append([]string{objectMeta.Name, strconv.Itoa(readyCount) + "/" + strconv.Itoa(totalCount), status, podSpec.Hostname, podStatus.PodIP, podStatus.HostIP, podSpec.NodeName, duration})
	}
	table.Render()

	return true
}


// Render Pod list
func renderNodeListInfo(nodes []corev1.Node) bool {
	if len(nodes) <= 0 {
		return false
	}
	//Variable for all pods
	var objectMeta metav1.ObjectMeta
	var externalIp, internalIp string
	now := time.Now()

	// Table setup
	table := table.GetTableObject()
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

	return true
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
