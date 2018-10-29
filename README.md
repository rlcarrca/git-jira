
[![GoDoc](https://godoc.org/github.com/tevjef/git-jira?status.svg)](https://godoc.org/github.com/tevjef/git-jira)
[![Build Status](https://travis-ci.org/tevjef/git-jira.svg?branch=master)](https://travis-ci.org/tevjef/git-jira)
[![Go Report Card](https://goreportcard.com/badge/github.com/tevjef/git-jira)](https://goreportcard.com/report/github.com/tevjef/git-jira)

# git-jira
A git utility to create a branch from a JIRA ticket


## Getting Started


### Usage

```
NAME:
   git-jira - Perform jira operations for a git repo

USAGE:
   git jira #ticket [global options]

VERSION:
   0.1.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --username value, -u value  Your Jira username, usually you email. [$JIRA_USERNAME]
   --token value, -t value     API token created on https://id.atlassian.com/manage/api-tokens [$JIRA_API_TOKEN]
   --base-url value, -b value  e.g mycompany.atlassian.com [$JIRA_BASE_URL]
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

#### Auth from CLI (flags)
```
$ git jira --username me@example.com --token y4bv87n4y5c845nyv84 --base company.atlassian.net ASD-124
```

#### Auth from CLI (environment variables)
```
$ export JIRA_USERNAME=me@example.com
$ export JIRA_API_TOKEN=y4bv87n4y5c845nyv84
$ export JIRA_BASE_URL=company.atlassian.net
$ git jira ASD-124
```

