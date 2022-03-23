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

	"ocr-deployer/models"

	"github.com/mitchellh/mapstructure"
	"github.com/smartcontractkit/integrations-framework/client"
	"github.com/smartcontractkit/integrations-framework/config"
	"github.com/smartcontractkit/integrations-framework/contracts"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy OffcahinAggregator contract and configure it",
	Long: `Deploy OffcahinAggregator contract and set it configuration:

1. Set Payees
2. Set Config
`,
	Run: func(cmd *cobra.Command, args []string) {
		deploy(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deployCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func deploy(cmd *cobra.Command, args []string) {
	spec, err := models.LoadSpec(args[0])
	if err != nil {
		log.Fatal(err)
	}

	// cl, err := ethclient.Dial(spec.ChainInfo.RPCURL)
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
	deployer := contracts.NewEthereumContractDeployer(ecl)

	options := contracts.DefaultOffChainAggregatorOptions()
	options.Decimals = 6

	fmt.Printf("ChainID %d\n", ecl.GetChainID())
	fmt.Printf("Wallet %s\n", ecl.Wallets[0].Address())
	fmt.Println(spec.Description)
	aggregator, err := deployer.DeployOffChainAggregator("0x326C977E6efc84E512bB9C30f76E30c160eD06FB", options)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("AggregatorContract %s\n", aggregator.Address())

	log.Println("deploy called")
	fmt.Println("Karamba")
}
