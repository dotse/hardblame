/*
*/
package cmd

import (
        "bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/ryanuber/columnize"
	"github.com/dotse/hardblame/lib"
)

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("group called")
	},
}

var groupCountCmd = &cobra.Command{
	Use:   "count",
	Short: "Cound number of domains per group",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		gr := SendGroup("count")
		// fmt.Printf("GR: %v\n", gr)
		PrintCounts(gr.Counts)
	},
}

func init() {
	rootCmd.AddCommand(groupCmd)
	groupCmd.AddCommand(groupCountCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// groupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// groupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func SendGroup(command string) sharklib.GroupResponse {

	data := sharklib.GroupPost{
		Command:	command,
	}

	bytebuf := new(bytes.Buffer)
	json.NewEncoder(bytebuf).Encode(data)

	status, buf, err := api.Post("/group", bytebuf.Bytes())
	if err != nil {
		log.Fatalf("Error from Api Post:", err)
	}
	if verbose {
		fmt.Printf("Status: %d\n", status)
	}

	var gr sharklib.GroupResponse

	err = json.Unmarshal(buf, &gr)
	if err != nil {
		log.Fatalf("Error from unmarshal: %v\n", err)
	}
	return gr
}

func PrintCounts(gc map[string]int) {
     var out []string
     if showhdr {
     	out = append(out, "Group|Count")
     }

     for group, count := range gc {
     	    row := fmt.Sprintf("%s|%d", group, count)
	    out = append(out, row)
     }
     fmt.Printf("%s\n", columnize.SimpleFormat(out))
}