package lib

import (
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

func GenerateHandlerCode(c *cli.Context) error {
	templatePath := c.String("template_path")
	outputPath := c.String("output_path")
	courseName := c.String("course_name")
	moduleName := c.String("module_name")
	taskInfos, err := parseYAMLFiles(path.Join(courseName, "tasks"))
	if err != nil {
		return err
	}

	err = GenerateCode(taskInfos, templatePath, outputPath, moduleName)
	return err
}

func parseYAMLFiles(dir string) ([]TaskInfo, error) {
	var taskInfos []TaskInfo

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".yaml" {
			continue
		}

		filePath := filepath.Join(dir, file.Name())
		taskInfo, err := parseYAMLFile(filePath)
		if err != nil {
			return nil, err
		}

		taskInfos = append(taskInfos, taskInfo)
	}

	return taskInfos, nil
}

func parseYAMLFile(filePath string) (TaskInfo, error) {
	var taskInfo TaskInfo

	yamlFile, err := os.Open(filePath)
	if err != nil {
		return taskInfo, err
	}
	defer yamlFile.Close()

	yamlData, err := ioutil.ReadAll(yamlFile)
	if err != nil {
		return taskInfo, err
	}

	err = yaml.Unmarshal(yamlData, &taskInfo)
	if err != nil {
		return taskInfo, err
	}

	return taskInfo, nil
}

func GenerateGenTasks(taskInfos []TaskInfo) []GenTask {
	var genTasks []GenTask

	for _, taskInfo := range taskInfos {
		genTask := GenTask{
			TaskID: taskInfo.TaskID,
			IsFree: taskInfo.IsFree,
			Methods: make([]struct {
				HandlerName     string
				HandlerFuncName string
				IsOutput        bool
			}, 0),
		}

		for _, goal := range taskInfo.Goals {
			if goal.StatusHandler != "" {
				genTask.Methods = append(genTask.Methods, struct {
					HandlerName     string
					HandlerFuncName string
					IsOutput        bool
				}{
					HandlerName:     goal.StatusHandler,
					HandlerFuncName: convertToCamelCase(goal.StatusHandler),
					IsOutput:        false,
				})
			}

			if goal.RunHandler != "" {
				genTask.Methods = append(genTask.Methods, struct {
					HandlerName     string
					HandlerFuncName string
					IsOutput        bool
				}{
					HandlerName:     goal.RunHandler,
					HandlerFuncName: convertToCamelCase(goal.RunHandler),
					IsOutput:        false,
				})
			}

			if goal.Contents != nil {
				for _, content := range goal.Contents {
					if content.SourceHandler != "" {
						genTask.Methods = append(genTask.Methods, struct {
							HandlerName     string
							HandlerFuncName string
							IsOutput        bool
						}{
							HandlerName:     content.SourceHandler,
							HandlerFuncName: convertToCamelCase(content.SourceHandler),
							IsOutput:        true,
						})
					}
				}
			}
		}

		genTasks = append(genTasks, genTask)
	}

	return genTasks
}

func convertToCamelCase(name string) string {
	words := strings.Split(name, "_")
	for i := range words {
		words[i] = strings.Title(words[i])
	}
	return strings.Join(words, "")
}

func GenerateCode(taskInfos []TaskInfo, templatesDir, outputDir, moduleName string) error {
	module := Module{
		ModuleName: moduleName,
		Tasks:      GenerateGenTasks(taskInfos),
	}

	err := processTemplates(module, templatesDir, outputDir)
	if err != nil {
		return err
	}

	return nil
}

func processTemplates(module Module, templatesDir, outputDir string) error {
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return err
	}

	err = filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(templatesDir, path)
		if err != nil {
			return err
		}

		outputPath := filepath.Join(outputDir, strings.TrimSuffix(relPath, ".tmpl"))
		outputDir := filepath.Dir(outputPath)

		if strings.HasPrefix(filepath.Base(path), "item.") {
			for _, task := range module.Tasks {
				taskOutputPath := filepath.Join(outputDir, task.TaskID+filepath.Ext(path))
				targetPath := strings.TrimSuffix(taskOutputPath, ".tmpl") + filepath.Ext(strings.TrimSuffix(path, ".tmpl"))
				if _, err := os.Stat(targetPath); err == nil {
					continue
				}
				err = renderTemplate(path, targetPath, task)
				if err != nil {
					return err
				}
			}
		} else {
			err = renderTemplate(path, outputPath, module)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func renderTemplate(templatePath, outputPath string, data interface{}) error {
	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return err
	}

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}).Parse(string(templateContent))
	if err != nil {
		return err
	}

	outputDir := filepath.Dir(outputPath)
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return err
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	err = tmpl.Execute(outputFile, data)
	if err != nil {
		return err
	}

	return nil
}
