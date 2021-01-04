package bot

import (
	"fmt"

	"github.com/DrDelphi/ElrondDSSC/data"
	"github.com/DrDelphi/ElrondDSSC/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) mainMenu(user *data.User) {
	if user.LastMenuID > 0 {
		b.tgBot.DeleteMessage(tgbotapi.DeleteMessageConfig{
			ChatID:    user.TgID,
			MessageID: user.LastMenuID,
		})
	}

	withdrawURL := fmt.Sprintf("%s/hook/transaction?receiver=%s&value=0&gasLimit=12000000&data=withdraw&callbackUrl=none",
		b.walletHook, utils.ContractAddress)
	claimURL := fmt.Sprintf("%s/hook/transaction?receiver=%s&value=0&gasLimit=6000000&data=claimRewards&callbackUrl=none",
		b.walletHook, utils.ContractAddress)
	compoundURL := fmt.Sprintf("%s/hook/transaction?receiver=%s&value=0&gasLimit=12000000&data=reDelegateRewards&callbackUrl=none",
		b.walletHook, utils.ContractAddress)

	msg := tgbotapi.NewMessage(user.TgID, "`Main Menu`")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏦 My Wallets", "MyWallets"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🥩 Delegate", "Delegate"),
			tgbotapi.NewInlineKeyboardButtonData("🐖 Undelegate", "Undelegate"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🥓 Compound", compoundURL),
			tgbotapi.NewInlineKeyboardButtonURL("😋 Claim Rewards", claimURL),
			tgbotapi.NewInlineKeyboardButtonURL("🍽 Withdraw", withdrawURL),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("ℹ️ Contract Info", "ContractInfo"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📜 Help", "MainHelp"),
			tgbotapi.NewInlineKeyboardButtonData("❕ About", "About"),
		),
	)
	if user.TgID == b.owner {
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💻 Nodes management", "NodesMenu"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("👮‍♂️ Admin control panel", "AdminMenu"),
			),
		)
	}
	msg.ReplyMarkup = keyboard
	msg.ParseMode = tgbotapi.ModeMarkdown
	resp, _ := b.tgBot.Send(msg)
	user.LastMenuID = resp.MessageID
}

func (b *Bot) walletsMenu(user *data.User) {
	if user.LastMenuID > 0 {
		b.tgBot.DeleteMessage(tgbotapi.DeleteMessageConfig{
			ChatID:    user.TgID,
			MessageID: user.LastMenuID,
		})
	}
	msg := tgbotapi.NewMessage(user.TgID, "`My Wallets Menu`")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Add", "AddWallet"),
			tgbotapi.NewInlineKeyboardButtonData("💰 Balances", "Balances"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📜 Help", "MyWalletsHelp"),
			tgbotapi.NewInlineKeyboardButtonData("🚪 Back", "Back"),
		),
	)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = tgbotapi.ModeMarkdown
	resp, _ := b.tgBot.Send(msg)
	user.LastMenuID = resp.MessageID
}

func (b *Bot) adminMenu(user *data.User) {
	if user.LastMenuID > 0 {
		b.tgBot.DeleteMessage(tgbotapi.DeleteMessageConfig{
			ChatID:    user.TgID,
			MessageID: user.LastMenuID,
		})
	}
	autoActivateURL := fmt.Sprintf("%s/hook/transaction?receiver=%s&value=0&gasLimit=6000000&data=setAutomaticActivation@796573&callbackUrl=none",
		b.walletHook, utils.ContractAddress)
	msg := tgbotapi.NewMessage(user.TgID, "`Admin Control Panel`")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Set owner address", "SetOwnerAddress"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Create DSSC", "CreateDSSC"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Change Service Fee", "ChangeServiceFee"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Modify Delegation Cap", "ModifyDelegationCap"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Enable Automatic Activation", autoActivateURL),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚪 Back", "Back"),
		),
	)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = tgbotapi.ModeMarkdown
	resp, _ := b.tgBot.Send(msg)
	user.LastMenuID = resp.MessageID
}

func (b *Bot) nodesMenu(user *data.User) {
	if user.LastMenuID > 0 {
		b.tgBot.DeleteMessage(tgbotapi.DeleteMessageConfig{
			ChatID:    user.TgID,
			MessageID: user.LastMenuID,
		})
	}
	msg := tgbotapi.NewMessage(user.TgID, "`Nodes management`")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Add Node", "AddNode"),
			tgbotapi.NewInlineKeyboardButtonData("🖥 My Nodes", "MyNodes"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚪 Back", "Back"),
		),
	)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = tgbotapi.ModeMarkdown
	resp, _ := b.tgBot.Send(msg)
	user.LastMenuID = resp.MessageID
}
