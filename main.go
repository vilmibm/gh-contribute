package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cli/safeexec"
	"github.com/spf13/cobra"
)

type contributeOpts struct {
	Repository string
}

func rootCmd() *cobra.Command {
	opts := contributeOpts{}
	cmd := &cobra.Command{
		Use:   "contribute [<repository>]",
		Short: "Suggest an issue to work on in a given repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Repository == "" {
				repo, err := resolveRepository()
				if err != nil {
					return err
				}
				opts.Repository = repo
			}
			return runContribute(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Repository, "repo", "R", "", "Repository to contribute to")

	return cmd
}

func resolveRepository() (string, error) {
	sout, _, err := gh("repo", "view")
	if err != nil {
		return "", err
	}
	viewOut := sout.String()
	viewOut = strings.Split(viewOut, "\n")[0]
	repo := strings.TrimSpace(strings.Split(viewOut, ":")[1])

	return repo, nil
}

func runContribute(opts contributeOpts) error {
	fmt.Println(opts.Repository)
	sout, _, err := gh("issue", "list")
	if err != nil {
		return err
	}
	fmt.Println(sout.String())

	return nil
}

func main() {
	rc := rootCmd()

	if err := rc.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

// gh shells out to gh, returning STDOUT/STDERR and any error
func gh(args ...string) (sout, eout bytes.Buffer, err error) {
	ghBin, err := safeexec.LookPath("gh")
	if err != nil {
		err = fmt.Errorf("could not find gh. Is it installed? error: %w", err)
		return
	}

	cmd := exec.Command(ghBin, args...)
	cmd.Stderr = &eout
	cmd.Stdout = &sout

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("failed to run gh. error: %w, stderr: %s", err, eout.String())
		return
	}

	return
}
