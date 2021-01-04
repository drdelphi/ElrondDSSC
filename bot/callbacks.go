package bot

import (
	"strconv"
	"strings"

	"github.com/DrDelphi/ElrondDSSC/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) callbackQueryReceived(cb *tgbotapi.CallbackQuery) {
	b.tgBot.AnswerCallbackQuery(tgbotapi.NewCallback(cb.ID, "Ok"))

	user := b.database.GetUserByTgID(int64(cb.From.ID))
	name := utils.FormatTgUser(cb.From)

	if user == nil {
		log.Warn("callback received from unknown user", "callback", cb.Data, "user", name)
		return
	}

	if cb.Data == "About" {
		b.sendMessage(user.TgID, utils.AboutMessage)
	}

	if cb.Data == "MainHelp" {
		b.sendMessage(user.TgID, utils.MainHelp)
	}

	if cb.Data == "MyWalletsHelp" {
		b.sendMessage(user.TgID, utils.MyWalletsHelp)
	}

	if cb.Data == "MyWallets" {
		b.walletsMenu(user)
	}

	if cb.Data == "AddWallet" {
		msg := tgbotapi.NewMessage(int64(cb.From.ID), utils.AddWalletMessage)
		msg.ReplyMarkup = tgbotapi.ForceReply{
			ForceReply: true,
			Selective:  false,
		}
		b.tgBot.Send(msg)
	}

	if cb.Data == "Balances" {
		b.sendBalances(user)
	}

	if cb.Data == "AdminMenu" && user.TgID == b.owner {
		b.adminMenu(user)
	}

	if cb.Data == "NodesMenu" && user.TgID == b.owner {
		b.nodesMenu(user)
	}

	if cb.Data == "SetOwnerAddress" && user.TgID == b.owner {
		ownerAddress := b.database.GetOwnerAddress()
		if ownerAddress != "" {
			b.sendMessage(user.TgID, "Old address: "+ownerAddress)
		}
		msg := tgbotapi.NewMessage(int64(cb.From.ID), utils.SetOwnerAddressMessage)
		msg.ReplyMarkup = tgbotapi.ForceReply{
			ForceReply: true,
			Selective:  false,
		}
		b.tgBot.Send(msg)
	}

	if cb.Data == "CreateDSSC" && user.TgID == b.owner {
		if utils.ContractAddress != "" {
			b.sendMessage(user.TgID, "⭕️ Contract already created")
			return
		}

		privateKey := b.database.GetOwnerPrivateKey()
		if privateKey == "" {
			b.sendMessage(user.TgID, "⭕️ Owner private key not set. You have to create the contract manually")
			return
		}

		txHash, err := b.networkManager.CreateDSSC(privateKey)
		if err == nil {
			b.sendMessage(user.TgID, "✅ Create DSSC transaction sent. Hash: "+txHash)
		} else {
			b.sendMessage(user.TgID, "⭕️ Failed to send create DSSC transaction: "+err.Error())
			return
		}
	}

	if cb.Data == "ContractInfo" {
		b.sendContractInfo(user)
	}

	if cb.Data == "Back" {
		b.mainMenu(user)
	}

	if cb.Data == "Delegate" {
		text := utils.DelegateAmountMessage
		msg := tgbotapi.NewMessage(user.TgID, text)
		msg.ReplyMarkup = tgbotapi.ForceReply{
			ForceReply: true,
			Selective:  false,
		}
		b.tgBot.Send(msg)
	}

	if cb.Data == "Undelegate" {
		text := utils.UndelegateAmountMessage
		msg := tgbotapi.NewMessage(user.TgID, text)
		msg.ReplyMarkup = tgbotapi.ForceReply{
			ForceReply: true,
			Selective:  false,
		}
		b.tgBot.Send(msg)
	}

	if strings.HasPrefix(cb.Data, ":") {
		params := strings.Split(cb.Data, "_")
		params[0] = strings.TrimPrefix(params[0], ":")

		if params[0] == "RemoveWallet" && len(params) == 2 {
			id, _ := strconv.ParseUint(params[1], 10, 32)
			for i, w := range user.Wallets {
				if w.ID != id {
					continue
				}

				err := b.database.RemoveWallet(id)
				if err != nil {
					b.sendMessage(user.TgID, "⭕️ Error removing wallet from database")
					b.reportError("⭕️ Error removing wallet from database")
					return
				}

				user.Wallets = append(user.Wallets[:i], user.Wallets[i+1:]...)
				b.sendMessage(user.TgID, "🗑 Wallet removed")

				return
			}

			b.sendMessage(user.TgID, "⭕️ Wallet not found")
		}
	}

	if cb.Data == "MyNodes" && user.TgID == b.owner {
		b.sendNodes(user)
	}

	if cb.Data == "AddNode" && user.TgID == b.owner {
		if utils.ContractAddress == "" {
			b.sendMessage(user.TgID, "⭕️ Contract Address not found")
			return
		}

		text := utils.AddNodeMessage
		msg := tgbotapi.NewMessage(user.TgID, text)
		msg.ReplyMarkup = tgbotapi.ForceReply{
			ForceReply: true,
			Selective:  false,
		}
		b.tgBot.Send(msg)
	}
}
