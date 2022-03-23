/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mitchellh/mapstructure"
	"github.com/smartcontractkit/integrations-framework/client"
	"github.com/smartcontractkit/integrations-framework/config"
	"github.com/smartcontractkit/integrations-framework/contracts/ethereum"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// getinfoCmd represents the getinfo command
var getinfoCmd = &cobra.Command{
	Use:   "getinfo",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		getInfo(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(getinfoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getinfoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getinfoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func getInfo(cmd *cobra.Command, args []string) {
	networkConfig := map[string]interface{}{}
	cfgb, err := ioutil.ReadFile(networks_cfg)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(cfgb, &networkConfig)
	if err != nil {
		log.Fatal(err)
	}
	networkSettings := config.ETHNetwork{}

	err = mapstructure.Decode(networkConfig, &networkSettings)
	if err != nil {
		log.Fatal(err)
	}

	ecl, err := client.NewEthereumClient(&networkSettings)
	if err != nil {
		log.Fatal(err)
	}

	ba := common.FromHex("0x2cC06F32d51Ced7619C6991759f91797A7fDb789")
	addr := common.BytesToAddress(ba)
	aggregator, err := ethereum.NewOffchainAggregator(addr, ecl.Client)
	if err != nil {
		log.Fatal(err)
	}
	a := bind.CallOpts{}
	desc, err := aggregator.Description(&a)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(desc)

}
