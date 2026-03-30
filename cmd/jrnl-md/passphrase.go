package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func promptPassphrase(msg string) (string, error) {
	fmt.Fprint(os.Stderr, msg)
	pass, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("reading passphrase: %w", err)
	}
	return string(pass), nil
}

func promptNewPassphrase() (string, error) {
	pass, err := promptPassphrase("New passphrase: ")
	if err != nil {
		return "", err
	}
	if pass == "" {
		return "", fmt.Errorf("passphrase cannot be empty")
	}
	confirm, err := promptPassphrase("Confirm passphrase: ")
	if err != nil {
		return "", err
	}
	if pass != confirm {
		return "", fmt.Errorf("passphrases do not match")
	}
	return pass, nil
}
