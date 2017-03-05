package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Repository fields
type Repository struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	PathSSH  string `json:"ssh_url"`
	PathHTTP string `json:"clone_url"`
}

type simpleErrorJSON struct {
	Message string `json:"message"`
}

// sort flags and args
func flagsort() (flags, args []string) {
	if len(os.Args) < 2 {
		return nil, nil
	}
	for _, v := range os.Args[1:] {
		if strings.HasPrefix(v, "--") {
			v = strings.TrimPrefix(v, "-")
		}
		if strings.HasPrefix(v, "-") {
			flags = append(flags, strings.TrimPrefix(v, "-"))
		} else {
			args = append(args, v)
		}
	}
	sort.Strings(flags)
	sort.Strings(args)
	if !sort.StringsAreSorted(flags) || !sort.StringsAreSorted(args) {
		echo("error: can't sort flags/arguments.")
	}
	return
}

// easy echo
func echo(f string, s ...interface{}) {
	if s == nil {
		fmt.Fprintln(os.Stderr, f)
		return
	}
	fmt.Fprintf(os.Stderr, f, s...)
}
func help() {
	echo(version)
	echo("Clone all your github repositories")
	echo("source: https://github.com/aerth/gh-sync")
	echo("\nUSAGE")
	echo("\tUse command line args:")
	echo("\tgh-sync user=you token=123 /tmp/clones")
	echo("\tOr use environment:")
	echo("\tGITHUB_USER=you GITHUB_TOKEN=123 gh-sync")
	echo("\nFLAGS")
	echo("\t[-d, -dryrun] print repository info but don't clone")
	echo("\t[-443, -https] use https when cloning (default: ssh)")
	echo("\t[-v, -verbose] too much information")
	echo("\nARGS")
	echo("\t[1,2,3] print only one column: 1=name, 2=https, 3=ssh")
	echo("\t[user=] define github account name")
	echo("\t[token=] define github API token")
	echo("\t[dir=] clone destination (default: cwd)")
	echo("\nEXAMPLES")
	echo("\t'gh-sync -d 1' print names of all repositories")
	echo("\t'gh-sync -d 2' print https link to all repositories")
	echo("\t'gh-sync -d 3' print ssh link to all repositories")
	echo("\t'gh-sync dir=clones' clone all repos into ./clones/{reponame} ")
	echo("\t'gh-sync clones' clone all repos into ./clones/{reponame} ")

}

func updater(repo Repository, clonepath string) {
	if !useHTTPS {
		echo("[ssh update] %s\n", repo.Name)
	} else {
		echo("[https update] %s\n", repo.Name)
	}

	var cmd *exec.Cmd
	if useHTTPS {
		cmd = exec.Command("git", "pull", "origin", "master")
	} else {
		cmd = exec.Command("git", "pull", "origin", "master")
	}
	if !filepath.IsAbs(clonepath) {
		echo("Clone path is not absolute")
		os.Exit(1)
	}
	cmd.Dir = clonepath
	b, err := cmd.CombinedOutput()
	if err != nil {
		echo("Error! %s %s\n", err.Error(), string(b))
	}

	if verbose {
		echo(string(b))
		echo("Done with update: %s\n", repo.Name)
	}
}
func getArgs(args []string) {
	for _, v := range args {
		if strings.HasPrefix(v, "user=") {
			userOrOrganization = "user"
			accountName = strings.TrimPrefix(v, "user=")
			continue
		}

		if strings.HasPrefix(v, "org=") {
			userOrOrganization = "org"
			accountName = strings.TrimPrefix(v, "org=")
			continue
		}

		if strings.HasPrefix(v, "token=") {
			accessToken = strings.TrimPrefix(v, "token=")
			continue
		}

		if strings.HasPrefix(v, "dir=") {
			outputDir = strings.TrimPrefix(v, "dir=")
			continue
		}
	}
}

func getFlags(flags []string) {
	if flags != nil {
		for _, v := range flags {
			switch v {

			case "h", "help":
				help()
				os.Exit(0)
			case "u", "upgrade", "update", "f":
				updateRepositories = true

			case "d", "dryrun":
				dryrun = true

			case "v", "verbose":
				verbose = true

			case "https", "443":
				useHTTPS = true
			}
		}
	}
	if verbose {
		echo("[verbose]")
		if dryrun {
			echo("[dry run]")
		}
		if updateRepositories {
			echo("[updating]")
		}

	}

}

// clone a repository
func cloner(repo Repository, clonepath string, count int, done chan string) {
	defer func() { done <- repo.Name }()
	if !useHTTPS {
		echo("[ssh clone] %s\n", repo.Name)
	} else {
		echo("[https clone] %s\n", repo.Name)
	}

	var cmd *exec.Cmd
	if useHTTPS {
		cmd = exec.Command("git", "clone", "-v", repo.PathHTTP, clonepath+repo.Name)
	} else {
		cmd = exec.Command("git", "clone", "-v", repo.PathSSH, clonepath+repo.Name)
	}
	b, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(b), "already exists") {
			if verbose {
				echo("%q already exists.\n", repo.Name)
			}
			if updateRepositories {
				updater(repo, clonepath)
			}
		} else if verbose {
			echo("Error! %s %s\n", err.Error(), string(b))
		}
		return
	}
	if verbose {
		echo(string(b))
		echo("Done with clone %s (#%v)\n", repo.Name, count)
	}

}

// fetch and unmarshal
func githubFetchList(path string) []Repository {
	res, err := http.Get(path)
	if err != nil {
		echo(err.Error())
		os.Exit(1)
	}

	defer res.Body.Close()
	if res.StatusCode == 404 {
		echo("error 404", path)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		echo(err.Error())
		os.Exit(1)
	}

	repos := []Repository{}
	err = json.Unmarshal(b, &repos)
	if err != nil {

		errorJSON := new(simpleErrorJSON)
		if err = json.Unmarshal(b, &errorJSON); err != nil {
			echo(err.Error())
		} else {
			echo("error: " + errorJSON.Message)
		}

		os.Exit(1)
	}

	for _, r := range repos {
		if verbose {
			echo("Found repo %q: %q\n", r.Name, r.PathSSH)
		}
	}

	return repos
}
