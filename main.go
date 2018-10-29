package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"gopkg.in/andygrunwald/go-jira.v1"
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
	var branchName = createBranchName(issue.Fields.Summary)

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

var execCommand = exec.Command

func generateGitCheckout(branch string) []string {
	args := []string{
		"git",
		"checkout",
		"-b",
		branch,
	}

	return args
}

func gitCheckout(branchName string) {
	var command = generateGitCheckout(branchName)
	executeCommand(command)
}

func generateGitCommit(message ...string) []string {
	args := []string{
		"git",
		"commit",
		"--allow-empty",
	}

	for _, m := range message {
		args = append(args, "-m")
		args = append(args, m)
	}

	return args
}

func gitCommit(message ...string) {
	var command = generateGitCommit(message...)
	executeCommand(command)
}

func executeCommand(command []string) *exec.Cmd {
	cmd := execCommand(command[0], command[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.WithError(err).Fatal("failure while executing command")
	}
	return cmd
}

func trimRules(input string) string {
	var output = strings.ToLower(input)

	output = strings.Replace(output, "android", "", -1)
	output = strings.Replace(output, "ios", "", -1)

	return strings.TrimSpace(output)
}

func createBranchName(issueTitle string) string {
	output := trimRules(issueTitle)
	output = issueTitleRegex.ReplaceAllString(output, "_")

	var lastRune rune
	output = strings.Map(func(r rune) rune {
		if lastRune == r && lastRune == '_' {
			return -1
		}

		lastRune = r
		return r
	}, output)

	return strings.Trim(output, "_")
}
