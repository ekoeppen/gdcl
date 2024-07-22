package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/40hz/newton/gdcl/v3/nsof"
	"log"
	"os"
)

var (
	input  string
	output string
	decode bool
	encode bool
)

var nsofCmd = &cobra.Command{
	Use:   "nsof",
	Short: "NSOF commands",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := os.ReadFile(input)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		if encode {
			return
		}
		if decode {
			var encoded nsof.Data = data
			decoded := encoded.Factory()
			log.Println(decoded)
			return
		}
		log.Fatalln("Specify either encode or decode")
	},
}

func init() {
	rootCmd.AddCommand(nsofCmd)
	nsofCmd.Flags().StringVarP(&input, "input", "i", "", "Input file")
	nsofCmd.Flags().StringVarP(&output, "output", "o", "", "Output file")
	nsofCmd.Flags().BoolVarP(&encode, "encode", "e", false, "Encode data")
	nsofCmd.Flags().BoolVarP(&decode, "decode", "d", true, "Decode data")
}
