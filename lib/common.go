package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

var targetURL = "https://api.lifeisfile.com"

func CheckAuthCompleted() (bool, string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, ""
	}

	tokenPath := filepath.Join(homeDir, ".config", "kurator", "token")
	_, err = os.Stat(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, ""
		}
		return false, ""
	}

	tokenBytes, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return false, ""
	}

	token := string(tokenBytes)
	return true, token
}

func ReplaceText(filePath, modifiedFilePath, oldText, newText string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	modifiedContent := strings.ReplaceAll(string(content), oldText, newText)

	err = ioutil.WriteFile(modifiedFilePath, []byte(modifiedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func ProxyHandler(method, path string) echo.HandlerFunc {
	return func(c echo.Context) error {
		req, err := http.NewRequest(method, targetURL+path, c.Request().Body)
		if err != nil {
			return err
		}

		req.Header = c.Request().Header

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if strings.Contains(path, "/auth_user") {
			var response struct {
				Token  string `json:"token"`
				UserID int64  `json:"user_id"`
			}
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			resp.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))

			err = json.Unmarshal(bodyBytes, &response)
			if err != nil {
				return err
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			cachePath := filepath.Join(homeDir, ".config", "kurator", "cache", "tokens")

			if _, err := os.Stat(cachePath); os.IsNotExist(err) {
				err := os.MkdirAll(cachePath, 0755)
				if err != nil {
					return err
				}
			}

			tokenFilePath := fmt.Sprintf("%s/%s", cachePath, response.Token)
			err = ioutil.WriteFile(tokenFilePath, []byte(strconv.FormatInt(response.UserID, 10)), 0644)
			if err != nil {
				return err
			}
		}

		for key, values := range resp.Header {
			for _, value := range values {
				c.Response().Header().Add(key, value)
			}
		}
		c.Response().WriteHeader(resp.StatusCode)

		_, err = io.Copy(c.Response().Writer, resp.Body)
		if err != nil {
			return err
		}

		return nil
	}
}

func SendPostRequest(url, data, token string) (string, error) {
	result := ""
	var jsonStr = []byte(data)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Token", token)

	fmt.Println(url, data)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		result = err.Error()
	} else {
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
			return result, fmt.Errorf("Backend replied with %d status", resp.StatusCode)
		}

		body, _ := ioutil.ReadAll(resp.Body)
		result = string(body)
		defer resp.Body.Close()
	}

	return result, err
}

func GetUserIDFromToken(token string) (int64, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}

	tokenFilePath := filepath.Join(homeDir, ".config", "kurator", "cache", "tokens", token)
	_, err = os.Stat(tokenFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, err
		}
		return 0, err
	}

	userIDBytes, err := ioutil.ReadFile(tokenFilePath)
	if err != nil {
		return 0, err
	}

	userID, err := strconv.ParseInt(string(userIDBytes), 10, 64)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func StringSliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func copyDirectory(src, dest string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", src)
	}

	err = os.Mkdir(dest, srcInfo.Mode())
	if err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	// Recursively copy the contents of the source directory to the destination directory
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			err = copyDirectory(srcPath, destPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, destPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}
