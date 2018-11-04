package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"gopkg.in/andygrunwald/go-jira.v1"
)

var Version string
var Commit string
var Date string

type Transitions string

const (
	REVIEW_STATUS     Transitions = "In (Code) Review"
	IN_PROGESS_STATUS Transitions = "In Progress"
	READY_FOR_QA      Transitions = "Ready for QA"
)

func (t Transitions) String() string {
	return string(t)
}

var token = ""
var baseUrl = ""
var username = ""
var dryRun = false

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    true,
	})

	app := cli.NewApp()
	app.Name = "git-jira"
	app.Version = Version + " " + Commit
	app.Compiled = time.Now()
	app.UsageText = "git jira #issue [global options]"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "username, u",
			EnvVar:      "JIRA_USERNAME",
			Usage:       "Your Jira username, usually your email.",
			Destination: &username,
		},
		cli.StringFlag{
			Name:        "token, t",
			EnvVar:      "JIRA_API_TOKEN",
			Usage:       "An API token created on https://id.atlassian.com/manage/api-tokens",
			Destination: &token,
		},
		cli.StringFlag{
			Name:        "base-url, b",
			EnvVar:      "JIRA_BASE_URL",
			Usage:       "e.g mycompany.atlassian.com",
			Destination: &baseUrl,
		},
		cli.BoolFlag{
			Name:        "dry-run, d",
			Usage:       "Do not run any git commands or modify any resource via Jira API.",
			Destination: &dryRun,
		},
	}


	app.Action = func(c *cli.Context) error {
		err := GitJira(c)
		if err != nil {
			log.Fatal(err.Error())
		}
		return err
	}

	app.Run(os.Args)
}

var issueIdRegex, _ = regexp.Compile(`\S+-\d+`)
var issueTitleRegex, _ = regexp.Compile("[^a-zA-Z0-9]+")

func pareseIssueId(ctx *cli.Context) string {
	issueRaw := ctx.Args().First()
	splits := strings.Split(issueRaw, "/")
	issueID := splits[len(splits)-1]

	if strings.TrimSpace(issueID) == "" {
		cli.ShowAppHelp(ctx)
		log.Exit(1)
	}

	if !issueIdRegex.MatchString(issueID) {
		exitFatal(issueID, nil, fmt.Sprintf("Incorrect issue format: %s", issueID))
	}

	return issueID
}

func GitJira(context *cli.Context) error {
	issueID := pareseIssueId(context)

	if username == "" {
		exitFatal(issueID, nil, "--username or JIRA_USERNAME was empty")
	}

	if token == "" {
		exitFatal(issueID, nil, "--token or JIRA_API_TOKEN was empty")
	}

	if baseUrl == "" {
		exitFatal(issueID, nil, "--base-url or JIRA_BASE_URL was empty")
	}

	tp := jira.BasicAuthTransport{
		Username: username,
		Password: token,
	}

	client, err := jira.NewClient(tp.Client(), "https://"+baseUrl)
	if err != nil {
		exitFatal(issueID, err, "error while creating Jira http client")
	}

	// 1.  Create branch from issue nam,e
	issue, _, err := client.Issue.Get(issueID, &jira.GetQueryOptions{Fields: "summary"})
	if err != nil {
		exitFatal(issueID, err, "error while getting Jira issue")
	}

	var issueType = getIssueType(issue.Fields.Type)
	var branchName = createBranchName(issue.Fields.Summary)

	var gitBranchName = fmt.Sprintf("%s/%s_%s", issueType, issueID, branchName)

	gitCheckout(gitBranchName)

	// 2. Create an empty first commit in the new branch
	var commitHeader = fmt.Sprintf("[%s] %s", issueID, issue.Fields.Summary)

	var issueLink = "https://" + baseUrl + "/browse/" + issueID

	gitCommit(commitHeader, "\t", issueLink)

	// 3. Set the issue to in progress
	transitions, _, err := client.Issue.GetTransitions(issueID)
	if err != nil {
		exitFatal(issueID, err, "error while getting all transitions")
	}

	for _, v := range transitions {
		if v.Name == IN_PROGESS_STATUS.String() {
			if dryRun {
				log.Println("dry-run: transition issue to 'In-Progress")
				break
			}
			_, err := client.Issue.DoTransition(issueID, v.ID)
			if err != nil {
				exitFatal(issueID, err, "error while transitioning issue to 'In-Progress' state")
			}
			break
		}
	}

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
	if dryRun {
		log.Println("dry-run: " + strings.Join(command, " "))
		return nil
	}

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

	output = strings.Trim(output, "_")

	if len(output) > 60 {
		output = output[:60]
	}

	return strings.Trim(output, "_")
}

func exitFatal(issueID string, err error, message string) {
	log.WithError(err).WithField("issueID", issueID).Fatalln(message)
}
