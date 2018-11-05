
[![GoDoc](https://godoc.org/github.com/tevjef/git-jira?status.svg)](https://godoc.org/github.com/tevjef/git-jira)
[![Build Status](https://travis-ci.org/tevjef/git-jira.svg?branch=master)](https://travis-ci.org/tevjef/git-jira)
[![Go Report Card](https://goreportcard.com/badge/github.com/tevjef/git-jira)](https://goreportcard.com/report/github.com/tevjef/git-jira)

# git-jira

1. Creates a branch from a Jira issue title
2. Creates an empty commit with the link to the Jira issue.
3. Sets the Jira issue to "In-Progress"

## Usage

```
NAME:
   git-jira 

USAGE:
   git jira #issue [global options]

VERSION:


COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --username value, -u value  Your Jira username, usually your email. [$JIRA_USERNAME]
   --token value, -t value     An API token created on https://id.atlassian.com/manage/api-tokens [$JIRA_API_TOKEN]
   --base-url value, -b value  e.g mycompany.atlassian.com [$JIRA_BASE_URL]
   --dry-run, -d               Do not run any git commands or modify any resource via Jira API.
   --help, -h                  show help
   --version, -v               print the version

```

### Examples

```
$ git jira ASD-124
```
```
$ git jira https://id.atlassian.com/browse/ASD-124
```

##### Auth by CLI flags
```
$ git jira --username me@example.com --token y4bv87n4y5c845nyv84 --base company.atlassian.net ASD-124
```

##### Auth by Environment Variables
```
$ export JIRA_USERNAME=me@example.com
$ export JIRA_API_TOKEN=y4bv87n4y5c845nyv84
$ export JIRA_BASE_URL=company.atlassian.net
$ git jira ASD-124
```

### Installation

##### Brew

```bash
brew install tevjef/tap/git-jira
```

##### Build from source

Go 1.11 required:
```bash
go install github.com/tevjef/git-jira
```


