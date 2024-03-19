package cmd

import (
	"fmt"
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
		add(imageName)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	//https://docs.docker.com/reference/dockerfile/
	addCmd.Flags().StringP("add", "a", "", "Add local or remote files and directories.")
	addCmd.Flags().StringToString("arg", map[string]string{}, "Use build-time variables.")
	addCmd.Flags().StringP("cmd", "c", "", "Specify default commands.")
	addCmd.Flags().StringP("copy", "c", "", "Copy files and directories.")
	addCmd.Flags().StringP("entrypoint", "entp", "", "Specify default executable.")
	addCmd.Flags().StringToStringP("env", "e", map[string]string{}, "Set environment variables.")
	addCmd.Flags().StringP("expose", "exps", "", "Describe which ports your application is listening on.")
	addCmd.Flags().StringP("healthcheck", "health", "", "Check a container's health on startup.")
	addCmd.Flags().StringToStringP("label", "l", map[string]string{}, "Add metadata to an image.")
	addCmd.Flags().StringP("maintainer", "m", "", "Specify the author of an image.")
	addCmd.Flags().StringSliceP("onbuild", "onbld", []string{}, "Specify instructions for when the image is used in a build.")
	addCmd.Flags().StringSliceP("run", "r", []string{}, "Execute build commands.")
	addCmd.Flags().StringP("sh", "s", "", "Set the default shell of an image.")
	addCmd.Flags().StringP("stopsignal", "stpsig", "", "Specify the system call signal for exiting a container.")
	addCmd.Flags().StringP("usr", "u", "", "Set user and group ID.")
	addCmd.Flags().StringP("volume", "v", "", "Create volume mounts.")
	addCmd.Flags().StringP("wrkdir", "w", "", "Change working directory.")
}

func add(imageName string) {
	tempName := shortuuid.New()
	fullPathDockerFile := filepath.Join("/tmp", tempName)
	f, err := os.Create(fullPathDockerFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	baseImageLine := fmt.Sprintf("FROM %s\n", imageName)
	f.WriteString(baseImageLine)
	if 


	}
}
