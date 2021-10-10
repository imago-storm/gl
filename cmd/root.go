package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/imago-storm/gl/gitlab"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gl",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		cmd.Help()
		os.Exit(0)
		fmt.Println("Stuff")
	},
}

var repoCmd = &cobra.Command{
	Use: "repo",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

var mrCmd = &cobra.Command{
	Use: "mr",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

var createMrCmd = &cobra.Command{
	Use: "create",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Calculate branch
		// Look for open mr
		path, err := os.Getwd()
		if err != nil {
			return err
		}
		repo, err := gitlab.OpenRepository(path)
		err = repo.CreateMergeRequest()
		return err
	},
}

var openMrCmd = &cobra.Command{
	Use: "open",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := gitlab.OpenRepositoryCwd()
		if err != nil {
			return fmt.Errorf("Failed to open repository: %s", err)
		}
		err = repo.OpenMergeRequest()
		return err
	},
}

var repoOpenCmd = &cobra.Command{
	Use: "open",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := os.Getwd()
		if err != nil {
			return err
		}
		log.Println("Current path: ", path)
		if err := gitlab.OpenRemote(path); err != nil {
			return err
		}
		return nil
	},
}

func Execute() {
	// fmt.Println("Command")
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// cobra.OnInitialize()

	repoCmd.AddCommand(repoOpenCmd)
	mrCmd.AddCommand(createMrCmd)
	mrCmd.AddCommand(openMrCmd)

	rootCmd.AddCommand(repoCmd)
	rootCmd.AddCommand(mrCmd)
}
