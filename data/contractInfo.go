package data

type ContractInfo struct {
	ServiceFee           float64
	MaxDelegationCap     float64
	InitialOwnerFunds    float64
	AutomaticActivation  bool
	WithDelegationCap    bool
	ChangeableServiceFee bool
	CreatedNonce         uint64
	UnBondPeriod         uint64
}
