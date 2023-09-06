package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"gl.biggo.pro/devopstrain/kurator/lib"
)

func main() {
	app := &cli.App{
		Name:  "Kurator",
		Usage: "Devopstrain course helper for students and developers",
		Commands: []*cli.Command{
			{
				Name:    "login",
				Usage:   "Login to the application",
				Aliases: []string{"l"},
				Action:  lib.LoginUserCLI,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "email",
						Usage:    "Email address",
						Required: true,
					},
				},
			},
			{
				Name:    "signup",
				Usage:   "Sing up (register new account)",
				Aliases: []string{"s"},
				Action:  lib.SignUpUserCLI,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "email",
						Usage:    "Email address",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Your name",
						Required: true,
					},
				},
			},
			{
				Name:  "course",
				Usage: "Manage courses",
				Subcommands: []*cli.Command{
					{
						Name:   "list",
						Usage:  "List all courses",
						Action: lib.ListCourses,
					},
					{
						Name:   "start",
						Usage:  "Start course validator",
						Action: lib.StartCourse,
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "dev",
								Usage: "Run dev commands. Use only if you're developer",
							},
						},
					},
				},
			},
			{
				Name:  "dev",
				Usage: "Develop courses",
				Subcommands: []*cli.Command{
					{
						Name:   "run-server",
						Usage:  "Run Dev Server",
						Action: lib.RunDevServer,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "course_name",
								Usage:    "Course name",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "handler_url",
								Usage:    "Your backend handler url",
								Required: true,
							},
						},
					},
					{
						Name:   "create-course",
						Usage:  "Create a new course",
						Action: lib.CreateCourse,
					},
					{
						Name:   "generate-code",
						Usage:  "Generate handler code",
						Action: lib.GenerateHandlerCode,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "course_name",
								Usage:    "Course name",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "template_path",
								Usage:    "path to backend specific templates",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "output_path",
								Usage:    "path to save the result code",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "module_name",
								Usage:    "module name to use in templates. Sample: gl.biggo.pro/devopstrain/kubernetes_handler",
								Required: true,
							},
						},
					},
				},
			},
			{
				Name:    "version",
				Usage:   "Version of the application",
				Aliases: []string{"v"},
				Action:  lib.Version,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
