package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lithammer/shortuuid"
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
		expose, _ := cmd.Flags().GetString("expose")
		healthCheck, _ := cmd.Flags().GetString("healthcheck")
		labels, _ := cmd.Flags().GetStringToString("label")
		mntr, _ := cmd.Flags().GetString("maintainer")
		onbuild, _ := cmd.Flags().GetStringSlice("onbuild")
		run, _ := cmd.Flags().GetStringSlice("run")
		shell, _ := cmd.Flags().GetString("shell")
		stopSignal, _ := cmd.Flags().GetString("stopsignal")
		user, _ := cmd.Flags().GetString("usr")
		volume, _ := cmd.Flags().GetString("volume")
		wrkdir, _ := cmd.Flags().GetString("wrkdir")
		save, _ := cmd.Flags().GetBool("output")
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
		add(&currentDockerFile, save)
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
	expose      string
	healthCheck string
	labels      map[string]string
	maintainer  string
	onbuild     []string
	run         []string
	shell       string
	stopSignal  string
	user        string
	volumes     string
	workDir     string
}

func init() {
	rootCmd.AddCommand(addCmd)
	//https://docs.docker.com/reference/dockerfile/
	addCmd.Flags().StringP("add", "a", "", "Add local or remote files and directories.")
	addCmd.Flags().StringToString("arg", map[string]string{}, "Use build-time variables.")
	addCmd.Flags().String("cmd", "", "Specify default commands.")
	addCmd.Flags().StringP("copy", "c", "", "Copy files and directories.")
	addCmd.Flags().String("entrypoint", "", "Specify default executable.")
	addCmd.Flags().StringToStringP("env", "e", map[string]string{}, "Set environment variables.")
	addCmd.Flags().String("expose", "", "Describe which ports your application is listening on.")
	addCmd.Flags().String("healthcheck", "", "Check a container's health on startup.")
	addCmd.Flags().StringToStringP("label", "l", map[string]string{}, "Add metadata to an image.")
	addCmd.Flags().StringP("maintainer", "m", "", "Specify the author of an image.")
	addCmd.Flags().StringSlice("onbuild", []string{}, "Specify instructions for when the image is used in a build.")
	addCmd.Flags().StringSliceP("run", "r", []string{}, "Execute build commands.")
	addCmd.Flags().StringP("sh", "s", "", "Set the default shell of an image.")
	addCmd.Flags().String("stopsignal", "", "Specify the system call signal for exiting a container.")
	addCmd.Flags().StringP("usr", "u", "", "Set user and group ID.")
	addCmd.Flags().StringP("volume", "v", "", "Create volume mounts.")
	addCmd.Flags().StringP("wrkdir", "w", "", "Change working directory.")
	//other Flags
	addCmd.Flags().BoolP("output", "o", false, "Save to Dockerfile")
}

func add(currentDockerFile *dockerFile, save bool) {
	tempName := shortuuid.New()
	fullPathDockerFile := filepath.Join("/tmp", tempName)
	f, err := os.Create(fullPathDockerFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if currentDockerFile.baseImage != "" {
		baseImageLine := fmt.Sprintf("FROM %s\n", currentDockerFile.baseImage)
		f.WriteString(baseImageLine)
	}
	if currentDockerFile.maintainer != "" {
		maintainerImageLine := fmt.Sprintf("MAINTAINER %s\n", currentDockerFile.maintainer)
		f.WriteString(maintainerImageLine)
	}
	if len(currentDockerFile.labels) != 0 {
		labelImageLine := "LABEL"
		for key, value := range currentDockerFile.labels {
			labelImageLine = fmt.Sprintf("%s %s=%s", labelImageLine, key, value)
		}
		f.WriteString(labelImageLine)
	}
	if save != false {
		err = saveDockerFile(f)
		if err != nil {
			panic(err)
		}
	}
}

func saveDockerFile(srcFile *os.File) (err error) {
	pathWd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	pathToSave := filepath.Join(pathWd, "Dockerfile")
	fSave, err := os.OpenFile(pathToSave, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer fSave.Close()
	srcFile.Seek(0, io.SeekStart)
	_, err = io.Copy(fSave, srcFile)
	if err != nil {
		panic(err)
	}
	return err
}

func writeSliceToDockerFile(srcFile *os.File, instruction string, data []string) {
	imageLine := instruction
	for _, value := range data {
		imageLine = fmt.Sprintf("%s %s", imageLine, value)
	}
	srcFile.WriteString(imageLine)
}

func writeMapToDockerFile(srcFile *os.File, instruction string, data map[string]string) {
	imageLine := instruction
	for key, value := range data {
		imageLine = fmt.Sprintf("%s %s=%s", imageLine, key, value)
	}
	srcFile.WriteString(imageLine)
}

func writeLineToDockerFile(srcFile *os.File, instruction string, data string) {

}
