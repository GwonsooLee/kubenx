package cmd

import (
	"fmt"
	"os"
	"strconv"
	"github.com/AlecAivazis/survey/v2"
	"github.com/olekukonko/tablewriter"
)

var (
	NO_STRING=""
	DEFAULT_NODE_LABEL_FILTERS=[]string{"app", "env"}
)

// Get Home Directory
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
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

//Convert string to Int64
func _string_to_int64(s string) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		Red(err)
		os.Exit(1)
	}

	return n
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
func _get_single_string_input(message string, error_msg string) string {
	var ret string
	prompt := &survey.Input{
		Message: fmt.Sprintf("%s:", message),
	}
	survey.AskOne(prompt, &ret)

	if ret == "" {
		Red(error_msg)
		os.Exit(1)
	}

	return ret
}
