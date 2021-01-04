package db

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/DrDelphi/ElrondDSSC/data"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

var log = logger.GetOrCreate("database")

// Database - holds the required fields of a database
type Database struct {
	path  string
	sqldb *sql.DB

	ownerAddress    string
	ownerPrivateKey string

	users    map[int64]*data.User
	usersMut sync.Mutex
}

// NewDatabase - creates a new Database object
func NewDatabase(databasePath string) (*Database, error) {
	sqldb, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		log.Error("can not open database", "error", err)
		return nil, err
	}

	db := &Database{
		path:  databasePath,
		sqldb: sqldb,
		users: make(map[int64]*data.User),
	}

	err = db.getSettings()
	if err != nil {
		log.Error("can not read settings from database", "error", err)
		_ = db.sqldb.Close()
		return nil, err
	}

	err = db.getUsers()
	if err != nil {
		log.Error("can not read users from database", "error", err)
		_ = db.sqldb.Close()
		return nil, err
	}

	return db, nil
}

// getSettings - reads the settings from the database
// it is called by NewDatabase
func (d *Database) getSettings() error {
	sql := "select * from Settings"
	row, err := d.sqldb.Query(sql)
	if err != nil {
		return err
	}

	defer row.Close()
	var (
		address    string
		privateKey string
	)
	for row.Next() {
		err = row.Scan(&address, &privateKey)
		if err != nil {
			log.Warn("can not read settings row from database", "error", err)
			continue
		}

		d.ownerAddress = address
		d.ownerPrivateKey = privateKey
		break
	}

	return nil
}

// getUsers - reads the users from the database
// it is called by NewDatabase
func (d *Database) getUsers() error {
	sql := "select * from Users"
	row, err := d.sqldb.Query(sql)
	if err != nil {
		return err
	}

	defer row.Close()
	var (
		id      uint64
		tgID    int64
		tgUser  string
		tgFirst string
		tgLast  string
	)
	for row.Next() {
		err = row.Scan(&id, &tgID, &tgUser, &tgFirst, &tgLast)
		if err != nil {
			log.Warn("can not read user row from database", "error", err)
			continue
		}

		first, err := base64.StdEncoding.DecodeString(tgFirst)
		if err != nil {
			first = []byte(tgFirst)
		}

		last, err := base64.StdEncoding.DecodeString(tgLast)
		if err != nil {
			last = []byte(tgLast)
		}

		user := &data.User{
			ID:      id,
			TgID:    tgID,
			TgUser:  tgUser,
			TgFirst: string(first),
			TgLast:  string(last),
		}

		wallets, err := d.getUserWallets(id)
		if err != nil {
			log.Warn("can not read user wallets", "error", err, "user", id)
			continue
		}

		user.Wallets = make([]*data.UserWallet, len(wallets))
		copy(user.Wallets, wallets)

		d.usersMut.Lock()
		d.users[tgID] = user
		d.usersMut.Unlock()
	}

	return nil
}

// getUserWallets - gets a user's wallets from the database
// it is called by getUsers
func (d *Database) getUserWallets(userID uint64) ([]*data.UserWallet, error) {
	sql := fmt.Sprintf("select * from UserWallets where (UserID = %v) and (Deleted = 0)", userID)
	row, err := d.sqldb.Query(sql)
	if err != nil {
		return nil, err
	}

	defer row.Close()
	wallets := make([]*data.UserWallet, 0)
	var (
		id      uint64
		uID     uint64
		address string
		deleted int
	)
	for row.Next() {
		err = row.Scan(&id, &uID, &address, &deleted)
		if err != nil {
			log.Warn("can not read user wallet row", "error", err, "user", userID)
			continue
		}

		wallet := &data.UserWallet{
			ID:      id,
			UserID:  uID,
			Address: address,
		}
		wallets = append(wallets, wallet)
	}

	return wallets, nil
}

// AddUser - adds a telegram user to the database
func (d *Database) AddUser(user *tgbotapi.User) error {
	sql := "insert into Users(TgID, TgUser, TgFirst, TgLast) values (?, ?, ?, ?)"
	statement, err := d.sqldb.Prepare(sql)
	if err != nil {
		log.Error("error adding user in database", "error", err)
		return err
	}

	first := base64.StdEncoding.EncodeToString([]byte(user.FirstName))
	last := base64.StdEncoding.EncodeToString([]byte(user.LastName))

	res, err := statement.Exec(user.ID, user.UserName, first, last)
	if err != nil {
		log.Error("error adding user in database", "error", err)
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Error("error adding user in database", "error", err)
		return err
	}

	u := &data.User{
		ID:      uint64(id),
		TgID:    int64(user.ID),
		TgUser:  user.UserName,
		TgFirst: user.FirstName,
		TgLast:  user.LastName,
		Wallets: make([]*data.UserWallet, 0),
	}
	d.usersMut.Lock()
	d.users[u.TgID] = u
	d.usersMut.Unlock()

	return nil
}

