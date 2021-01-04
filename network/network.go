package network

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/DrDelphi/ElrondDSSC/data"
	"github.com/DrDelphi/ElrondDSSC/utils"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-sdk/erdgo"
)

var log = logger.GetOrCreate("network")

// NetworkManager - holds the required fields of a network manager
type NetworkManager struct {
	networkConfig *data.NetworkConfig
	networkAPI    string
	networkProxy  string
	metaObserver  string
	fDenomination *big.Float
	Proxy         *erdgo.ElrondProxy
}

// NewNetworkManager - creates a new NetworkManager object
func NewNetworkManager(cfg *data.AppConfig) (*NetworkManager, error) {
	bytes, err := utils.GetHTTP(cfg.MetaObserver + "/network/config")
	if err != nil {
		log.Error("can not get network config from meta observer", "error", err)
		return nil, err
	}

	networkConfig := &data.NetworkConfig{}
	err = json.Unmarshal(bytes, networkConfig)
	if err != nil {
		log.Error("can not unmarshal network config", "error", err)
		return nil, err
	}

	fDenomination := big.NewFloat(1)
	for i := 0; i < networkConfig.Data.Config.ErdDenomination; i++ {
		fDenomination.Mul(fDenomination, big.NewFloat(10))
	}

	networkManager := &NetworkManager{
		networkConfig: networkConfig,
		networkAPI:    cfg.NetworkAPI,
		networkProxy:  cfg.NetworkProxy,
		metaObserver:  cfg.MetaObserver,
		fDenomination: fDenomination,
		Proxy:         erdgo.NewElrondProxy(cfg.NetworkProxy),
	}

	return networkManager, nil
}

// GetDenominator - returns a big.Float of 10 ^ decimals
func (nm *NetworkManager) GetDenominator() *big.Float {
	return nm.fDenomination
}

// GetUserActiveStake - retrieves an address' active stake delegated in the DSSC
func (nm *NetworkManager) GetUserActiveStake(address string) (*big.Float, error) {
	pubkey, _ := erdgo.Bech32ToPubkey(address)
	hexAddress := hex.EncodeToString(pubkey)
	iStake, err := nm.queryScIntResult(utils.ContractAddress, "getUserActiveStake", []string{hexAddress})
	if err != nil {
		return nil, err
	}
	fStake := big.NewFloat(0).SetInt(iStake)
	fStake.Quo(fStake, nm.fDenomination)

	return fStake, nil
}

// GetUserUnBondable - retrieves an address' unbondable stake from the DSSC
func (nm *NetworkManager) GetUserUnBondable(address string) (*big.Float, error) {
	pubkey, _ := erdgo.Bech32ToPubkey(address)
	hexAddress := hex.EncodeToString(pubkey)
	iStake, err := nm.queryScIntResult(utils.ContractAddress, "getUserUnBondable", []string{hexAddress})
	if err != nil {
		return nil, err
	}
	fStake := big.NewFloat(0).SetInt(iStake)
	fStake.Quo(fStake, nm.fDenomination)

	return fStake, nil
}

// GetUserUnStakedValue - retrieves an address' unstaked value from the DSSC
func (nm *NetworkManager) GetUserUnStakedValue(address string) (*big.Float, error) {
	pubkey, _ := erdgo.Bech32ToPubkey(address)
	hexAddress := hex.EncodeToString(pubkey)
	iStake, err := nm.queryScIntResult(utils.ContractAddress, "getUserUnStakedValue", []string{hexAddress})
	if err != nil {
		return nil, err
	}
	fStake := big.NewFloat(0).SetInt(iStake)
	fStake.Quo(fStake, nm.fDenomination)

	return fStake, nil
}

// GetClaimableRewards - retrieves an address' claimable rewards from the DSSC
func (nm *NetworkManager) GetClaimableRewards(address string) (*big.Float, error) {
	pubkey, _ := erdgo.Bech32ToPubkey(address)
	hexAddress := hex.EncodeToString(pubkey)
	iStake, err := nm.queryScIntResult(utils.ContractAddress, "getClaimableRewards", []string{hexAddress})
	if err != nil {
		return nil, err
	}
	fStake := big.NewFloat(0).SetInt(iStake)
	fStake.Quo(fStake, nm.fDenomination)

	return fStake, nil
}

