package lib

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

var (
	//go:embed web
	res embed.FS
)

func extractEmbeddedFiles(embeddedFiles embed.FS, targetPath string) error {
	err := os.MkdirAll(targetPath, 0755)
	if err != nil {
		return err
	}

	err = fs.WalkDir(embeddedFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		content, err := fs.ReadFile(embeddedFiles, path)
		if err != nil {
			return err
		}

		targetFilePath := filepath.Join(targetPath, path)

		err = ioutil.WriteFile(targetFilePath, content, 0644)
		if err != nil {
			return err
		}

		fmt.Printf("Saved file: %s\n", targetFilePath)
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func extraMiddleware(handlerURL string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("handlerURL", handlerURL)
			return next(c)
		}
	}
}

func RunDevServer(c *cli.Context) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	webPath := path.Join(homeDir, ".config", "kurator")
	if _, err := os.Stat(webPath); os.IsNotExist(err) {
		err := os.MkdirAll(webPath, 0755)
		if err != nil {
			return err
		}
	}
	err = extractEmbeddedFiles(res, webPath)
	if err != nil {
		return err
	}

	courseName := c.String("course_name")
	mainJSFilePath := webPath + "/web/main_origin.js"
	modifiedJSFilePath := webPath + "/web/main.js"

	err = ReplaceText(mainJSFilePath, modifiedJSFilePath, "kubernetes", courseName)
	if err != nil {
		log.Fatal(err)
	}

	err = ReplaceText(modifiedJSFilePath, modifiedJSFilePath, "https://api.k8school.lifeisfile.com", "http://localhost:4321")
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	e.Use(extraMiddleware(c.String("handler_url")))

	e.Static("/", webPath+"/web")

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		if code == http.StatusNotFound {
			c.File(webPath + "/web/index.html")
		} else {
			e.DefaultHTTPErrorHandler(err, c)
		}
	}

	e.PUT("/auth_user", ProxyHandler("PUT", "/auth_user"))
	e.PUT("/logout_user", ProxyHandler("PUT", "/logout_user"))

	e.GET("/course/:name", GetCourse)
	e.GET("/course/:name/:taskNumber", GetCourseTask)
	e.POST("/baseHandler", BaseHandler)

	e.Logger.Fatal(e.Start(":4321"))
	return nil
}

func GetCourse(c echo.Context) error {
	name := c.Param("name")
	dat, err := ioutil.ReadFile(fmt.Sprintf("%s/base.yaml", name))
	if err != nil {
		fmt.Println(err)
	}

	ci := CourseInfo{}

	err = yaml.Unmarshal(dat, &ci)
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(fmt.Sprintf("%s/tasks", name))
	if err != nil {
		return err
	}
	var sortedFiles []string
	for _, f := range files {
		sortedFiles = append(sortedFiles, f.Name())
	}
	sort.Slice(sortedFiles, func(i, j int) bool {
		var prev, next int
		parts := strings.Split(sortedFiles[i], ".")
		if len(parts) > 0 {
			prev, _ = strconv.Atoi(parts[0])
		}
		parts = strings.Split(sortedFiles[j], ".")
		if len(parts) > 0 {
			next, _ = strconv.Atoi(parts[0])
		}
		return prev < next
	})
	var taskList []TaskListItem

	var completedTasks []string
	//TODO: implement list of completed tasks

	for n, filename := range sortedFiles {
		tsi := TaskShortInfo{}
		dat, err = ioutil.ReadFile(fmt.Sprintf("%s/tasks/%s", name, filename))
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(dat, &tsi)
		if err != nil {
			return err
		}
		locked := false

		isCompleted := false
		if StringSliceContains(completedTasks, tsi.TaskID) {
			isCompleted = true
		}

		taskList = append(taskList, TaskListItem{TaskTitle: tsi.TaskTitle, TaskID: tsi.TaskID, PositionNumber: n, IsLocked: locked, IsCompleted: isCompleted})
	}
	ci.TaskList = taskList
	return c.JSON(http.StatusOK, ci)
}

func GetCourseTask(c echo.Context) error {

	name := c.Param("name")
	taskNumber := c.Param("taskNumber")
	dat, err := ioutil.ReadFile(fmt.Sprintf("%s/tasks/%s.yaml", name, taskNumber))
	if err != nil {
		return err
	}
	ti := TaskInfo{}
	err = yaml.Unmarshal(dat, &ti)
	if err != nil {
		return err
	}
	for n, goal := range ti.Goals {
		for k, _ := range goal.Contents {
			ti.Goals[n].Contents[k].KuratorRequest.Payload = ""
			ti.Goals[n].Contents[k].KuratorRequest.APIVersion = ""
			ti.Goals[n].Contents[k].KuratorRequest.Type = ""
			ti.Goals[n].Contents[k].KuratorRequest.Command = ""
			ti.Goals[n].Contents[k].KuratorRequest.Files = []string{}
			ti.Goals[n].Contents[k].KuratorRequest.Args = []string{}
		}
	}

	//TODO: implement
	// resultJSON, err := storage.ReadObj(fmt.Sprintf("goal-list-%d", tkn.UserID))
	// if err != nil {
	// 	raven.CaptureError(err, nil)
	// 	return err
	// }
	// var completedGoals []string
	// if string(resultJSON) != "{}" {
	// 	err = json.Unmarshal(resultJSON, &completedGoals)
	// 	if err != nil {
	// 		raven.CaptureError(err, nil)
	// 	}
	// }
	// for _, goal := range completedGoals {
	// 	for n, stepGoal := range ti.Goals {
	// 		if goal == stepGoal.ID {
	// 			ti.Goals[n].IsCompleted = true
	// 		}
	// 	}
	// }

	return c.JSON(http.StatusOK, ti)
}