// GetUserByTgID - returns a user by its Telegram ID
func (d *Database) GetUserByTgID(tgID int64) *data.User {
	d.usersMut.Lock()
	defer d.usersMut.Unlock()

	return d.users[tgID]
}

// GetUserByTgUser - same as GetUserByID with the addition that if the Telegram username,
// firstname or lastname have changed, it also updates them in the database
func (d *Database) GetUserByTgUser(tgUser *tgbotapi.User) *data.User {
	user := d.GetUserByTgID(int64(tgUser.ID))
	if user == nil {
		return nil
	}

	if user.TgUser != tgUser.UserName || user.TgFirst != tgUser.FirstName || user.TgLast != tgUser.LastName {
		user.TgUser = tgUser.UserName
		user.TgFirst = tgUser.FirstName
		user.TgLast = tgUser.LastName
		_ = d.updateUser(user)
	}

	return user
}

// GetUsers - returns all registered users
func (d *Database) GetUsers() map[int64]*data.User {
	m := make(map[int64]*data.User)

	d.usersMut.Lock()
	for k, v := range d.users {
		m[k] = v
	}
	d.usersMut.Unlock()

	return m
}

// updateUser - updates in the database a user's Telegram credentials
func (d *Database) updateUser(user *data.User) error {
	first := base64.StdEncoding.EncodeToString([]byte(user.TgFirst))
	last := base64.StdEncoding.EncodeToString([]byte(user.TgLast))

	sql := fmt.Sprintf("update Users set TgUser = ?, TgFirst = ?, TgLast = ? where ID = %v", user.ID)
	statement, err := d.sqldb.Prepare(sql)
	if err != nil {
		log.Warn("can not update user in database", "error", err, "user", user)
		return err
	}

	_, err = statement.Exec(user.TgUser, first, last)
	if err != nil {
		log.Warn("can not update user in database", "error", err, "user", user)
		return err
	}

	return nil
}

// GetOwnerAddress - returns the owner's address
func (d *Database) GetOwnerAddress() string {
	return d.ownerAddress
}

// GetOwnerPrivateKey - returns the owner's private key
func (d *Database) GetOwnerPrivateKey() string {
	return d.ownerPrivateKey
}

// SetOwnerAddress - saves the owner's address in database
func (d *Database) SetOwnerAddress(address string) error {
	sql := fmt.Sprintf("update Settings set OwnerAddress = ?, OwnerPrivateKey = ? where OwnerAddress = '%s'", d.ownerAddress)
	if d.ownerAddress == "" {
		sql = "insert into Settings(OwnerAddress, OwnerPrivateKey) values(?, ?)"
	}
	statement, err := d.sqldb.Prepare(sql)
	if err != nil {
		log.Error("can not set owner address in database", "error", err)
		return err
	}

	_, err = statement.Exec(address, d.ownerPrivateKey)
	if err != nil {
		log.Error("can not set owner address in database", "error", err)
		return err
	}

	d.ownerAddress = address

	return nil
}

// SetOwnerPrivateKey - saves the owner's private key in database
func (d *Database) SetOwnerPrivateKey(privateKey string) error {
	sql := fmt.Sprintf("update Settings set OwnerAddress = ?, OwnerPrivateKey = ? where OwnerAddress = '%s'", d.ownerAddress)
	if d.ownerAddress == "" {
		sql = "insert into Settings(OwnerAddress, OwnerPrivateKey) values(?, ?)"
	}
	statement, err := d.sqldb.Prepare(sql)
	if err != nil {
		log.Error("can not set owner private key in database", "error", err)
		return err
	}

	_, err = statement.Exec(d.ownerAddress, privateKey)
	if err != nil {
		log.Error("can not set owner private key in database", "error", err)
		return err
	}

	d.ownerPrivateKey = privateKey

	return nil
}

// AddUserWallet - adds a user wallet in the database
func (d *Database) AddUserWallet(user *data.User, address string) error {
	sql := "insert into UserWallets(UserID, Address) values(?, ?)"
	statement, err := d.sqldb.Prepare(sql)
	if err != nil {
		log.Error("can not add user wallet in database", "error", err)
		return err
	}

	res, err := statement.Exec(user.ID, address)
	if err != nil {
		log.Error("can not add user wallet in database", "error", err)
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Error("can not add user wallet in database", "error", err)
		return err
	}

	wallet := &data.UserWallet{
		ID:      uint64(id),
		UserID:  user.ID,
		Address: address,
	}
	user.Wallets = append(user.Wallets, wallet)

	return nil
}

// RemoveWallet - marks a user wallet as deleted in the database
func (d *Database) RemoveWallet(id uint64) error {
	sql := fmt.Sprintf("update UserWallets set Deleted = 1 where ID = %v", id)
	statement, err := d.sqldb.Prepare(sql)
	if err != nil {
		log.Error("can not remove user wallet from database", "error", err)
		return err
	}

	_, err = statement.Exec()
	if err != nil {
		log.Error("can not remove user wallet from database", "error", err)
		return err
	}

	return nil
}
