package deployer

EthereumOffchainAggregator represents the offchain aggregation contract
type EthereumOffchainAggregator struct {
	client       *client.EthereumClient
	ocr          *ethereum.OffchainAggregator
	callerWallet client.BlockchainWallet
	address      *common.Address
}

// Fund sends specified currencies to the contract
func (o *EthereumOffchainAggregator) Fund(fromWallet client.BlockchainWallet, ethAmount, linkAmount *big.Int) error {
	return o.client.Fund(fromWallet, o.address.Hex(), ethAmount, linkAmount)
}

// GetContractData retrieves basic data for the offchain aggregator contract
func (o *EthereumOffchainAggregator) GetContractData(ctxt context.Context) (*contracts.OffchainAggregatorData, error) {
	opts := &bind.CallOpts{
		From:    common.HexToAddress(o.callerWallet.Address()),
		Pending: true,
		Context: ctxt,
	}

	lr, err := o.ocr.LatestRoundData(opts)
	if err != nil {
		return &contracts.OffchainAggregatorData{}, err
	}
	latestRound := contracts.RoundData(lr)

	return &contracts.OffchainAggregatorData{
		LatestRoundData: latestRound,
	}, nil
}

// SetPayees sets wallets for the contract to pay out to?
func (o *EthereumOffchainAggregator) SetPayees(
	fromWallet client.BlockchainWallet,
	transmitters, payees []common.Address,
) error {
	opts, err := o.client.TransactionOpts(fromWallet, *o.address, big.NewInt(0), nil)
	if err != nil {
		return err
	}

	tx, err := o.ocr.SetPayees(opts, transmitters, payees)
	if err != nil {
		return err
	}
	return o.client.WaitForTransaction(tx.Hash())
}

// SetConfig sets offchain reporting protocol configuration including participating oracles
func (o *EthereumOffchainAggregator) SetConfig(
	fromWallet client.BlockchainWallet,
	chainlinkNodes []client.Chainlink,
	ocrConfig contracts.OffChainAggregatorConfig,
) error {
	// Gather necessary addresses and keys from our chainlink nodes to properly configure the OCR contract
	for _, node := range chainlinkNodes {
		ocrKeys, err := node.ReadOCRKeys()
		if err != nil {
			return err
		}
		primaryOCRKey := ocrKeys.Data[0]
		ethKeys, err := node.ReadETHKeys()
		if err != nil {
			return err
		}
		primaryEthKey := ethKeys.Data[0]
		p2pKeys, err := node.ReadP2PKeys()
		if err != nil {
			return err
		}
		primaryP2PKey := p2pKeys.Data[0]

		// Need to convert the key representations
		var onChainSigningAddress [20]byte
		var configPublicKey [32]byte
		offchainSigningAddress, err := hex.DecodeString(primaryOCRKey.Attributes.OffChainPublicKey)
		if err != nil {
			return err
		}
		decodeConfigKey, err := hex.DecodeString(primaryOCRKey.Attributes.ConfigPublicKey)
		if err != nil {
			return err
		}

		// https://stackoverflow.com/questions/8032170/how-to-assign-string-to-bytes-array
		copy(onChainSigningAddress[:], common.HexToAddress(primaryOCRKey.Attributes.OnChainSigningAddress).Bytes())
		copy(configPublicKey[:], decodeConfigKey)

		oracleIdentity := ocrConfigHelper.OracleIdentity{
			TransmitAddress:       common.HexToAddress(primaryEthKey.Attributes.Address),
			OnChainSigningAddress: onChainSigningAddress,
			PeerID:                primaryP2PKey.Attributes.PeerID,
			OffchainPublicKey:     offchainSigningAddress,
		}
		oracleIdentityExtra := ocrConfigHelper.OracleIdentityExtra{
			OracleIdentity:                  oracleIdentity,
			SharedSecretEncryptionPublicKey: ocrTypes.SharedSecretEncryptionPublicKey(configPublicKey),
		}

		ocrConfig.OracleIdentities = append(ocrConfig.OracleIdentities, oracleIdentityExtra)
	}

	signers, transmitters, threshold, encodedConfigVersion, encodedConfig, err := ocrConfigHelper.ContractSetConfigArgs(
		ocrConfig.DeltaProgress,
		ocrConfig.DeltaResend,
		ocrConfig.DeltaRound,
		ocrConfig.DeltaGrace,
		ocrConfig.DeltaC,
		ocrConfig.AlphaPPB,
		ocrConfig.DeltaStage,
		ocrConfig.RMax,
		ocrConfig.S,
		ocrConfig.OracleIdentities,
		ocrConfig.F,
	)
	if err != nil {
		return err
	}

	// Set Payees
	opts, err := o.client.TransactionOpts(fromWallet, *o.address, big.NewInt(0), nil)
	if err != nil {
		return err
	}

	tx, err := o.ocr.SetPayees(opts, transmitters, transmitters)
	if err != nil {
		return err
	}
	err = o.client.WaitForTransaction(tx.Hash())
	if err != nil {
		return err
	}

	// Set Config
	opts, err = o.client.TransactionOpts(fromWallet, *o.address, big.NewInt(0), nil)
	if err != nil {
		return err
	}

	tx, err = o.ocr.SetConfig(opts, signers, transmitters, threshold, encodedConfigVersion, encodedConfig)
	if err != nil {
		return err
	}
	return o.client.WaitForTransaction(tx.Hash())
}

// RequestNewRound requests the OCR contract to create a new round
func (o *EthereumOffchainAggregator) RequestNewRound(fromWallet client.BlockchainWallet) error {
	opts, err := o.client.TransactionOpts(fromWallet, *o.address, big.NewInt(0), nil)
	if err != nil {
		return err
	}
	tx, err := o.ocr.RequestNewRound(opts)
	if err != nil {
		return err
	}
	log.Info().Str("Contract Address", o.address.Hex()).Msg("New OCR round requested")
	return o.client.WaitForTransaction(tx.Hash())
}

// Link returns the LINK contract address on the EVM chain
func (o *EthereumOffchainAggregator) Link(ctxt context.Context) (common.Address, error) {
	opts := &bind.CallOpts{
		From:    common.HexToAddress(o.callerWallet.Address()),
		Pending: true,
		Context: ctxt,
	}
	return o.ocr.LINK(opts)
}

// GetLatestAnswer returns the latest answer from the OCR contract
func (o *EthereumOffchainAggregator) GetLatestAnswer(ctxt context.Context) (*big.Int, error) {
	opts := &bind.CallOpts{
		From:    common.HexToAddress(o.callerWallet.Address()),
		Pending: true,
		Context: ctxt,
	}
	return o.ocr.LatestAnswer(opts)
}

func (o *EthereumOffchainAggregator) Address() string {
	return o.address.Hex()
}

// GetLatestRound returns data from the latest round
func (o *EthereumOffchainAggregator) GetLatestRound(ctxt context.Context) (*contracts.RoundData, error) {
	opts := &bind.CallOpts{
		From:    common.HexToAddress(o.callerWallet.Address()),
		Pending: true,
		Context: ctxt,
	}

	roundData, err := o.ocr.LatestRoundData(opts)
	if err != nil {
		return nil, err
	}

	return &contracts.RoundData{
		RoundId:         roundData.RoundId,
		Answer:          roundData.Answer,
		AnsweredInRound: roundData.AnsweredInRound,
		StartedAt:       roundData.StartedAt,
		UpdatedAt:       roundData.UpdatedAt,
	}, err
}
