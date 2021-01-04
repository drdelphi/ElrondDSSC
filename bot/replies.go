package bot

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/DrDelphi/ElrondDSSC/data"
	"github.com/DrDelphi/ElrondDSSC/utils"
	"github.com/ElrondNetwork/elrond-sdk/erdgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) privateReplyReceived(message *tgbotapi.Message) {
	user := b.database.GetUserByTgID(int64(message.From.ID))
	name := utils.FormatTgUser(message.From)
	log.Info("reply received", "reply to message", message.ReplyToMessage.Text, "message", message.Text, "user", name)

	if user == nil {
		log.Warn("reply received from unknown user", "reply", message.Text, "user", name)
		return
	}

	var err error
	fileName := ""
	if message.Document != nil {
		fileName, err = b.downloadFile(message)
		if err != nil {
			log.Error("error downloading file", "file", message.Document.FileName, "user", name, "error", err)
			b.reportError("error downloading file: " + err.Error())
		}
	}

	if message.ReplyToMessage.Text == utils.SetOwnerAddressMessage && user.TgID == b.owner {
		b.setOwnerAddress(message, user, fileName)
	}

	if message.ReplyToMessage.Text == utils.AddWalletMessage {
		if !erdgo.IsValidBech32Address(message.Text) {
			b.sendMessage(user.TgID, "⭕️ Invalid address")
			return
		}

		err = b.database.AddUserWallet(user, message.Text)
		if err == nil {
			b.sendMessage(user.TgID, "✅ Wallet added")
		} else {
			b.sendMessage(user.TgID, "⭕️ Error adding wallet in database")
		}
	}

	if message.ReplyToMessage.Text == utils.DelegateAmountMessage {
		amount, err := strconv.ParseFloat(message.Text, 32)
		if err != nil {
			b.sendMessage(user.TgID, "⭕️ Invalid amount")
			return
		}

		if amount < 10 {
			b.sendMessage(user.TgID, "⭕️ Minimum amount is 10 eGLD")
			return
		}

		fAmount := big.NewFloat(amount)
		fAmount.Mul(fAmount, b.networkManager.GetDenominator())
		iAmount, _ := fAmount.Int(nil)

		url := fmt.Sprintf("%s/hook/transaction?receiver=%s&value=%v&gasLimit=12000000&data=delegate&callbackUrl=none",
			b.walletHook, utils.ContractAddress, iAmount)
		text := fmt.Sprintf("%.4f eGLD", amount)

		msg := tgbotapi.NewMessage(user.TgID, "Delegate")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(text, url),
			),
		)
		b.tgBot.Send(msg)
	}

	if message.ReplyToMessage.Text == utils.UndelegateAmountMessage {
		amount, err := strconv.ParseFloat(message.Text, 32)
		if err != nil {
			b.sendMessage(user.TgID, "⭕️ Invalid amount")
			return
		}

		if amount < 10 {
			b.sendMessage(user.TgID, "⭕️ Minimum amount is 10 eGLD")
			return
		}

		fAmount := big.NewFloat(amount)
		fAmount.Mul(fAmount, b.networkManager.GetDenominator())
		iAmount, _ := fAmount.Int(nil)
		bytesAmount := iAmount.Bytes()
		strBytesAmount := hex.EncodeToString(bytesAmount)

		url := fmt.Sprintf("%s/hook/transaction?receiver=%s&value=0&gasLimit=12000000&data=unDelegate@%s&callbackUrl=none",
			b.walletHook, utils.ContractAddress, strBytesAmount)
		text := fmt.Sprintf("%.4f eGLD", amount)

		msg := tgbotapi.NewMessage(user.TgID, "Undelegate")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(text, url),
			),
		)
		b.tgBot.Send(msg)
	}

	if message.ReplyToMessage.Text == utils.ChangeServiceFeeMessage && user.TgID == b.owner {
		fee, err := strconv.ParseFloat(message.Text, 32)
		if err != nil || fee < 0 || fee > 100 {
			b.sendMessage(user.TgID, "⭕️ Invalid fee")
			return
		}

		iFee := uint64(fee * 100)
		strFee := hex.EncodeToString([]byte{byte(iFee / 256), byte(iFee % 256)})

		url := fmt.Sprintf("%s/hook/transaction?receiver=%s&value=0&gasLimit=6000000&data=changeServiceFee@%s&callbackUrl=none",
			b.walletHook, utils.ContractAddress, strFee)
		text := fmt.Sprintf("%.2f%%", fee)

		msg := tgbotapi.NewMessage(user.TgID, "Change service fee")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(text, url),
			),
		)
		b.tgBot.Send(msg)
	}

	if message.ReplyToMessage.Text == utils.ModifyDelegationCapMessage && user.TgID == b.owner {
		cap, err := strconv.ParseFloat(message.Text, 32)
		if err != nil {
			b.sendMessage(user.TgID, "⭕️ Invalid fee")
			return
		}

		fCap := big.NewFloat(cap)
		fCap.Mul(fCap, b.networkManager.GetDenominator())

		iCap, _ := fCap.Int(nil)
		strCap := hex.EncodeToString(iCap.Bytes())

		url := fmt.Sprintf("%s/hook/transaction?receiver=%s&value=0&gasLimit=6000000&data=modifyTotalDelegationCap@%s&callbackUrl=none",
			b.walletHook, utils.ContractAddress, strCap)
		text := fmt.Sprintf("%.2f eGLD", cap)

		msg := tgbotapi.NewMessage(user.TgID, "Modify delegation cap")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(text, url),
			),
		)
		b.tgBot.Send(msg)
	}

	if message.ReplyToMessage.Text == utils.AddNodeMessage && user.TgID == b.owner {
		b.addNode(message, user, fileName)
	}
}

