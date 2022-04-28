/*
 *
 */
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/dotse/hardblame/lib"

	"github.com/spf13/cobra"
)

var hardenizeCmd = &cobra.Command{
	Use:   "hardenize",
	Short: "no op",
}

var hardenizeFetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "no op",
        Run: func(cmd *cobra.Command, args []string) {
                SendHardenize("fetch")
        },
}

var hardenizeTestCmd = &cobra.Command{
	Use:   "test",
	Short: "test",
        Run: func(cmd *cobra.Command, args []string) {
                SendHardenize("test")
        },
}

func init() {
	rootCmd.AddCommand(hardenizeCmd)
	hardenizeCmd.AddCommand(hardenizeFetchCmd, hardenizeTestCmd)

	hardenizeCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Debug output")
}

func SendHardenize(command string) {

	data := sharklib.HardenizePost{
		Command:	command,
	}

	bytebuf := new(bytes.Buffer)
	json.NewEncoder(bytebuf).Encode(data)

	status, buf, err := api.Post("/hardenize", bytebuf.Bytes())
	if err != nil {
		log.Println("Error from Api Post:", err)
		return
	}
	if true {
		fmt.Printf("Status: %d\n", status)
	}

	var hr sharklib.HardenizeResponse

	err = json.Unmarshal(buf, &hr)
	if err != nil {
		log.Fatalf("Error from unmarshal: %v\n", err)
	}
	fmt.Printf("Result: %s\n", hr.Message)
}
