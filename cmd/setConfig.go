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
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"ocr-deployer/models"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mitchellh/mapstructure"
	"github.com/smartcontractkit/integrations-framework/client"
	"github.com/smartcontractkit/integrations-framework/config"
	"github.com/smartcontractkit/integrations-framework/contracts"
	"github.com/smartcontractkit/integrations-framework/contracts/ethereum"
	"github.com/smartcontractkit/libocr/offchainreporting/confighelper"
	"github.com/smartcontractkit/libocr/offchainreporting/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var spec_file string
var aggregator_address string

// setConfigCmd represents the setConfig command
var setConfigCmd = &cobra.Command{
	Use:   "setConfig",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		setConfig(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(setConfigCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setConfigCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setConfigCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringVarP(&spec_file, "spec", "s", "", "path to spec file")
	rootCmd.PersistentFlags().StringVarP(&aggregator_address, "address", "a", "", "address of deployed aggregator")
}

func setConfig(cmd *cobra.Command, args []string) {

	spec, err := models.LoadSpec(spec_file)
	if err != nil {
		log.Fatal(err)
	}
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

	ba := common.FromHex(aggregator_address)
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

	fmt.Printf("Spec description: %s\nContract description: %s\n", spec.Description, desc)

	//Готовим конфиг
	offchainConfig := contracts.DefaultOffChainAggregatorConfig(len(spec.Nodes))
	offchainConfig.DeltaRound = time.Minute * 15
	offchainConfig.DeltaProgress = time.Minute * 20

	payees := []common.Address{}
	for _, node := range spec.Nodes {
		ocrKey := client.OCRKey{
			Data: client.OCRKeyData{
				ID: node.OCRKeyID,
				Attributes: client.OCRKeyAttributes{
					OffChainPublicKey:     node.OCRKey,
					OnChainSigningAddress: node.SigningAddress,
					ConfigPublicKey:       node.ConfigPublicKey,
				},
			},
		}
		ethKey := client.ETHKey{
			Data: client.ETHKeyData{
				Attributes: client.ETHKeyAttributes{
					Address: node.OracleAddress,
				},
			},
		}

		p2pKey := client.P2PKey{
			Data: client.P2PKeyData{
				Attributes: client.P2PKeyAttributes{
					ID:        0,
					PeerID:    node.P2PKey,
					PublicKey: node.P2PPublicKey,
				},
			},
		}

		// Need to convert the key representations
		var onChainSigningAddress [20]byte
		var configPublicKey [32]byte
		offchainSigningAddress, err := hex.DecodeString(ocrKey.Data.Attributes.OffChainPublicKey)
		if err != nil {
			log.Fatal(err)
		}
		decodeConfigKey, err := hex.DecodeString(ocrKey.Data.Attributes.ConfigPublicKey)
		if err != nil {
			log.Fatal(err)
		}

		// https://stackoverflow.com/questions/8032170/how-to-assign-string-to-bytes-array
		copy(onChainSigningAddress[:], common.HexToAddress(ocrKey.Data.Attributes.OnChainSigningAddress).Bytes())
		copy(configPublicKey[:], decodeConfigKey)

		oracleIdentity := confighelper.OracleIdentity{
			TransmitAddress:       common.HexToAddress(ethKey.Data.Attributes.Address),
			OnChainSigningAddress: onChainSigningAddress,
			PeerID:                p2pKey.Data.Attributes.PeerID,
			OffchainPublicKey:     offchainSigningAddress,
		}
		oracleIdentityExtra := confighelper.OracleIdentityExtra{
			OracleIdentity:                  oracleIdentity,
			SharedSecretEncryptionPublicKey: types.SharedSecretEncryptionPublicKey(configPublicKey),
		}

		offchainConfig.OracleIdentities = append(offchainConfig.OracleIdentities, oracleIdentityExtra)
		payees = append(payees, common.HexToAddress(node.OperatorAddress))
	}

	signers, transmitters, threshold, encodedConfigVersion, encodedConfig, err := confighelper.ContractSetConfigArgs(
		offchainConfig.DeltaProgress,
		offchainConfig.DeltaResend,
		offchainConfig.DeltaRound,
		offchainConfig.DeltaGrace,
		offchainConfig.DeltaC,
		offchainConfig.AlphaPPB,
		offchainConfig.DeltaStage,
		offchainConfig.RMax,
		offchainConfig.S,
		offchainConfig.OracleIdentities,
		offchainConfig.F,
	)
	if err != nil {
		log.Fatal(err)
	}
	opts, err := ecl.TransactionOpts(ecl.Wallets[0])
	if err != nil {
		log.Fatal(err)
	}

	// nonce, err := ecl.Client.PendingNonceAt(context.Background(), common.HexToAddress(ecl.DefaultWallet.Address()))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// opts.Nonce = big.NewInt(int64(nonce) + 1)
	// fmt.Printf("Nonce is %d\n", nonce)
	tx, err := aggregator.SetPayees(opts, transmitters, payees)
	if err != nil {
		log.Fatalf("SetPayees error: %v", err)
	}
	fmt.Printf("SetPayees Tx: %s\n", tx.Hash().String())
	r, err := bind.WaitMined(context.TODO(), ecl.Client, tx)
	if err != nil {
		log.Fatalf("SetPayees waiting error: %v", err)
	}
	fmt.Printf("Payees mined with gas %d\n", r.GasUsed)

	// nonce, err = ecl.Client.PendingNonceAt(context.Background(), common.HexToAddress(ecl.DefaultWallet.Address()))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// opts.Nonce = big.NewInt(int64(nonce) + 1)
	// fmt.Printf("Nonce is %d\n", nonce)
	time.Sleep(time.Second * 30)
	opts, err = ecl.TransactionOpts(ecl.Wallets[0])
	if err != nil {
		log.Fatal(err)
	}
	tx, err = aggregator.SetConfig(opts, signers, transmitters, threshold, encodedConfigVersion, encodedConfig)
	if err != nil {
		log.Fatalf("SetConfig error: %v", err)
	}
	fmt.Printf("SetConfig Tx: %s\n", tx.Hash().String())
	r, err = bind.WaitMined(context.TODO(), ecl.Client, tx)
	if err != nil {
		log.Fatalf("SetConfig mined with gas %d\n", r.GasUsed)
	}
	fmt.Printf("SetConfig mined with gas %d\n", r.GasUsed)

}
