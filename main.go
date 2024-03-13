package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/term"
)

func main() {
	ctx := context.Background()
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer apiClient.Close()

	containers, err := apiClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		panic(err)
	}

	for _, ctr := range containers {
		fmt.Printf("%s %s (status: %s)\n", ctr.ID, ctr.Image, ctr.Status)
	}

	fmt.Printf("Start container ubuntu\n")
	id, err := apiClient.ContainerCreate(ctx, &container.Config{
		OpenStdin:    true,
		Hostname:     "sandbox",
		Tty:          true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"/bin/bash"},
		Image:        "ubuntu:22.04"},
		&container.HostConfig{}, &network.NetworkingConfig{}, &specs.Platform{}, "sandbox")
	if err != nil {
		fmt.Errorf("Error:%s", err)
	}
	fmt.Printf("Done, %s Container Creating\n", id.ID)
	err = apiClient.ContainerStart(ctx, id.ID, container.StartOptions{})
	if err != nil {
		fmt.Errorf("Error:%s", err)
	}
	fmt.Println("Container sandbox is starting")
	fmt.Println("Now attaching sandbox container")
	waiter, err := apiClient.ContainerAttach(ctx, id.ID, types.ContainerAttachOptions{
		Stderr:     true,
		Stdout:     true,
		Stdin:      true,
		Stream:     true,
		DetachKeys: "ctrl-d",
	})
	defer waiter.Close()
	go io.Copy(os.Stdout, waiter.Reader)
	go io.Copy(os.Stderr, waiter.Reader)
	if err != nil {
		panic(err)
	}

	fd := int(os.Stdin.Fd())
	var oldState *term.State
	if term.IsTerminal(fd) {
		oldState, err = term.MakeRaw(fd)
		if err != nil {
			panic(err)
		}
	}
	waiter.Conn.Write([]byte("\n"))
	go func() {
		for {
			consoleReader := bufio.NewReaderSize(os.Stdin, 1)
			input, _ := consoleReader.ReadByte()
			//https://donsnotes.com/tech/charsets/ascii.html#EOL
			if input == 4 {
				value := 0
				err = apiClient.ContainerStop(ctx, id.ID, container.StopOptions{Timeout: &value})
				if err != nil {
					panic(err)
				}
				break
			}
			waiter.Conn.Write([]byte{input})
		}
	}()

	statusCh, errCh := apiClient.ContainerWait(context.Background(), id.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}
	terminal.Restore(fd, oldState)
	fmt.Println("\nContainer sandbox commit to image")
	resp, err := apiClient.ContainerCommit(ctx, id.ID, container.CommitOptions{Reference: "hello:0.0.1"})
	if err != nil {
		panic(err)
	}
	fmt.Printf("ID is :%s\n", resp.ID)
	err = apiClient.ContainerRemove(ctx, id.ID, container.RemoveOptions{Force: true})
	if err != nil {
		panic(err)
	}
	fmt.Printf("ID:%s is stopping\n", id.ID)
}
