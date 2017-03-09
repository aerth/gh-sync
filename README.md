# 'gh-sync'

## What it does

  * Download or List all repositories by a user
  * Gather repository information from github API
  * Extract JSON 'ssh-url' value from each repo
  * Use git program to clone each repository
  * Option '-d' lists repositories without cloning them
  * No api token required

## Install gh-sync

### Compile from source code

  1. ```go get -v -x -d -u github.com/aerth/gh-sync```
  2. ```cd $GOPATH/src/github.com/aerth/gh-sync```
  3. ```make```
  4. (install to '/usr/local/bin/' ) ```su -c 'make install'```

  Install to $HOME/bin with ```make install PREFIX=$HOME/bin/```

### Download binary for your architecture
  * https://github.com/aerth/gh-sync/releases/latest


![Screenshot of 'gh-sync'](https://github.com/aerth/'gh-sync'/blob/master/example/'gh-sync'-screenshot.png?raw=true)

## Author

Copyright (c) 2017, aerth <aerth@riseup.net>

Contributions are welcome, visit https://github.com/aerth/gh-sync and create a pull-request

MIT LICENSE
