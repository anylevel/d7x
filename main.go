package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/anylevel/d7x/cmd"
)

func main() {
	if _, err := os.Stat("/var/run/docker.sock"); errors.Is(err, os.ErrNotExist) {
		fmt.Println(`Error connect to docker daemon
Try:
sudo systemctl start docker
or
install docker`)
		os.Exit(1)
	}
	os.Setenv("DOCKER_API_VERSION", "1.41")
	cmd.Execute()
}
