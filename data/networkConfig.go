package data

// NetworkConfig - holds the Elrond network configuration as received from a node's API
type NetworkConfig struct {
	Data struct {
		Config struct {
			ErdChainID                  string `json:"erd_chain_id"`
			ErdDenomination             int    `json:"erd_denomination"`
			ErdGasPerDataByte           uint64 `json:"erd_gas_per_data_byte"`
			ErdLatestTagSoftwareVersion string `json:"erd_latest_tag_software_version"`
			ErdMetaConsensusGroupSize   uint64 `json:"erd_meta_consensus_group_size"`
			ErdMinGasLimit              uint64 `json:"erd_min_gas_limit"`
			ErdMinGasPrice              uint64 `json:"erd_min_gas_price"`
			ErdMinTransactionVersion    int    `json:"erd_min_transaction_version"`
			ErdNumMetachainNodes        uint64 `json:"erd_num_metachain_nodes"`
			ErdNumNodesInShard          uint64 `json:"erd_num_nodes_in_shard"`
			ErdNumShardsWithoutMeta     uint32 `json:"erd_num_shards_without_meta"`
			ErdRoundDuration            int64  `json:"erd_round_duration"`
			ErdShardConsensusGroupSize  uint64 `json:"erd_shard_consensus_group_size"`
			ErdStartTime                int64  `json:"erd_start_time"`
		} `json:"config"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}
