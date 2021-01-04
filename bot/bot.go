package bot

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/DrDelphi/ElrondDSSC/data"
	"github.com/DrDelphi/ElrondDSSC/db"
	"github.com/DrDelphi/ElrondDSSC/network"
	"github.com/DrDelphi/ElrondDSSC/utils"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk/erdgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var log = logger.GetOrCreate("bot")

// Bot - holds the required fields of the bot application
type Bot struct {
	tgBot          *tgbotapi.BotAPI
	owner          int64
	walletHook     string
	database       *db.Database
	networkManager *network.NetworkManager
}

// NewBot - creates a new Bot object
func NewBot(cfg *data.AppConfig, database *db.Database, networkManager *network.NetworkManager) (*Bot, error) {
	tgBot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Error("can not create telegram bot", "error", err)
		return nil, err
	}

	telegramBot := &Bot{
		tgBot:          tgBot,
		owner:          cfg.BotOwner,
		walletHook:     cfg.WalletHook,
		database:       database,
		networkManager: networkManager,
	}

	return telegramBot, nil
}

// StartTasks - starts bot's tasks
func (b *Bot) StartTasks() {
	go func() {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates, err := b.tgBot.GetUpdatesChan(u)
		if err != nil {
			log.Error("can not get Telegram bot updates", "error", err)
			panic(err)
		}
		for update := range updates {
			if update.Message != nil {
				if update.Message.Chat.IsPrivate() {
					if update.Message.IsCommand() {
						b.privateCommandReceived(update.Message)
						continue
					}
					if update.Message.ReplyToMessage != nil {
						b.privateReplyReceived(update.Message)
						continue
					}
					if update.Message.Document != nil {
						_, _ = b.downloadFile(update.Message)
						continue
					}
				}
			}
			if update.CallbackQuery != nil {
				b.callbackQueryReceived(update.CallbackQuery)
			}
		}
	}()

	// read the DSSC address
	go func() {
		oldAddress := b.database.GetOwnerAddress()
		for {
			time.Sleep(time.Second * 10)
			if oldAddress != b.database.GetOwnerAddress() || utils.ContractAddress == "" {
				address := b.database.GetOwnerAddress()
				utils.ContractAddress = ""
				txs, err := b.networkManager.GetLastTxs("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqylllslmq6y6", 3000, "receiver")
				if err != nil {
					continue
				}

				for _, tx := range txs {
					if tx.Sender == address && tx.Status == "success" {
						for _, scr := range tx.SmartContractResults {
							if len(scr.Data) != 70 || !strings.HasPrefix(string(scr.Data), "@6f6b@") { // @ok@
								continue
							}
							hexAddress := strings.TrimPrefix(string(scr.Data), "@6f6b@")
							hexBytes, _ := hex.DecodeString(hexAddress)
							utils.ContractAddress, err = erdgo.PubkeyToBech32(hexBytes)
							if err == nil {
								break
							}
						}
						break
					}
				}

				oldAddress = b.database.GetOwnerAddress()
			}
		}
	}()
}

func (b *Bot) reportError(text string) {
	msg := tgbotapi.NewMessage(b.owner, "⛔️ "+text)
	b.tgBot.Send(msg)
}

