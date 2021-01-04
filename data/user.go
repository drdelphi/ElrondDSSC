package data

// User - holds the required fields of a user
type User struct {
	ID      uint64
	TgID    int64
	TgUser  string
	TgFirst string
	TgLast  string
	Wallets []*UserWallet

	LastMenuID int
}
