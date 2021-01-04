package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-sdk/erdgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func GetHTTP(address string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func PostHTTP(address, body string) ([]byte, error) {
	resp, err := http.Post(address, "", strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	return ioutil.ReadAll(resp.Body)
}

func GetValidatorPubKeyFromPrivateKey(skBytes []byte) ([]byte, error) {
	gen := signing.NewKeyGenerator(mcl.NewSuiteBLS12())

	sk, err := gen.PrivateKeyFromByteArray(skBytes)
	if err != nil {
		return nil, err
	}

	return sk.GeneratePublic().ToByteArray()
}

func FormatTgUser(user *tgbotapi.User) string {
	name := fmt.Sprintf("%s %s [%v]", user.FirstName, user.LastName, user.ID)
	name = strings.TrimSpace(name)
	name = strings.Replace(name, "  ", " ", 1)
	if user.UserName != "" {
		name = fmt.Sprintf("@%s (%s)", user.UserName, name)
	}

	return name
}

func GetValidatorKeyFromPrivateKey(skBytes []byte) ([]byte, error) {
	gen := signing.NewKeyGenerator(mcl.NewSuiteBLS12())

	sk, err := gen.PrivateKeyFromByteArray(skBytes)
	if err != nil {
		return nil, err
	}

	return sk.GeneratePublic().ToByteArray()
}

func GetStakeSig(address string, validatorKeyFilename string) ([]byte, error) {
	gen := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	signer := singlesig.NewBlsSigner()

	skBytes, err := erdgo.LoadPrivateKeyFromPemFile(validatorKeyFilename)
	if err != nil {
		return nil, err
	}

	sk, err := gen.PrivateKeyFromByteArray(skBytes)
	if err != nil {
		return nil, err
	}

	addrBytes, err := erdgo.Bech32ToPubkey(address)
	if err != nil {
		return nil, err
	}

	return signer.Sign(sk, addrBytes)
}

func IsValidNodeKey(v string) bool {
	re := regexp.MustCompile("[0-9a-fA-F]{192}")
	return re.MatchString(v)
}
