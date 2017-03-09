package main

// Copyright 2017 aerth <aerth@riseup.net
// clone all your github repos (ssh or https)

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	verbose, dryrun, useHTTPS, updateRepositories bool
	githubBaseURL, userOrOrganization             string
	accountName, outputDir, accessToken           string
	version                                       string
)
var numarg int

func init() {
	if version == "" {
		version = "gh-sync v10 (go get)"
	} else {
		version = "gh-sync v10-" + version
	}
	s, err := exec.LookPath("git")
	if err != nil {
		echo("fatal error: can't find 'git' in PATH (%q)\n", os.Getenv("PATH"))
		os.Exit(2)
	} else {
		if verbose {
			echo("using git: %q\n", s)
		}
	}

	accessToken = os.Getenv("GITHUB_TOKEN")
	accountName = os.Getenv("GITHUB_USER")
}

func main() {

	flags, args := flagsort()
	getFlags(flags)
	getArgs(args)
	if verbose {
		echo("Flags: %q\n", flags)
		echo("Args: %q\n", args)
	}
	for _, arg := range args {
		switch arg {
		case "1":
			numarg = 1
		case "2":
			numarg = 2
		case "3":
			numarg = 3
		default:
			outputDir = arg

		}
	}

	if verbose && numarg != 0 {
		echo("[column] %s", numarg)
	}

	if githubBaseURL == "" {
		githubBaseURL = "api.github.com"
	}
	if accountName == "" {
		if accountName = os.Getenv("GITHUB_USER"); accountName == "" {
			echo("gh-sync error: No user name. Set $GITHUB_USER or use user=USERNAME")
			os.Exit(1)
		}
	}
	if accessToken == "" {
		if accessToken = os.Getenv("GITHUB_TOKEN"); accessToken == "" {
			//echo("gh-sync error: No API token. Visit https://github.com/settings/tokens/new and use token=XXXXXX or set GITHUB_TOKEN environmental variable")
			//os.Exit(1)
		}
	}
	if userOrOrganization == "" {
		if accessToken != "" {
			userOrOrganization = "user"
		} else {
			userOrOrganization = "users" // or "org"
		}
	}
	github(githubBaseURL, userOrOrganization, accountName, accessToken, outputDir)
	os.Exit(0)
}
func getwd() string {
	var err error
	var wd string
	wd, err = os.Getwd()
	if err != nil {
		echo(err.Error())
		os.Exit(1)
	}
	return wd
}
func github(apiHost, apiPath, apiUser, apiToken, clonepath string) {
	wd := getwd()
	if clonepath == "." || clonepath == "" {
		clonepath = wd
	} else {
		if !filepath.IsAbs(clonepath) {
			clonepath = wd + "/" + clonepath
		}
	}
	if !strings.HasSuffix(clonepath, "/") {
		clonepath = clonepath + "/"
	}
	var path string
	if apiToken != "" {
		path = fmt.Sprintf("https://%s/%s/repos?access_token=%s&per_page=9000", apiHost, apiPath, apiToken)

	} else {
		path = fmt.Sprintf("https://%s/users/%s/repos?per_page=9000", apiHost, apiUser)
	}
	repos := githubFetchList(path)
	if verbose {
		echo("Found %v repositories.\n", len(repos))
		if !dryrun {
			echo("Starting clone operation to directory: %q\n", clonepath)
		}
	}

	done := make(chan string, len(repos))
	quit := make(chan int)

	for i, repo := range repos {

		// print
		switch numarg {
		case 1:
			echo("%s\n", repo.Name)
		case 2:
			echo("%s\n", repo.PathHTTP)
		case 3:
			echo("%s\n", repo.PathSSH)
		default:
			echo("%s %s %s\n", repo.Name, repo.PathHTTP, repo.PathSSH)
		}
		if !dryrun {
			go cloner(repo, clonepath, i, done)
		}
	}

	if !dryrun {
		// wait to finish
		for i := 0; i < len(repos)-1; i++ {
			select {
			case name := <-done:
				echo("Clone finished: %q\n", name)
			case <-time.After(time.Minute * 10):
				echo("Been ten minutes dude")
			case sig := <-quit:
				os.Exit(sig)
			}
		}
	}
	// end main()
}
