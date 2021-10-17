package cmd

import (
	"fmt"
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
		repo, err := gitlab.OpenRepositoryCwd()
		if err != nil {
			return err
		}
		open, _ := cmd.Flags().GetBool("view")
		branch, _ := cmd.Flags().GetString("branch")
		mr, err := repo.CreateMergeRequest(&gitlab.CreateMergeRequest{Branch: &branch})
		if err != nil {
			return err
		}
		if open {
			gitlab.OpenBrowser(mr.WebURL)
		}
		return err
	},
}

var openMrCmd = &cobra.Command{
	Use: "view",
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
	Use: "view",
	RunE: func(cmd *cobra.Command, args []string) error {

		repo, err := gitlab.OpenRepositoryCwd()
		if err != nil {
			return err
		}
		return repo.OpenRemoteURL()
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

	createMrCmd.PersistentFlags().Bool("view", false, "Open created MR immediately in your browser")
	createMrCmd.PersistentFlags().String("branch", "master", "Base branch to create MR to (default repository branch by default)")
	mrCmd.AddCommand(createMrCmd)

	mrCmd.AddCommand(openMrCmd)

	rootCmd.AddCommand(repoCmd)
	rootCmd.AddCommand(mrCmd)
}