func BaseHandler(c echo.Context) error {

	handlerURL := c.Get("handlerURL").(string)
	token := c.Request().Header.Get("Token")

	authCompleted, devtoken := CheckAuthCompleted()
	if !authCompleted {
		return fmt.Errorf("authentication not completed")
	}

	rh := new(RequestHandler)
	if err := c.Bind(rh); err != nil {
		return err
	}

	userID, err := GetUserIDFromToken(token)
	if err != nil {
		// TODO: Handle error
		fmt.Println(err)
	} else {
		// Use the userID
		rh.UserID = userID
		fmt.Println(userID)
	}

	dat, err := ioutil.ReadFile(fmt.Sprintf("%s/tasks/%d.yaml", rh.CourseName, rh.TaskNumber))
	if err != nil {
		return err
	}
	ti := TaskInfo{}
	err = yaml.Unmarshal(dat, &ti)
	if err != nil {
		return err
	}

	for _, goal := range ti.Goals {

		//Check if current method has client websocket dependency call
		for _, content := range goal.Contents {
			if content.KuratorRequest.Payload != "" {
				if goal.RunHandler == rh.Method || (content.SourceHandler == rh.Method) {
					//TODO: Check for rh.CacheKey, use own cache to return the result to avoid hitting the client when result is cached on handler side
					// Client websocket call is required
					rand.Seed(time.Now().UnixNano())
					randomNumber := rand.Intn(10101010)

					kr := KuratorRequest{
						ApiVersion: content.KuratorRequest.APIVersion,
						Payload:    content.KuratorRequest.Payload,
						Type:       content.KuratorRequest.Type,
						Files:      content.KuratorRequest.Files,
						Args:       content.KuratorRequest.Args,
						UserID:     userID,
						CourseName: rh.CourseName,
						SeqID:      int64(randomNumber),
					}
					dataJson, _ := json.Marshal(kr)
					result, err := SendPostRequest(targetURL+"/run_kurator_request", string(dataJson), devtoken)
					if err != nil {
						res := CheckStatusResult{
							Status:   "user_error",
							Expected: "**Kurator** должен быть запущен: `kurator course start -dev`",
							Current:  "**Kurator** не подключен к серверу",
						}
						return c.JSON(http.StatusOK, res)
					}
					kresp := KuratorResponse{}
					err = json.Unmarshal([]byte(result), &kresp)
					if err != nil {
						return c.JSON(http.StatusInternalServerError, ResponseEmpty{})
					}
					rh.KuratorCommandExitCode = kresp.CommandExitCode
					rh.KuratorCommandOutput = kresp.CommandOutput
					rh.BoolResponse = kresp.BoolResponse
				}
			}
		}
	}

	//
	// rh.IsPaid = isPaid
	dataJson, _ := json.Marshal(rh)
	result, err := SendPostRequest(handlerURL, string(dataJson), "")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	// Try to get Status field to mark this goal as completed
	rfh := ResponseFromHandler{}
	err = json.Unmarshal([]byte(result), &rfh)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ResponseEmpty{})
	}
	if rfh.Status == "done" {
		//Mark this goal as completed for user
		fmt.Println("Task number is", rh.TaskNumber)
		//Check if all goals of current task are completed and if so mark task as completed as well
		//TODO: implement
		//UpdateUserGoals(rh)
	}
	return c.JSONBlob(http.StatusOK, []byte(result))
}

func CreateCourse(c *cli.Context) error {
	courseName := c.Args().First()
	if courseName == "" {
		return errors.New("missing course name")
	}

	validCourseName := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !validCourseName.MatchString(courseName) {
		return errors.New("invalid course name. It must contains only a-z, 0-9 and dash")
	}

	err := copyDirectory("assets/initial_yamls", courseName)
	if err != nil {
		return err
	}

	fmt.Printf("Initial structure for %s course is created in folder %s. You may now start editing these yamls and view them in web browser. Be sure to start server using:\n\n", courseName, courseName)
	fmt.Printf("kurator dev run-server --course_name %s --handler_url http://localhost:8888/courseHandler \n\n", courseName)

	return nil
}
