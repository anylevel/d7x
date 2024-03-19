/*
Copyright B) 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/lithammer/shortuuid"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "A brief description of your command...",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		imageName := args[1]
		volumes, _ := cmd.Flags().GetStringSlice("volume")
		sandbox(name, imageName, volumes)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringSliceP("volume", "v", []string{}, "add mounts to container")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

var config container.Config = container.Config{
	OpenStdin:    true,
	Tty:          true,
	AttachStdin:  true,
	AttachStdout: true,
	AttachStderr: true,
	Cmd:          []string{"/bin/sh"},
}

var attachOptions container.AttachOptions = container.AttachOptions{
	Stderr:     true,
	Stdout:     true,
	Stdin:      true,
	Stream:     true,
	DetachKeys: "ctrl-d",
}

var removeOptions container.RemoveOptions = container.RemoveOptions{
	Force: true,
}

var ctx context.Context = context.Background()

func checkExistImage(apiclient *client.Client, imageName string) {
	images, err := apiclient.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		if imageName == image.RepoTags[0] {
			return
		}
	}
	fmt.Printf("%s is not on the host\nTry->docker pull %s\n", imageName, imageName)
	os.Exit(1)
}

func getMountsFromSlice(volumes []string) (result []mount.Mount) {
	for _, value := range volumes {
		separated := strings.Split(value, ":")
		result = append(result, mount.Mount{
			Type:   mount.TypeBind,
			Source: separated[0],
			Target: separated[1],
		})
	}
	return result
}

func sandbox(name string, imageName string, volumes []string) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer apiClient.Close()
	checkExistImage(apiClient, imageName)
	volumesToContainer := []mount.Mount{}
	if len(volumes) != 0 {
		volumesToContainer = getMountsFromSlice(volumes)
	}
	fmt.Printf("Start container %s  <- from image: %s\n", name, imageName)
	config.Image = imageName
	config.Hostname = name

	id, err := apiClient.ContainerCreate(
		ctx,
		&config,
		&container.HostConfig{Mounts: volumesToContainer},
		&network.NetworkingConfig{},
		&specs.Platform{},
		name)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Done, %s Container Creating\n", name)
	err = apiClient.ContainerStart(ctx, id.ID, container.StartOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s is starting, Attaching to terminal...\n", name)
	waiter, err := apiClient.ContainerAttach(ctx, id.ID, attachOptions)
	if err != nil {
		panic(err)
	}
	defer waiter.Close()
	go io.Copy(os.Stdout, waiter.Reader)
	go io.Copy(os.Stderr, waiter.Reader)
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
	term.Restore(fd, oldState)
	repoWithTag := fmt.Sprintf("%s:%s", name, shortuuid.New())
	_, err = apiClient.ContainerCommit(ctx, id.ID, container.CommitOptions{Reference: repoWithTag})
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nContainer %s save to image->%s\n", name, repoWithTag)
	err = apiClient.ContainerRemove(ctx, id.ID, removeOptions)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Removing %s container...\n", name)
}
