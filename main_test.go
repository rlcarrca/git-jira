package main

import (
	"os"
	"os/exec"
	"reflect"
	"testing"

	"gopkg.in/andygrunwald/go-jira.v1"
)

func Test_createBranchName(t *testing.T) {
	type args struct {
		issueTitle string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "", args: args{"Android | Home Screen"}, want: "home_screen"},
		{name: "", args: args{"Android | Home || Screen"}, want: "home_screen"},
		{name: "", args: args{"| Android | Home Screen"}, want: "home_screen"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createBranchName(tt.args.issueTitle); got != tt.want {
				t.Errorf("createBranchName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_trimRules(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "removes android from string", args: args{"Home Screen | android |"}, want: "home screen |  |"},
		{name: "removes ios from string", args: args{"Home Screen | ios |"}, want: "home screen |  |"},
		{name: "trims spaces in string", args: args{"  "}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trimRules(tt.args.input); got != tt.want {
				t.Errorf("trimRules() = %v, want %v", got, tt.want)
			}
		})
	}
}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func Test_generateGitCommit(t *testing.T) {
	type args struct {
		message []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "single line commit message", args: args{[]string{"aaa"}}, want: []string{"git", "commit", "--allow-empty", "-m", "aaa"}},
		{name: "multiline commit message", args: args{[]string{"aaa", "bbb", "ccc"}}, want: []string{"git", "commit", "--allow-empty", "-m", "aaa", "-m", "bbb", "-m", "ccc"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateGitCommit(tt.args.message...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateGitCommit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateGitCheckout(t *testing.T) {
	type args struct {
		branch string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "", args: args{"aaa"}, want: []string{"git", "checkout", "-b", "aaa"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateGitCheckout(tt.args.branch); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateGitCheckout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getIssueType(t *testing.T) {
	type args struct {
		issue jira.IssueType
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Story issue maps to feature", args: args{jira.IssueType{Name: "Story"}}, want: "feature"},
		{name: "Bug issue maps to bug", args: args{jira.IssueType{Name: "Bug"}}, want: "bug"},
		{name: "Any issue maps to feature", args: args{jira.IssueType{Name: "Anything"}}, want: "feature"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getIssueType(tt.args.issue); got != tt.want {
				t.Errorf("getIssueType() = %v, want %v", got, tt.want)
			}
		})
	}
}