func (b *Bot) sendMessage(userID int64, text string) {
	msg := tgbotapi.NewMessage(userID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	b.tgBot.Send(msg)
}

func (b *Bot) sendBalances(user *data.User) {
	b.sendMessage(user.TgID, "`Balances`")

	if len(user.Wallets) == 0 {
		b.sendMessage(user.TgID, "⭕️ No wallets added")
		return
	}

	ownerAddress := b.database.GetOwnerAddress()
	if ownerAddress == "" {
		b.sendMessage(user.TgID, "⭕️ The owner didn't set up the DSSC yet")
		return
	}

	for i, w := range user.Wallets {
		account, err := b.networkManager.Proxy.GetAccount(w.Address)
		text := fmt.Sprintf("`Wallet %v/%v`", i+1, len(user.Wallets))
		if err == nil {
			balance, err := account.GetBalance(18)
			if err == nil {
				text += fmt.Sprintf("\n\r`Balance:` %.4f eGLD", balance)
			} else {
				text += "\n\r❌ Balance error"
			}
		} else {
			text += "\n\r❌ Balance error"
		}

		keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow())

		activeStake, err := b.networkManager.GetUserActiveStake(w.Address)
		if err == nil && activeStake != nil {
			fActiveStake, _ := activeStake.Float64()
			if fActiveStake > 0 {
				text += fmt.Sprintf("\n\r`Delegated:` %.4f eGLD", fActiveStake)
			}
		}

		unstaked, err := b.networkManager.GetUserUnStakedValue(w.Address)
		if err == nil && unstaked != nil {
			fUnstaked, _ := unstaked.Float64()
			if fUnstaked > 0 {
				text += fmt.Sprintf("\n\r`Undelegated:` %.4f eGLD", fUnstaked)
				list, err := b.networkManager.GetUserUnDelegatedList(w.Address)
				if err == nil {
					for i := 0; i < len(list); i += 2 {
						iAmount := big.NewInt(0).SetBytes(list[i])
						fAmount := big.NewFloat(0).SetInt(iAmount)
						fAmount.Quo(fAmount, b.networkManager.GetDenominator())
						amount, _ := fAmount.Float64()

						iRounds := big.NewInt(0).SetBytes(list[i+1])
						seconds := iRounds.Uint64() * 6 // TODO: get the round duration from network config
						text += fmt.Sprintf("\n\r    - %.4f eGLD (ETA: %v:%02v:%02v)", amount, seconds/3600, seconds/60%60, seconds%60)
					}
				}
			}
		}

		unbondable, err := b.networkManager.GetUserUnBondable(w.Address)
		if err == nil && unbondable != nil {
			fUnbondable, _ := unbondable.Float64()
			if fUnbondable > 0 {
				text += fmt.Sprintf("\n\r`Can withdraw:` %.4f eGLD", fUnbondable)
			}
		}

		claimable, err := b.networkManager.GetClaimableRewards(w.Address)
		if err == nil && claimable != nil {
			fClaimable, _ := claimable.Float64()
			if fClaimable > 0 {
				text += fmt.Sprintf("\n\r`Claimable rewards:` %.4f eGLD", fClaimable)
			}
		}

		keyboard.InlineKeyboard[0] = append(keyboard.InlineKeyboard[0],
			tgbotapi.NewInlineKeyboardButtonData("🗑 Remove", fmt.Sprintf(":RemoveWallet_%v", w.ID)))

		msg := tgbotapi.NewMessage(user.TgID, text)
		msg.ReplyMarkup = keyboard
		msg.ParseMode = tgbotapi.ModeMarkdown
		b.tgBot.Send(msg)
	}
}

func (b *Bot) sendContractInfo(user *data.User) {
	b.sendMessage(user.TgID, "`Contract Info`")

	ownerAddress := b.database.GetOwnerAddress()
	if ownerAddress == "" {
		b.sendMessage(user.TgID, "⭕️ The owner didn't set up the DSSC yet")
		return
	}

	if utils.ContractAddress == "" {
		b.sendMessage(user.TgID, "⭕️ Contract Address not found")
		return
	}

	text := fmt.Sprintf("`Contract address`: %s", utils.ContractAddress)

	info, err := b.networkManager.GetContractInfo(utils.ContractAddress)
	if err == nil {
		text += fmt.Sprintf("\n\r`Service fee:` %.2f%%", info.ServiceFee)
		if info.ChangeableServiceFee {
			text += " (changeable)"
		}
		if info.WithDelegationCap {
			text += fmt.Sprintf("\n\r`Max delegation cap:` %v eGLD", uint64(info.MaxDelegationCap))
		}
		text += fmt.Sprintf("\n\r`Initial owner funds:` %v eGLD", uint64(info.InitialOwnerFunds))
		text += fmt.Sprintf("\n\r`Unbond period:` %v", info.UnBondPeriod)
		text += fmt.Sprintf("\n\r`Automatic activation:` %v", info.AutomaticActivation)
		text += fmt.Sprintf("\n\r`Created at nonce:` %v", info.CreatedNonce)
	}

	if user.TgID != b.owner {
		b.sendMessage(user.TgID, text)
		return
	}

	text += "\n\r"

	numNodes, err := b.networkManager.GetNumNodes()
	if err == nil {
		text += fmt.Sprintf("\n\r`Nodes:` %v", numNodes)
	}

	numUsers, err := b.networkManager.GetNumUsers()
	if err == nil {
		text += fmt.Sprintf("\n\r`Delegators:` %v", numUsers)
	}

	totalActiveStake, err := b.networkManager.GetTotalActiveStake()
	if err == nil {
		fTotalActiveStake, _ := totalActiveStake.Float64()
		text += fmt.Sprintf("\n\r`Total active stake:` %.4f eGLD", fTotalActiveStake)
	}

	totalCumulatedRewards, err := b.networkManager.GetTotalCumulatedRewards()
	if err == nil {
		fTotalCumulatedRewards, _ := totalCumulatedRewards.Float64()
		text += fmt.Sprintf("\n\r`Total cumulated rewards:` %.4f eGLD", fTotalCumulatedRewards)
	}

	totalUnStaked, err := b.networkManager.GetTotalUnStaked()
	if err == nil {
		fTotalUnStaked, _ := totalUnStaked.Float64()
		text += fmt.Sprintf("\n\r`Total unstaked:` %.4f eGLD", fTotalUnStaked)
	}

	totalUnStakedFromNodes, err := b.networkManager.GetTotalUnStakedFromNodes()
	if err == nil {
		fTotalUnStakedFromNodes, _ := totalUnStakedFromNodes.Float64()
		text += fmt.Sprintf("\n\r`Total unstaked from nodes:` %.4f eGLD", fTotalUnStakedFromNodes)
	}

	totalUnBondedFromNodes, err := b.networkManager.GetTotalUnBondedFromNodes()
	if err == nil {
		fTotalUnBondedFromNodes, _ := totalUnBondedFromNodes.Float64()
		text += fmt.Sprintf("\n\r`Total unbonded from nodes:` %.4f eGLD", fTotalUnBondedFromNodes)
	}

	b.sendMessage(user.TgID, text)
}

