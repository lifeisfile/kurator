package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

type Course struct {
	ShortName string `json:"short_name"`
	FullName  string `json:"full_name"`
	Author    string `json:"author"`
	URL       string `json:"url"`
	IsPaid    bool   `json:"is_paid"`
	IsDraft   bool   `json:"is_draft"`
}

func ListCourses(c *cli.Context) error {
	authCompleted, token := CheckAuthCompleted()
	if !authCompleted {
		return fmt.Errorf("authentication not completed")
	}

	url := targetURL + "/course_list"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var courses []Course
	err = json.Unmarshal(body, &courses)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Short Name", "Full Name", "Author", "URL", "Payment", "Is Draft"})

	for _, course := range courses {
		table.Append([]string{
			course.ShortName,
			course.FullName,
			course.Author,
			course.URL,
			fmt.Sprintf("%t", course.IsPaid),
			fmt.Sprintf("%t", course.IsDraft),
		})
	}

	table.Render()

	return nil
}

func sendToken(conn *websocket.Conn, token string) error {

	return conn.WriteMessage(websocket.TextMessage, []byte(token))
}

func handleServerMessage(conn *websocket.Conn, message []byte, isDev bool) {
	kr := KuratorRequest{}
	err := json.Unmarshal(message, &kr)
	if err == nil {
		// it is JSON
		if kr.IsDev && !isDev {
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Run dev commands on non-dev client"}`))
			fmt.Println("Refused to run local command")
			return
		}
		kresp := KuratorResponse{
			SeqID: kr.SeqID,
		}
		switch kr.Type {
		case "command":
			var cmd *exec.Cmd

			switch os := runtime.GOOS; os {
			case "linux", "darwin":
				cmd = exec.Command("bash", "-c", kr.Payload)
			case "windows":
				cmd = exec.Command("powershell", "-Command", kr.Payload)
			default:
				fmt.Printf("Unsupported operating system: %s", os)
			}
			fmt.Println("Command run", kr.Payload)
			output, err := cmd.CombinedOutput()
			exitCode := cmd.ProcessState.ExitCode()
			if err != nil {
				fmt.Println("Failed to run command:", err)
				fmt.Println(string(output))
			}
			kresp.CommandOutput = string(output)
			kresp.CommandExitCode = exitCode
		case "contains":
			if len(kr.Files) != 0 {
				content, err := ioutil.ReadFile(kr.Files[0])
				if err == nil {
					kresp.BoolResponse = strings.Contains(string(content), strings.TrimSpace(kr.Payload))
					kresp.CommandOutput = string(content)
				}
			}
		default:
			kresp.CommandOutput = "TYPE_NOT_SUPPORTED:" + version
		}
		dataJson, _ := json.Marshal(kresp)

		conn.WriteMessage(websocket.TextMessage, dataJson)
	}
}

func StartCourse(c *cli.Context) error {
	isDev := c.Bool("dev")

	authCompleted, token := CheckAuthCompleted()
	if !authCompleted {
		return fmt.Errorf("authentication not completed")
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	var conn *websocket.Conn
	go func() {
		// Wait for the interrupt signal
		<-interrupt
		fmt.Println("\nInterrupt signal received. Exiting...")
		if conn != nil {
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		}

		os.Exit(0)
	}()

	u := url.URL{Scheme: "wss", Host: "api.lifeisfile.com", Path: "/ws"}
	log.Printf("Connecting to %s", u.String())

	done := make(chan struct{})

	for {

		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Println("Connection failed. Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
			continue
		}

		// Send the token as the first message
		err = sendToken(conn, token)
		if err != nil {
			conn.Close()
			log.Println("Connection failed. Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("Connected!")

		// Start a goroutine to handle incoming messages from the server
		go func() {
			defer close(done)
			for {
				if conn != nil {
					_, message, err := conn.ReadMessage()
					if err != nil {
						if websocket.IsCloseError(err, websocket.CloseUnsupportedData) {
							log.Printf("error: %v", err)
							os.Exit(0)
						}
						if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
							log.Printf("error: %v", err)
							conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
							if err != nil {
								log.Println("Connection failed. Retrying in 5 seconds...")
								time.Sleep(5 * time.Second)
								continue
							} else {
								sendToken(conn, token)
							}
						}
						time.Sleep(5 * time.Second)
						continue
					}
					log.Printf("Received message from server: %s\n", message)

					handleServerMessage(conn, message, isDev)
				} else {
					conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
					if err != nil {
						log.Println("Connection failed. Retrying in 5 seconds...")
						time.Sleep(5 * time.Second)
						continue
					} else {
						sendToken(conn, token)
					}
				}
			}
		}()

		for {
			select {
			case <-done:
				return nil
			}
		}
	}

}
