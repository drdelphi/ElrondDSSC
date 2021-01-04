package bot

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) downloadFile(message *tgbotapi.Message) (string, error) {
	doc := message.Document
	if doc == nil {
		return "", errors.New("nil document object")
	}

	file, err := b.tgBot.GetFile(tgbotapi.FileConfig{FileID: doc.FileID})
	if err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("%v-%v-%s", message.From.ID, time.Now().Unix(), doc.FileName)
	url := file.Link(b.tgBot.Token)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	f, err := os.Create(fileName)
	if err != nil {
		return "", err
	}

	_, err = f.Write(data)
	if err != nil {
		return "", err
	}

	err = f.Close()
	if err != nil {
		return "", err
	}

	return fileName, nil
}