func (b *Bot) sendNodes(user *data.User) {
	if utils.ContractAddress == "" {
		b.sendMessage(user.TgID, "⭕️ Contract Address not found")
		return
	}

	list, err := b.networkManager.GetAllNodeStates()
	if err != nil {
		b.sendMessage(user.TgID, "⭕️ Can not get all nodes states")
		return
	}

	n := len(list) / 2
	for i := 0; i < len(list); i += 2 {
		state := string(list[i])
		key := hex.EncodeToString(list[i+1])
		text := fmt.Sprintf("`Node %v/%v`\n\r`Key:` %s\n\r`State:` %s", i/2+1, n, key, state)

		stakeURL := fmt.Sprintf("%s/hook/transaction?receiver=%s&value=0&gasLimit=12000000&data=stakeNodes@%s&callbackUrl=none",
			b.walletHook, utils.ContractAddress, key)
		unStakeURL := fmt.Sprintf("%s/hook/transaction?receiver=%s&value=0&gasLimit=12000000&data=unStakeNodes@%s&callbackUrl=none",
			b.walletHook, utils.ContractAddress, key)
		unBondURL := fmt.Sprintf("%s/hook/transaction?receiver=%s&value=0&gasLimit=12000000&data=unBondNodes@%s&callbackUrl=none",
			b.walletHook, utils.ContractAddress, key)
		reStakeUnStakedURL := fmt.Sprintf("%s/hook/transaction?receiver=%s&value=0&gasLimit=12000000&data=reStakeUnStakedNodes@%s&callbackUrl=none",
			b.walletHook, utils.ContractAddress, key)
		unJailURL := fmt.Sprintf("%s/hook/transaction?receiver=%s&value=0&gasLimit=12000000&data=unJailNodes@%s&callbackUrl=none",
			b.walletHook, utils.ContractAddress, key)
		removeURL := fmt.Sprintf("%s/hook/transaction?receiver=%s&value=0&gasLimit=12000000&data=removeNodes@%s&callbackUrl=none",
			b.walletHook, utils.ContractAddress, key)

		msg := tgbotapi.NewMessage(user.TgID, text)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("Stake", stakeURL),
				tgbotapi.NewInlineKeyboardButtonURL("Unstake", unStakeURL),
				tgbotapi.NewInlineKeyboardButtonURL("Unbond", unBondURL),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("Restake", reStakeUnStakedURL),
				tgbotapi.NewInlineKeyboardButtonURL("Unjail", unJailURL),
				tgbotapi.NewInlineKeyboardButtonURL("Remove", removeURL),
			),
		)
		b.tgBot.Send(msg)
	}
}
