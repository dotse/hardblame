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

var pings, fetches, updates int

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Send a ping request to the sharkd server, used for debugging",
	Run: func(cmd *cobra.Command, args []string) {
		PingSharkdServer()
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	pingCmd.Flags().IntVarP(&pings, "count", "c", 1, "ping counter to send to server")
}

func PingSharkdServer() {

	data := sharklib.PingPost{
		Pings: pings,
	}

	bytebuf := new(bytes.Buffer)
	json.NewEncoder(bytebuf).Encode(data)

	status, buf, err := api.Post("/ping", bytebuf.Bytes())
	if err != nil {
		log.Println("Error from Api Post:", err)
		return
	}
	if verbose {
		fmt.Printf("Status: %d\n", status)
	}

	var pr sharklib.PingResponse

	err = json.Unmarshal(buf, &pr)
	if err != nil {
		log.Fatalf("Error from unmarshal: %v\n", err)
	}

	fmt.Printf("Pings: %d Pongs: %d Message: %s\n", pr.Pings, pr.Pongs, pr.Message)
}
