package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	Host string
	Port int
	User string
	Pass string
}

// Determine oxox configuration settings from environment variables.
func NewConfig() (*Config, error) {
	host, err := getHostname()
	if err != nil {
		return nil, err
	}

	port, err := getPort()
	if err != nil {
		return nil, err
	}

	pass, err := getPassword()
	if err != nil {
		return nil, err
	}

	return &Config{
		Host: host,
		Port: port,
		User: getUsername(),
		Pass: pass,
	}, nil
}

func getHostname() (string, error) {
	host, ok := os.LookupEnv("PROXMOX_HOST")

	if !ok {
		return "", errors.New("ensure PROXMOX_HOST environment variable is set")
	}

	return host, nil
}

func getPort() (int, error) {
	portStr, ok := os.LookupEnv("PROXMOX_PORT")

	port := 8006
	if ok {
		p, err := strconv.Atoi(portStr)
		if err != nil {
			return p, errors.New("ensure PROXMOX_PORT is a valid port number")
		}
		port = p
	}

	return port, nil
}

func getUsername() string {
	username, ok := os.LookupEnv("PROXMOX_USER")
	if !ok {
		username = "root@pam"
	}

	return username
}

func getPassword() (string, error) {
	password, ok := os.LookupEnv("PROXMOX_PASS")

	if !ok {
		return "", errors.New("ensure PROXMOX_PASS environment variable is set")
	}

	return password, nil
}
