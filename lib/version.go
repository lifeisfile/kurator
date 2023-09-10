package lib

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var version = "0.0.4"

func Version(c *cli.Context) error {
	fmt.Printf("Version: %s\n\n", version)
	return nil
}
