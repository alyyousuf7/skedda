package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// Config contains configurations for the CLI
type Config struct {
	Username string
	Password string
}

func loadConfig(configPath string) (Config, error) {
	raw, err := ioutil.ReadFile(path.Join(configPath, "config"))
	if err != nil {
		return Config{}, err
	}

	config := Config{}
	if err := json.Unmarshal(raw, &config); err != nil {
		return Config{}, err
	}

	return config, err
}

func saveConfig(configPath string, config Config) error {
	raw, err := json.Marshal(config)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configPath, 0777); err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(configPath, "config"), raw, 0777)
}

func readInput(prompt, defaultVal string, secret bool) string {
	reader := bufio.NewReader(os.Stdin)

	defaultValDisplay := defaultVal
	if secret {
		defaultValDisplay = "*****"
	}

	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValDisplay)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	var b []byte
	var v string
	var err error

	if secret {
		b, err = terminal.ReadPassword(int(syscall.Stdin))
		v = string(b)
	} else {
		v, err = reader.ReadString('\n')
	}
	if err != nil {
		fmt.Println("\n", err)
		return readInput(prompt, defaultVal, secret)
	}

	v = strings.TrimSpace(v)
	if v == "" {
		return defaultVal
	}
	return string(v)
}

func readCredentials(oldUsername, oldPassword string) (string, string) {
	return readInput("Username", oldUsername, false), readInput("Password", oldPassword, true)
}
