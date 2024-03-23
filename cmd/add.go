package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		imageName := args[0]
		adds, _ := cmd.Flags().GetString("add")
		arg, _ := cmd.Flags().GetStringToString("arg")
		command, _ := cmd.Flags().GetString("cmd")
		copy, _ := cmd.Flags().GetString("copy")
		entryPoint, _ := cmd.Flags().GetString("entrypoint")
		envs, _ := cmd.Flags().GetStringToString("env")
		expose, _ := cmd.Flags().GetStringSlice("expose")
		healthCheck, _ := cmd.Flags().GetString("healthcheck")
		labels, _ := cmd.Flags().GetStringToString("label")
		mntr, _ := cmd.Flags().GetString("maintainer")
		onbuild, _ := cmd.Flags().GetStringSlice("onbuild")
		run, _ := cmd.Flags().GetStringSlice("run")
		shell, _ := cmd.Flags().GetString("shell")
		stopSignal, _ := cmd.Flags().GetString("stopsignal")
		user, _ := cmd.Flags().GetString("usr")
		volume, _ := cmd.Flags().GetStringSlice("volume")
		wrkdir, _ := cmd.Flags().GetString("wrkdir")
		noSave, _ := cmd.Flags().GetBool("nosave")
		noBuild, _ := cmd.Flags().GetBool("nobuild")
		saveName, _ := cmd.Flags().GetString("name")
		currentDockerFile := dockerFile{
			baseImage:   imageName,
			add:         adds,
			args:        arg,
			cmd:         command,
			copy:        copy,
			entryPoint:  entryPoint,
			envs:        envs,
			expose:      expose,
			healthCheck: healthCheck,
			labels:      labels,
			maintainer:  mntr,
			onbuild:     onbuild,
			run:         run,
			shell:       shell,
			stopSignal:  stopSignal,
			user:        user,
			volumes:     volume,
			workDir:     wrkdir,
		}
		add(&currentDockerFile, noSave, noBuild, saveName)
	},
}

type dockerFile struct {
	baseImage   string
	add         string
	args        map[string]string
	cmd         string
	copy        string
	entryPoint  string
	envs        map[string]string
	expose      []string
	healthCheck string
	labels      map[string]string
	maintainer  string
	onbuild     []string
	run         []string
	shell       string
	stopSignal  string
	user        string
	volumes     []string
	workDir     string
}

var imageBuildOptions types.ImageBuildOptions = types.ImageBuildOptions{
	Dockerfile: "Dockerfile",
	Remove:     true,
}

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

type BuildOutput struct {
	Stream string `json:"stream"`
	Aux    string `json:"aux"`
}

func init() {
	rootCmd.AddCommand(addCmd)
	//https://docs.docker.com/reference/dockerfile/
	addCmd.Flags().StringP("add", "a", "", "Add local or remote files and directories.")
	addCmd.Flags().StringToString("arg", nil, "Use build-time variables.")
	addCmd.Flags().String("cmd", "", "Specify default commands.")
	addCmd.Flags().StringP("copy", "c", "", "Copy files and directories.")
	addCmd.Flags().String("entrypoint", "", "Specify default executable.")
	addCmd.Flags().StringToStringP("env", "e", nil, "Set environment variables.")
	addCmd.Flags().StringSlice("expose", nil, "Describe which ports your application is listening on.")
	addCmd.Flags().String("healthcheck", "", "Check a container's health on startup.")
	addCmd.Flags().StringToStringP("label", "l", nil, "Add metadata to an image.")
	addCmd.Flags().StringP("maintainer", "m", "", "Specify the author of an image.")
	addCmd.Flags().StringSlice("onbuild", nil, "Specify instructions for when the image is used in a build.")
	addCmd.Flags().StringSliceP("run", "r", nil, "Execute build commands.")
	addCmd.Flags().StringP("sh", "s", "", "Set the default shell of an image.")
	addCmd.Flags().String("stopsignal", "", "Specify the system call signal for exiting a container.")
	addCmd.Flags().StringP("usr", "u", "", "Set user and group ID.")
	addCmd.Flags().StringSliceP("volume", "v", nil, "Create volume mounts.")
	addCmd.Flags().StringP("wrkdir", "w", "", "Change working directory.")
	//other Flags
	addCmd.Flags().Bool("nosave", false, "Save to Dockerfile")
	addCmd.Flags().Bool("nobuild", false, "Dont build image from Dockerfile")
	addCmd.Flags().String("name", "", "Name for saved image")
}

