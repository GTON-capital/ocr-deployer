package deployer

import (
	"math/big"

	// /"ocr-deploy-script/client"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/portto/solana-go-sdk/types"
	ifClient "github.com/smartcontractkit/integrations-framework/client"
	bindings "github.com/smartcontractkit/integrations-framework/contracts"
	"github.com/smartcontractkit/integrations-framework/contracts/ethereum"

	ocrConfigHelper "github.com/smartcontractkit/libocr/offchainreporting/confighelper"
)

// EthereumContractDeployer provides the implementations for deploying ETH (EVM) based contracts
type ContractDeployer struct {
	eth *ifClient.EthereumClient
}

// NewEthereumContractDeployer returns an instantiated instance of the ETH contract deployer
func NewEthereumContractDeployer(ethClient *ifClient.EthereumClient) *ContractDeployer {
	return &ContractDeployer{
		eth: ethClient,
	}
}

// DefaultOffChainAggregatorOptions returns some base defaults for deploying an OCR contract
func DefaultOffChainAggregatorOptions() bindings.OffchainOptions {
	return bindings.OffchainOptions{
		MaximumGasPrice:         uint32(500000000),
		ReasonableGasPrice:      uint32(28000),
		MicroLinkPerEth:         uint32(500),
		LinkGweiPerObservation:  uint32(500),
		LinkGweiPerTransmission: uint32(500),
		MinimumAnswer:           big.NewInt(1),
		MaximumAnswer:           big.NewInt(5000),
		Decimals:                8,
		Description:             "Test OCR",
	}
}

// DefaultOffChainAggregatorConfig returns some base defaults for configuring an OCR contract
func DefaultOffChainAggregatorConfig() bindings.OffChainAggregatorConfig {
	return bindings.OffChainAggregatorConfig{
		AlphaPPB:         1,
		DeltaC:           time.Minute * 10,
		DeltaGrace:       time.Second,
		DeltaProgress:    time.Second * 30,
		DeltaStage:       time.Second * 10,
		DeltaResend:      time.Second * 10,
		DeltaRound:       time.Second * 20,
		RMax:             4,
		S:                []int{1, 1, 1, 1, 1},
		N:                5,
		F:                1,
		OracleIdentities: []ocrConfigHelper.OracleIdentityExtra{},
	}
}

// DeployOffChainAggregator deploys the offchain aggregation contract to the EVM chain
func (e *ContractDeployer) DeployOffChainAggregator(
	fromWallet ifClient.EthereumWallet,
	offchainOptions bindings.OffchainOptions,
) (bindings.OffchainAggregator, error) {
	address, _, instance, err := e.eth.DeployContract("OffChain Aggregator", func(
		auth *bind.TransactOpts,
		backend bind.ContractBackend,
	) (common.Address, *types.Transaction, interface{}, error) {
		gasPrice, err := e.eth AdjustGasPrice()
		if err != nil {
			return common.Address{}, nil, nil, err
		}
		auth.GasPrice = gasPrice
		linkAddress := common.HexToAddress(e.eth.NetworkConfig.Config().LinkTokenAddress)
		return ethereum.DeployOffchainAggregator(auth,
			backend,
			offchainOptions.MaximumGasPrice,
			offchainOptions.ReasonableGasPrice,
			offchainOptions.MicroLinkPerEth,
			offchainOptions.LinkGweiPerObservation,
			offchainOptions.LinkGweiPerTransmission,
			linkAddress,
			offchainOptions.MinimumAnswer,
			offchainOptions.MaximumAnswer,
			offchainOptions.BillingAccessController,
			offchainOptions.RequesterAccessController,
			offchainOptions.Decimals,
			offchainOptions.Description)
	})
	if err != nil {
		return nil, err
	}
	return &EthereumOffchainAggregator{
		client:       e.eth,
		ocr:          instance.(*ethereum.OffchainAggregator),
		callerWallet: fromWallet,
		address:      address,
	}, err
}

// func (e *ContractDeployer) AdjustGasPrice() (*big.Int, error) {
// 	gasPrice, err := e.eth.Client.SuggestGasPrice(context.Background())
// 	if err != nil {
// 		return nil, err
// 	}
// 	chainId := e.eth.Network.ChainID()
// 	if chainId == big.NewInt(30) || chainId == big.NewInt(31) || chainId == big.NewInt(33) {
// 		x, y, z := big.NewInt(0), big.NewInt(0), big.NewInt(0)
// 		x.Add(gasPrice, y.Div(z.Mul(gasPrice, big.NewInt(2)), big.NewInt(100)))
// 		return x, nil
// 	} else {
// 		return gasPrice, nil
// 	}
// }

// NewEthereumClient returns an instantiated instance of the Ethereum client that has connected to the server
// func NewEthereumClient(network ifClient.BlockchainClientURLFn) (*RskClient, error) {
// 	cl, err := ethclient.Dial(network.URL())
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &ContractDeployer{
// 		&ifClient.EthereumClient{
// 			Network: network,
// 			Client:  cl,
// 		},
// 	}, nil
// }
