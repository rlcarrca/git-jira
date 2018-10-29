package main

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"gopkg.in/andygrunwald/go-jira.v1"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func main() {
	app := cli.NewApp()
	app.Name = "git-jira"
	app.Version = "0.1.0"
	app.Compiled = time.Now()
	app.UsageText = "git jira #ticket [global options]"
	app.Usage = "Perform jira operations for a git repo"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "username, u",
			EnvVar: "JIRA_USERNAME",
			Usage:  "Your Jira username, usually you email.",
		},
		cli.StringFlag{
			Name:   "token, t",
			EnvVar: "JIRA_API_TOKEN",
			Usage:  "API token created on https://id.atlassian.com/manage/api-tokens",
		},
		cli.StringFlag{
			Name:   "base-url, b",
			EnvVar: "JIRA_BASE_URL",
			Usage:  "e.g mycompany.atlassian.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		err := dostuff(c)
		if err != nil {
			log.Fatal(err.Error())
		}
		return err
	}

	app.Run(os.Args)
}

var issueIdRegex, _ = regexp.Compile(`\S+-\d+`)
var issueTitleRegex, _ = regexp.Compile("[^a-zA-Z0-9]+")

func dostuff(context *cli.Context) error {
	username := context.String("username")
	token := context.String("token")
	baseUrl := context.String("base-url")
	issueRaw := context.Args().First()

	splits := strings.Split(issueRaw, "/")
	issueID := splits[len(splits)-1]

	if issueID == "" {
		errors.New("The supplied issue is empty")
	}

	if !issueIdRegex.MatchString(issueID) {
		errors.New(fmt.Sprintf("Incorrect issue format: %s", issueID))
	}

	tp := jira.BasicAuthTransport{
		Username: username,
		Password: token,
	}

	client, err := jira.NewClient(tp.Client(), "https://"+baseUrl)
	if err != nil {
		return err
	}

	issue, _, err := client.Issue.Get(issueID, &jira.GetQueryOptions{Fields: "summary"})

	var issueType = getIssueType(issue.Fields.Type)
	var branchName = createBranchName(trimRules(issue.Fields.Summary))

	var gitBranchName = fmt.Sprintf("%s/%s_%s", issueType, issueID, branchName)

	gitCheckout(gitBranchName)

	var commitHeader = fmt.Sprintf("[%s] %s", issueID, issue.Fields.Summary)

	gitCommit(commitHeader, "\t", issue.Self)


	// TODO transition ticket to in-progress
	return nil
}

func getIssueType(issue jira.IssueType) string {
	if strings.ToLower(issue.Name) == "story" {
		return "feature"
	}

	if strings.ToLower(issue.Name) == "bug" {
		return "bug"
	}

	// Fallback to using a feature branch
	return "feature"
}

func gitCheckout(branchName string) {
	cmd := exec.Command("git", "checkout", "-b",  branchName)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func gitCommit(message ...string) {
	args := []string{
		"commit",
		"--allow-empty",
	}

	for _, m := range message {
		args = append(args, "-m")
		args = append(args, m)
	}

	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func trimRules(input string) string {
	var output = input

	output = strings.Replace(strings.ToLower(output), "android", "", -1)
	output = strings.Replace(strings.ToLower(output), "ios", "", -1)

	return output
}

// Android | Founders' debit home
func createBranchName(issueTitle string) string {
	replacementString := issueTitleRegex.ReplaceAllString(issueTitle, "_")

	var lastRune rune
	replacementString = strings.Map(func(r rune) rune {
		if lastRune == r {
			return -1
		}

		lastRune = r

		return r
	}, replacementString)

	return strings.Trim(replacementString, "_")
}

/*


at my last gig I wrote something similar for PivotalTracker (jira lite). I also wrote a git extension that:

`git pivotal #tickedId`

• creates branch with correct name (bug/ * feature/ *) X
• create appropriate branch name from ticket name X
• creates a commit with the body and URL of the ticket. ? // git commit -m "My head line" -m "My content line."
• marks the ticket as in progress. (edited)
Because my laziness knows no bounds





*/
