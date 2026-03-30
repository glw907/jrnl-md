package main

import (
	"fmt"
	"os"

	"golang.org/x/term"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
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

func encryptJournal(journalPath, journalName string, cfg config.Config, configPath string) error {
	if journalEncrypted(cfg.Journals[journalName], cfg) {
		return fmt.Errorf("journal %q is already encrypted", journalName)
	}
	passphrase, err := promptNewPassphrase()
	if err != nil {
		return err
	}
	return reencryptJournal(journalPath, journalName, cfg, configPath, false, passphrase, true)
}

func decryptJournal(journalPath, journalName string, cfg config.Config, configPath string) error {
	if !journalEncrypted(cfg.Journals[journalName], cfg) {
		return fmt.Errorf("journal %q is not encrypted", journalName)
	}
	passphrase, err := promptPassphrase(fmt.Sprintf("Passphrase for journal %q: ", journalName))
	if err != nil {
		return err
	}
	return reencryptJournal(journalPath, journalName, cfg, configPath, true, passphrase, false)
}

func reencryptJournal(journalPath, journalName string, cfg config.Config, configPath string, fromEncrypt bool, passphrase string, toEncrypt bool) error {
	fj := journal.NewFolderJournal(journalPath, journalOptions(cfg, fromEncrypt, passphrase))
	if err := fj.Load(); err != nil {
		return fmt.Errorf("loading journal: %w", err)
	}

	oldFiles, err := fj.DayFiles()
	if err != nil {
		return fmt.Errorf("listing day files: %w", err)
	}

	fj.MarkAllModified()
	fj.SetEncryption(toEncrypt, passphrase)
	if err := fj.Save(); err != nil {
		return fmt.Errorf("saving journal: %w", err)
	}

	for _, f := range oldFiles {
		os.Remove(f)
	}

	jcfg := cfg.Journals[journalName]
	jcfg.Encrypt = boolPtr(toEncrypt)
	cfg.Journals[journalName] = jcfg
	if err := config.Save(cfg, configPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	verb := "encrypted"
	if !toEncrypt {
		verb = "decrypted"
	}
	fmt.Fprintf(os.Stderr, "Journal %q %s (%d files).\n", journalName, verb, len(oldFiles))
	return nil
}
