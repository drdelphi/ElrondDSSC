package data

// AppConfig holds the application configuration read from config.json
type AppConfig struct {
	BotToken     string `json:"botToken"`
	BotOwner     int64  `json:"botOwner"`
	DatabasePath string `json:"databasePath"`
	NetworkAPI   string `json:"networkAPI"`
	NetworkProxy string `json:"networkProxy"`
	MetaObserver string `json:"metaObserver"`
	WalletHook   string `json:"walletHook"`
}
