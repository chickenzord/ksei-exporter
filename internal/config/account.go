package config

import (
	"strings"
)

type Account struct {
	Username string
	Password string
}

type Accounts []Account

func (a *Accounts) Decode(val string) error {
	accounts := accountsFromString(val)

	*a = accounts

	return nil
}

func accountsFromString(str string) []Account {
	accounts := []Account{}

	pairs := strings.Fields(str)

	for _, pair := range pairs {
		frags := strings.SplitN(pair, ":", 2)

		accounts = append(accounts, Account{
			Username: frags[0],
			Password: frags[1],
		})
	}

	return accounts
}
