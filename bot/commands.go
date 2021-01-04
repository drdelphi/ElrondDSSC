package bot

import (
	"fmt"

	"github.com/DrDelphi/ElrondDSSC/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) privateCommandReceived(message *tgbotapi.Message) {
	cmd := message.Command()
	args := message.CommandArguments()
	name := utils.FormatTgUser(message.From)

	user := b.database.GetUserByTgUser(message.From)
	log.Info("command received", "command", cmd, "args", args, "user", name)
	if user == nil {
		err := b.database.AddUser(message.From)
		if err == nil {
			log.Info("new user registered", "user", name)
		} else {
			b.reportError(fmt.Sprintf("Can not add user '%s' in database. %s", name, err))
			return
		}

		user = b.database.GetUserByTgID(int64(message.From.ID))
		if user == nil {
			return
		}
	}

	if cmd == "start" {
		b.mainMenu(user)
	}
}
