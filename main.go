package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cli/safeexec"
	"github.com/spf13/cobra"
)

type contributeOpts struct {
	Repository string
}

func rootCmd() *cobra.Command {
	opts := contributeOpts{}
	cmd := &cobra.Command{
		Use:   "contribute",
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
	viewOut := strings.Split(sout.String(), "\n")[0]
	repo := strings.TrimSpace(strings.Split(viewOut, ":")[1])

	return repo, nil
}

func runContribute(opts contributeOpts) error {
	hwIssues, err := issuesByLabel(opts.Repository, "help wanted")
	if err != nil {
		return err
	}

	gfIssues, err := issuesByLabel(opts.Repository, "good first issue")
	if err != nil {
		return err
	}

	choices := []issue{}

	seen := map[int]bool{}
	for _, issueList := range [][]issue{hwIssues, gfIssues} {
		for _, issue := range issueList {
			if _, ok := seen[issue.Number]; ok {
				continue
			}
			seen[issue.Number] = true

			blocked := false
			for _, label := range issue.Labels {
				blocked = (label.Name == "blocked")
			}
			if blocked {
				continue
			}

			cutoff, _ := time.ParseDuration("8760h")
			elapsed := time.Since(issue.CreatedAt)
			if elapsed > cutoff {
				continue
			}

			choices = append(choices, issue)
		}
	}

	var chosen *issue
	var prLookupErr error
	for _, ix := range rand.Perm(len(choices)) {
		pr, prLookupErr := hasPR(opts.Repository, choices[ix])
		if prLookupErr != nil {
			continue
		}
		if !pr {
			chosen = &choices[ix]
			break
		}
	}

	if chosen == nil {
		if prLookupErr != nil {
			return fmt.Errorf("failed talking to API about issues: %w", prLookupErr)
		}

		fmt.Println("Unable to find a good issue to suggest :( try perusing the repository with `gh issue list`.")

		return nil
	}

	// TODO make prettier?

	fmt.Printf("A good issue to work on in %s might be...\n\n", opts.Repository)
	fmt.Printf("Issue #%d: %s\n\n", chosen.Number, chosen.Title)
	fmt.Printf("You can check out the issue with `gh issue view -R%s %d`\n", opts.Repository, chosen.Number)
	fmt.Printf("Or express interest in working on it: `gh issue comment -R%s %d -b\"Hi! I'm interested in working on this\"`\n\n", opts.Repository, chosen.Number)
	fmt.Printf("View this issue on the web: %s\n", chosen.URL)

	return nil
}

func main() {
	rc := rootCmd()

	if err := rc.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func hasPR(repository string, i issue) (bool, error) {
	repoParts := strings.Split(repository, "/")
	query := fmt.Sprintf(`query {
		repository(name:"%s", owner:"%s") {
			issue(number: %d){
				timelineItems(last:10, itemTypes:[CROSS_REFERENCED_EVENT]){
					edges {
						node {
							...on CrossReferencedEvent {
								willCloseTarget}}}}}}}`,
		repoParts[0],
		repoParts[1],
		i.Number)

	sout, _, err := gh("api", "graphql", "-f", fmt.Sprintf("query=%s", query))
	if err != nil {
		return false, err
	}

	type response struct {
		Data struct {
			Repository struct {
				Issue struct {
					TimelineItems struct {
						Edges []struct {
							Node struct {
								WillCloseTarget bool
							}
						}
					}
				}
			}
		}
	}

	var resp response

	err = json.Unmarshal(sout.Bytes(), &resp)
	if err != nil {
		return false, err
	}

	edges := resp.Data.Repository.Issue.TimelineItems.Edges
	for _, edge := range edges {
		if edge.Node.WillCloseTarget {
			return true, nil
		}
	}

	return false, nil
}

type label struct {
	Name string
}

type issue struct {
	Number    int
	Title     string
	URL       string `json:"url"`
	CreatedAt time.Time
	Labels    []label
}

func issuesByLabel(repository, label string) ([]issue, error) {
	sout, _, err := gh("issue", "list", "-l", label, "-R", repository, "--json", "number,title,labels,url,createdAt")
	if err != nil {
		return nil, err
	}
	var result []issue
	err = json.Unmarshal(sout.Bytes(), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
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