// GetLastTxs - retrieves from the API the last in / out transactions to / from a specified address
func (nm *NetworkManager) GetLastTxs(address string, size int, inout string) ([]*indexer.Transaction, error) {
	endpoint := fmt.Sprintf("%s/transactions?from=0&size=%v&%s=%s", nm.networkAPI, size, inout, address)
	bytes, err := utils.GetHTTP(endpoint)
	if err != nil {
		return nil, err
	}

	list := make([]*indexer.Transaction, 0)
	err = json.Unmarshal(bytes, &list)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (nm *NetworkManager) queryScIntResult(scAddress, funcName string, args []string) (*big.Int, error) {
	query := &data.ScQuery{
		ScAddress: scAddress,
		FuncName:  funcName,
		Args:      args,
	}
	host := fmt.Sprintf("%s/vm-values/int", nm.networkProxy)
	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	bytes, err := utils.PostHTTP(host, string(body))
	if err != nil {
		return nil, err
	}
	res := &data.ScIntResult{}
	err = json.Unmarshal(bytes, res)
	if err != nil {
		return nil, err
	}
	intRes, ok := big.NewInt(0).SetString(res.Data.Data, 10)
	if !ok {
		return nil, errors.New("invalid result: " + res.Error)
	}

	return intRes, nil
}

func (nm *NetworkManager) queryScQueryResult(scAddress, funcName string, args []string) ([][]byte, error) {
	query := &data.ScQuery{
		ScAddress: scAddress,
		FuncName:  funcName,
		Args:      args,
	}
	host := fmt.Sprintf("%s/vm-values/query", nm.networkProxy)
	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	bytes, err := utils.PostHTTP(host, string(body))
	if err != nil {
		return nil, err
	}
	res := &data.ScQueryResult{}
	err = json.Unmarshal(bytes, res)
	if err != nil {
		return nil, err
	}

	return res.Data.Data.ReturnData, nil
}

// CreateDSSC - sends a create DSSC transaction
func (nm *NetworkManager) CreateDSSC(privateKey string) (string, error) {
	privateKeyBytes, _ := hex.DecodeString(privateKey)
	address, _ := erdgo.GetAddressFromPrivateKey(privateKeyBytes)
	account, err := nm.Proxy.GetAccount(address)
	if err != nil {
		return "", err
	}

	tx := &erdgo.Transaction{
		Nonce:    account.Nonce,
		RcvAddr:  "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqylllslmq6y6",
		SndAddr:  address,
		Value:    "1250000000000000000000",
		GasPrice: 1000000000,
		GasLimit: 60000000,
		ChainID:  nm.networkConfig.Data.Config.ErdChainID,
		Version:  1,
		Data:     []byte("createNewDelegationContract@00@00"),
	}
	err = erdgo.SignTransaction(tx, privateKeyBytes)
	if err != nil {
		return "", err
	}

	return nm.Proxy.SendTransaction(tx)
}

// GetContractInfo - retrieves details about the DSSC
func (nm *NetworkManager) GetContractInfo(address string) (*data.ContractInfo, error) {
	query, err := nm.queryScQueryResult(address, "getContractConfig", make([]string, 0))
	if err != nil {
		log.Error("can not get contract info", "error", err)
		return nil, err
	}

	if len(query) != 9 {
		return nil, errors.New("invalid response")
	}

	iServiceFee := big.NewInt(0).SetBytes(query[1])
	fServiceFee := big.NewFloat(0).SetInt(iServiceFee)
	fServiceFee.Quo(fServiceFee, big.NewFloat(100))
	serviceFee, _ := fServiceFee.Float64()

	iMaxDelegationCap := big.NewInt(0).SetBytes(query[2])
	fMaxDelegationCap := big.NewFloat(0).SetInt(iMaxDelegationCap)
	fMaxDelegationCap.Quo(fMaxDelegationCap, nm.fDenomination)
	maxDelegationCap, _ := fMaxDelegationCap.Float64()

	iInitialOwnerFunds := big.NewInt(0).SetBytes(query[3])
	fInitialOwnerFunds := big.NewFloat(0).SetInt(iInitialOwnerFunds)
	fInitialOwnerFunds.Quo(fInitialOwnerFunds, nm.fDenomination)
	initialOwnerFunds, _ := fInitialOwnerFunds.Float64()

	iCreatedNonce := big.NewInt(0).SetBytes(query[7])

	iUnBondPeriod := big.NewInt(0).SetBytes(query[8])

	info := &data.ContractInfo{
		ServiceFee:           serviceFee,
		MaxDelegationCap:     maxDelegationCap,
		InitialOwnerFunds:    initialOwnerFunds,
		AutomaticActivation:  string(query[4]) == "true",
		WithDelegationCap:    string(query[5]) == "true",
		ChangeableServiceFee: string(query[6]) == "true",
		CreatedNonce:         iCreatedNonce.Uint64(),
		UnBondPeriod:         iUnBondPeriod.Uint64(),
	}

	return info, nil
}

func (nm *NetworkManager) getScIntNoArgs(fnc string) (*big.Float, error) {
	i, err := nm.queryScIntResult(utils.ContractAddress, fnc, make([]string, 0))
	if err != nil {
		log.Error("can not get SC int result", "error", err)
		return nil, err
	}

	f := big.NewFloat(0).SetInt(i)
	f.Quo(f, nm.fDenomination)

	return f, nil
}

// GetTotalActiveStake - retrieves the total active stake from the DSSC
func (nm *NetworkManager) GetTotalActiveStake() (*big.Float, error) {
	return nm.getScIntNoArgs("getTotalActiveStake")
}

// GetTotalUnStaked - retrieves the total active stake from the DSSC
func (nm *NetworkManager) GetTotalUnStaked() (*big.Float, error) {
	return nm.getScIntNoArgs("getTotalUnStaked")
}

// GetTotalCumulatedRewards - retrieves the total cumulated rewards from the DSSC
func (nm *NetworkManager) GetTotalCumulatedRewards() (*big.Float, error) {
	query := &data.ScQuery{
		ScAddress: utils.ContractAddress,
		FuncName:  "getTotalCumulatedRewards",
		Caller:    "erd1qqqqqqqqqqqqqqqpqqqqqqqqlllllllllllllllllllllllllllsr9gav8",
	}
	host := fmt.Sprintf("%s/vm-values/int", nm.networkProxy)
	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	bytes, err := utils.PostHTTP(host, string(body))
	if err != nil {
		return nil, err
	}
	res := &data.ScIntResult{}
	err = json.Unmarshal(bytes, res)
	if err != nil {
		return nil, err
	}
	intRes, ok := big.NewInt(0).SetString(res.Data.Data, 10)
	if !ok {
		return nil, errors.New("invalid result: " + res.Error)
	}

	f := big.NewFloat(0).SetInt(intRes)
	f.Quo(f, nm.fDenomination)

	return f, nil
}

// GetTotalUnStakedFromNodes - retrieves the total unstaked from nodes from the DSSC
func (nm *NetworkManager) GetTotalUnStakedFromNodes() (*big.Float, error) {
	return nm.getScIntNoArgs("getTotalUnStakedFromNodes")
}

// GetTotalUnBondedFromNodes - retrieves the total unbonded from nodes from the DSSC
func (nm *NetworkManager) GetTotalUnBondedFromNodes() (*big.Float, error) {
	return nm.getScIntNoArgs("getTotalUnBondedFromNodes")
}

// GetNumUsers - retrieves the number of delegators from the DSSC
func (nm *NetworkManager) GetNumUsers() (uint64, error) {
	iUsers, err := nm.queryScIntResult(utils.ContractAddress, "getNumUsers", make([]string, 0))
	if err != nil {
		return 0, err
	}

	return iUsers.Uint64(), nil
}

// GetNumNodes - retrieves the number of nodes from the DSSC
func (nm *NetworkManager) GetNumNodes() (uint64, error) {
	iNodes, err := nm.queryScIntResult(utils.ContractAddress, "getNumNodes", make([]string, 0))
	if err != nil {
		return 0, err
	}

	return iNodes.Uint64(), nil
}

// GetUserUnDelegatedList - retrieves an address' undelegated list from the DSSC
func (nm *NetworkManager) GetUserUnDelegatedList(address string) ([][]byte, error) {
	pubkey, _ := erdgo.Bech32ToPubkey(address)
	hexAddress := hex.EncodeToString(pubkey)
	query, err := nm.queryScQueryResult(utils.ContractAddress, "getUserUnDelegatedList", []string{hexAddress})
	if err != nil {
		log.Error("can not get user undelegated list", "error", err)
		return nil, err
	}

	if len(query)%2 != 0 {
		return nil, errors.New("invalid response")
	}

	return query, nil
}

// GetAllNodeStates - retrieves all nodes states list from the DSSC
func (nm *NetworkManager) GetAllNodeStates() ([][]byte, error) {
	query, err := nm.queryScQueryResult(utils.ContractAddress, "getAllNodeStates", make([]string, 0))
	if err != nil {
		log.Error("can not get nodes states list", "error", err)
		return nil, err
	}

	if len(query)%2 != 0 {
		return nil, errors.New("invalid response")
	}

	return query, nil
}