func add(currentDockerFile *dockerFile, noSave bool, noBuild bool, saveName string) {
	pathWd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	pathToSave := filepath.Join(pathWd, "Dockerfile")
	f, err := os.OpenFile(pathToSave, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if currentDockerFile.baseImage != "" {
	}
	if currentDockerFile.maintainer != "" {
		writeLineToDockerFile(f, "MAINTAINER", currentDockerFile.maintainer)
	}
	writeLineToDockerFile(f, "FROM", currentDockerFile.baseImage)
	if len(currentDockerFile.labels) != 0 {
		writeMapToDockerFile(f, "LABEL", currentDockerFile.labels)
	}
	if len(currentDockerFile.args) != 0 {
		writeMapToDockerFile(f, "ARG", currentDockerFile.args)
	}
	if len(currentDockerFile.envs) != 0 {
		writeMapToDockerFile(f, "ENV", currentDockerFile.envs)
	}
	if currentDockerFile.add != "" {
		writeLineToDockerFile(f, "ADD", currentDockerFile.add)
	}
	if currentDockerFile.copy != "" {
		writeLineToDockerFile(f, "COPY", currentDockerFile.copy)
	}
	if currentDockerFile.workDir != "" {
		writeLineToDockerFile(f, "WORKDIR", currentDockerFile.workDir)
	}
	if len(currentDockerFile.run) != 0 {
		writeSliceToDockerFile(f, "RUN", currentDockerFile.run)
	}
	if len(currentDockerFile.onbuild) != 0 {
		writeSliceToDockerFile(f, "ONBUILD", currentDockerFile.onbuild)
	}
	if currentDockerFile.stopSignal != "" {
		writeLineToDockerFile(f, "STOPSIGNAL", currentDockerFile.stopSignal)
	}
	if currentDockerFile.shell != "" {
		writeLineToDockerFile(f, "SHELL", currentDockerFile.shell)
	}
	if currentDockerFile.user != "" {
		writeLineToDockerFile(f, "USER", currentDockerFile.user)
	}
	if currentDockerFile.healthCheck != "" {
		writeLineToDockerFile(f, "HEALTHCHECK", currentDockerFile.healthCheck)
	}
	if len(currentDockerFile.volumes) != 0 {
		writeSliceToDockerFile(f, "VOLUME", currentDockerFile.volumes)
	}
	if len(currentDockerFile.expose) != 0 {
		writeSliceToDockerFile(f, "EXPOSE", currentDockerFile.expose)
	}
	if currentDockerFile.entryPoint != "" {
		writeLineToDockerFile(f, "ENTRYPOINT", currentDockerFile.entryPoint)
	}
	if currentDockerFile.cmd != "" {
		writeLineToDockerFile(f, "CMD", currentDockerFile.cmd)
	}
	if noBuild != true {
		err = createImage(currentDockerFile.baseImage, saveName)
		if err != nil {
			panic(err)
		}
	}
	if noSave {
		err = os.Remove("Dockerfile")
		if err != nil {
			panic(err)
		}
	}
}

func writeSliceToDockerFile(srcFile *os.File, instruction string, data []string) {
	imageLine := instruction
	for _, value := range data {
		imageLine = fmt.Sprintf("%s %s", imageLine, value)
	}
	imageLine = fmt.Sprintf("%s\n", imageLine)
	srcFile.WriteString(imageLine)
}

func writeMapToDockerFile(srcFile *os.File, instruction string, data map[string]string) {
	imageLine := instruction
	for key, value := range data {
		imageLine = fmt.Sprintf("%s %s=%s", imageLine, key, value)
	}
	imageLine = fmt.Sprintf("%s\n", imageLine)
	srcFile.WriteString(imageLine)
}

func writeLineToDockerFile(srcFile *os.File, instruction string, data string) {
	imageLine := fmt.Sprintf("%s %s", instruction, data)
	imageLine = fmt.Sprintf("%s\n", imageLine)
	srcFile.WriteString(imageLine)
}

func createImage(imageName string, saveName string) (err error) {
	ctx := context.Background()
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	defer apiClient.Close()
	tar, err := archive.TarWithOptions("./", &archive.TarOptions{})
	if err != nil {
		return err
	}
	imageName = fmt.Sprintf("%s-d7x", imageName)
	if saveName != "" {
		imageName = saveName
	}
	imageBuildOptions.Tags = []string{imageName}
	res, err := apiClient.ImageBuild(ctx, tar, imageBuildOptions)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	err = print(res.Body)
	if err != nil {
		return err
	}

	return nil
}

func print(rd io.Reader) error {
	var lastLine []byte
	output := &BuildOutput{}
	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		lastLine = scanner.Bytes()
		json.Unmarshal(lastLine, output)
		switch {
		case output.Stream != "" && output.Stream != "\n":
			fmt.Print(output.Stream)
		case output.Aux != "":
			fmt.Print(output.Aux)
		}
	}

	errLine := &ErrorLine{}
	json.Unmarshal(lastLine, errLine)
	if errLine.Error != "" {
		return errors.New(errLine.Error)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
