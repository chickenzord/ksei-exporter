package ksei

import (
	"os"
	"strings"
)

type Account struct {
	Username string
	Password string
}

func accountsFromEnv() []Account {
	accounts := []Account{}

	pairs := strings.Fields(os.Getenv("KSEI_ACCOUNTS"))

	for _, pair := range pairs {
		frags := strings.SplitN(pair, ":", 2)
		accounts = append(accounts, Account{
			Username: frags[0],
			Password: frags[1],
		})
	}

	return accounts
}
