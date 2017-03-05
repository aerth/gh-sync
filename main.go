package main

// Copyright 2017 aerth <aerth@riseup.net
// clone all your github repos (ssh or https)

import (
  "fmt"
  "sort"
  "os/exec"
  "os"
  "encoding/json"
  "strings"
  "net/http"
  "io/ioutil"
  "time"
)
var (
  verbose, dryrun, useHTTPS bool
  githubBaseURL, userOrOrganization, accountName, outputDir, accessToken string
  version string
)

// sort flags and args
func flagsort() (flags,args []string){
    if len(os.Args) < 2 {
      return nil,nil
    }
    for _,v := range os.Args[1:] {
      if strings.HasPrefix(v, "--"){
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
func init(){
  if version == "" {
    version = "gh-sync v10 (go get)"
  } else {
    version = "gh-sync v10-"+version
  }
  s, err := exec.LookPath("git")
  if err != nil {
    echo("fatal error: can't find 'git' in PATH (%q)\n", os.Getenv("PATH"))
    os.Exit(2)
  } else {
    if verbose{
    echo("using git: %q\n",s)
  }
  }

  accessToken = os.Getenv("GITHUB_TOKEN")
  accountName = os.Getenv("GITHUB_USER")
}

// easy echo
func echo(f string, s ...interface{}){
  if s == nil {
    fmt.Fprintln(os.Stderr, f)
    return
  }
  fmt.Fprintf(os.Stderr, f, s...)
}
func help(){
  echo(version)
  echo("Clone all your github repositories")
  echo("source: https://github.com/aerth/gh-sync")
  echo("\nUSAGE")
  echo("\tUse command line args:")
  echo("\tgh-sync user=you token=123 dir=/tmp/clones")
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

}

func getArgs(args []string) {
  for _, v := range args {
    if strings.HasPrefix(v, "user="){
      userOrOrganization = "user"
      accountName = strings.TrimPrefix(v, "user=")
      continue
    }

    if strings.HasPrefix(v, "org="){
      userOrOrganization = "org"
      accountName = strings.TrimPrefix(v, "org=")
      continue
    }

    if strings.HasPrefix(v, "token="){
      accessToken = strings.TrimPrefix(v, "token=")
      continue
    }
  }
}
var numarg int
func getFlags(flags []string){
  if flags != nil {
  for _, v := range flags {
  switch v {

  case "h", "help":
    help()
    os.Exit(0)

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

  }

}
func main(){

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
      numarg =2
    case "3":
      numarg =3
    }
  }

if userOrOrganization == "" {
  userOrOrganization = "user" // or "org"
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
    echo("gh-sync error: No API token. Visit https://github.com/settings/tokens/new and use token=XXXXXX or set GITHUB_TOKEN environmental variable")
    os.Exit(1)
  }
}
    github(githubBaseURL, userOrOrganization, accountName, accessToken, outputDir)
  os.Exit(0)
}

func github(apiHost, apiPath, apiUser, apiToken, clonepath string){
if clonepath==""{
  clonepath = "./"
}
if !strings.HasSuffix(clonepath, "/") {
  clonepath=clonepath+"/"
}

path := fmt.Sprintf("https://%s/%s/repos?access_token=%s&per_page=9000", apiHost, apiPath, apiToken)
repos := githubFetchList(path)
if verbose{
  echo("Found %v repositories.", len(repos))
  echo("Starting clone operation")
}

done := make(chan string, len(repos))
quit := make(chan int)

for i, repo := range repos {
  time.Sleep(500*time.Millisecond)
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



// wait to finish
for i:=0; i < len(repos); i++ {

  select {
      case name := <- done:
        echo("Clone finished: %q\n", name)
      case <- time.After(time.Minute*10):
        echo("Been ten minutes dude")
      case sig := <- quit:
        os.Exit(sig)
  }
}
// end main()
}

type Repository struct {
  ID int `json:"id"`
  Name string `json:"name"`
  PathSSH string `json:"ssh_url"`
  PathHTTP string `json:"clone_url"`
}

func cloner(repo Repository, clonepath string, count int, done chan string){
  if !useHTTPS {
    echo("[ssh clone] %s", repo.Name)
  } else {
    echo("[https clone] %s", repo.Name)
  }

  var cmd *exec.Cmd
  if useHTTPS {
    cmd = exec.Command("git", "clone", repo.PathHTTP, clonepath+repo.Name)
  } else {
    cmd = exec.Command("git", "clone", repo.PathSSH, clonepath+repo.Name)
  }
  b, err := cmd.CombinedOutput()
  if err != nil {
    echo("Error! %s %s\n",err.Error(), string(b))
    done <- repo.Name
    return
  }

  if verbose {
    echo(string(b))
  }
  fmt.Println("Dont with:", count, repo.Name, string(b))
  done <- repo.Name
}
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

  errorJSON := new(JSONerror)
  if err = json.Unmarshal(b, &errorJSON); err != nil {
    echo(err.Error())
  } else {
    echo("error: "+errorJSON.Message)
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

type JSONerror struct {
  Message string `json:"message"`
}