func (b *Bot) addNode(message *tgbotapi.Message, user *data.User, fileName string) {
	if fileName == "" {
		b.sendMessage(user.TgID, "⭕️ No pem file received")
		return
	}

	privateKey, err := erdgo.LoadPrivateKeyFromPemFile(fileName)
	if err != nil {
		b.sendMessage(user.TgID, "⭕️ Invalid pem file")
		return
	}

	publicKey, err := utils.GetValidatorKeyFromPrivateKey(privateKey)
	if err != nil {
		b.sendMessage(user.TgID, "⭕️ Can not derive public key from private key")
		return
	}

	sig, err := utils.GetStakeSig(utils.ContractAddress, fileName)
	if err != nil {
		b.sendMessage(user.TgID, "⭕️ Error signing with BLS key")
		return
	}

	url := fmt.Sprintf("%s/hook/transaction?receiver=%s&value=0&gasLimit=6000000&data=addNodes@%s@%s&callbackUrl=none",
		b.walletHook, utils.ContractAddress, hex.EncodeToString(publicKey), hex.EncodeToString(sig))
	msg := tgbotapi.NewMessage(user.TgID, "Add node")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Send transaction", url),
		),
	)
	b.tgBot.Send(msg)
}

func (b *Bot) setOwnerAddress(message *tgbotapi.Message, user *data.User, fileName string) {
	var err error
	address := message.Text
	privateKey := make([]byte, 0)

	if fileName != "" {
		if strings.Contains(strings.ToLower(fileName), ".pem") {
			privateKey, err = erdgo.LoadPrivateKeyFromPemFile(fileName)
		} else {
			if strings.Contains(strings.ToLower(fileName), ".json") {
				privateKey, err = erdgo.LoadPrivateKeyFromJsonFile(fileName, message.Caption)
			} else {
				b.sendMessage(user.TgID, "⭕️ Unknown file type")
				return
			}
		}
		if err != nil {
			log.Error("error loading wallet file", "error", err, "file", fileName)
			b.sendMessage(user.TgID, "⭕️ Invalid file: "+err.Error())
			return
		}
	}

	if len(privateKey) > 0 {
		address, err = erdgo.GetAddressFromPrivateKey(privateKey)
		if err != nil {
			b.sendMessage(user.TgID, "⭕️ Invalid file")
			return
		}
	}

	if !erdgo.IsValidBech32Address(address) {
		b.sendMessage(user.TgID, "⭕️ Invalid address")
		return
	}

	err = b.database.SetOwnerAddress(address)
	if err == nil {
		b.sendMessage(user.TgID, "✅ Owner address updated")
	} else {
		b.sendMessage(user.TgID, "⭕️ Error setting owner address: "+err.Error())
	}

	if len(privateKey) > 0 {
		privateKeyStr := hex.EncodeToString(privateKey)
		err = b.database.SetOwnerPrivateKey(privateKeyStr)
		if err == nil {
			b.sendMessage(user.TgID, "✅ Owner private key updated")
		} else {
			b.sendMessage(user.TgID, "⭕️ Error setting owner private key: "+err.Error())
		}
	}
}
