package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"os"
	"reflect"
	"strconv"
	"github.com/AlecAivazis/survey/v2"
	"github.com/olekukonko/tablewriter"
	"unsafe"
)

var (
	ALL_NAMESPACE=""
	NO_STRING=""
	DEFAULT_NODE_LABEL_FILTERS=[]string{"app", "env"}
	//STATIS VALUE
	KUBENX_HOMEDIR 		= ".kubenx"
	SSH_DEFAULT_PATH 	= "ssh"
	TARGET_DEFAULT_PORT = "22"
	AWS_IAM_ANNOTATION 	= "eks.amazonaws.com/role-arn"
	AUTH_API_VERSION 	= "client.authentication.k8s.io/v1alpha1"
	AUTH_COMMAND		= "aws"


	//Color Definition
	Red    = color.New(color.FgRed).PrintlnFunc()
	Blue   = color.New(color.FgBlue).PrintlnFunc()
	Green  = color.New(color.FgGreen).PrintlnFunc()
	Yellow = color.New(color.FgYellow).PrintlnFunc()
	Cyan   = color.New(color.FgCyan).PrintlnFunc()

	//OPEN_ID_CA_FINGERPRINT
	CA_FINGERPRINT = "9e99a48a9960b14926bb7f3b02e22da2b0ab7280"

	//Error Message
	NO_FILE_EXCEPTION = "No file exists... Please check the file path"
)

// Get Home Directory
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

//Figure out if string is in array
func isStringInArr(s string, arr []string) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}

	return false
}

// Get Table
func _get_table_object() *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")

	return table
}

//Convert Int32 to String
func _int32_to_string(num int32) string {
	return strconv.FormatInt(int64(num), 10)
}

//Convert int to string
func _string_to_int(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		Red(err)
		os.Exit(1)
	}
	return n
}

//Get only one input from user
func getSingleStringInput(message string) (string, error) {
	var ret string
	prompt := &survey.Input{
		Message: fmt.Sprintf("%s:", message),
	}
	survey.AskOne(prompt, &ret)

	if ret == "" {
		return NO_STRING, fmt.Errorf("Choice has been canceled")
	}

	return ret, nil
}

func BytesToString(bytes []byte) (s string) {
	hdr := *(*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: hdr.Data,
		Len:  hdr.Len,
	}))
}
func StringToBytes(str string) []byte {
	hdr := *(*reflect.StringHeader)(unsafe.Pointer(&str))
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: hdr.Data,
		Len:  hdr.Len,
		Cap:  hdr.Len,
	}))
}
