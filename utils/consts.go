package utils

const (
	// DefaultConfigPath - default application configuration file path
	DefaultConfigPath = "./config.json"

	// AboutMessage -
	AboutMessage = "`Elrond Delegation System SC interaction Bot ©2021 by Elrond Community`"
	// MainHelp -
	MainHelp = "`MyWallets` - menu for adding the wallets you wish to delegate from and the bot will monitor " +
		"your delegations and rewards\n\r" +
		"`Contract Info` - displays details about the Delegation SC (address, fee, etc.)"
	// MyWalletsHelp -
	MyWalletsHelp = "`Add` - here you can add a wallet to be managed by the bot\n\r" +
		"`Balances` - here you can see each of your wallet's delegations, balances and claimable rewards"

	// SetOwnerAddressMessage -
	SetOwnerAddressMessage = "Send owner's address or PEM/JSON file (for JSONs, first write the password, then attach the file)"
	// AddWalletMessage -
	AddWalletMessage = "Send the wallet's address"
	// DelegateAmountMessage -
	DelegateAmountMessage = "Send amount to delegate"
	// UndelegateAmountMessage -
	UndelegateAmountMessage = "Send amount to undelegate"

	// AddNodeMessage -
	AddNodeMessage = "Send the node's validatorKey.pem"
)
