package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"syscall"

	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
)

type LoginResponse struct {
	Token string `json:"token"`
}

func LoginUserCLI(c *cli.Context) error {
	email := c.String("email")
	password := getPassword()

	if email == "" {
		return fmt.Errorf("email is required")
	}

	token, err := LoginUser(email, password)
	if err != nil {
		return err
	}

	err = saveToken(token)
	if err != nil {
		return err
	}

	fmt.Printf("Login successful. Token: %s\n", token)
	return nil
}

func SignUpUserCLI(c *cli.Context) error {
	email := c.String("email")
	name := c.String("name")

	if email == "" {
		return fmt.Errorf("email is required")
	}

	err := SignupUser(email, name)
	if err != nil {
		return err
	}

	return nil
}

func LoginUser(email, password string) (string, error) {
	url := targetURL + "/auth_user"

	payload := map[string]string{
		"email":    email,
		"password": password,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("server returned status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var loginResponse LoginResponse
	err = json.Unmarshal(body, &loginResponse)
	if err != nil {
		return "", err
	}

	return loginResponse.Token, nil
}

func SignupUser(email, name string) error {
	url := targetURL + "/users"

	payload := struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}{
		Email: email,
		Name:  name,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to signup user. Status: %s. Response: %s", resp.Status, string(body))
	}

	fmt.Println("User signup successful. Your password is sent to your email.")
	return nil
}

func saveToken(token string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".config", "kurator")
	err = os.MkdirAll(configDir, 0700)
	if err != nil {
		return err
	}

	filePath := filepath.Join(configDir, "token")
	err = ioutil.WriteFile(filePath, []byte(token), 0600)
	if err != nil {
		return err
	}

	return nil
}

func getPassword() string {
	fmt.Print("Enter password: ")
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return string(bytePassword)
}
