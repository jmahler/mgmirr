# NAME

rgm - A tool for mirroring multiple RPM Git repos in to one.

[![PkgGoDev](https://pkg.go.dev/badge/github.com/jmahler/rgm)](https://pkg.go.dev/github.com/jmahler/rgm)
[![Test Status](https://github.com/jmahler/rgm/workflows/Tests/badge.svg)](https://github.com/jmahler/rgm/actions?query=workflow%3ATests)
[![codecov](https://codecov.io/gh/jmahler/rgm/branch/main/graph/badge.svg)](https://codecov.io/gh/jmahler/rgm)

# DESCRIPTION

A common situation with Git repos for RPM packages is that there
are multiple repos for different distros with similar contents.
CentOS is similar to Redhat is similar to Fedora, etc.

  https://src.fedoraproject.org/rpms/patch<br>
  https://git.centos.org/rpms/patch

Because they are in different repos it is difficult to perform
Git operations (e.g. diff) that would be easy if they were in
one repo.  This tool mirrors these multiple git repos in to
one repo.


    $ cat config.json
    [...]
        "Remotes": [
            {
                "Name": "fedora",
                "URLs": ["https://src.fedoraproject.org/rpms/{{.RPM}}.git"]
            },
            {
                "Name": "centos",
                "URLs": ["https://git.centos.org/rpms/{{.RPM}}.git"]
            }
        ]
    [...]
    $ rgm -C patch.rpm -c config.json -r patch
    $ cd patch.rpm/
    $ git branch
    [...]
    centos/c4
    centos/c5
    [...]
    fedora/f30
    fedora/f31

# INSTALL HOWTO

This is a summary of the install steps that can also be found
by looking at the Github workflow (`.github/workflows/test.yml`).

The following steps were verified using a plain Ubuntu 18.04 Docker
container.  It is different than the Github Workflow because it doesn't
have all the extras included with workflows.

First, get a host or Docker container with Ubuntu 18.04.
<pre>
$ docker pull ubuntu:18.04
$ docker run -it ubuntu:18.04 /bin/bash
</pre>

Install the necessary system packages.
<pre>
$ apt install git golang cmake libssh2-1-dev libssl-dev zlib1g-dev libpcre3-dev
$ apt install tzdata
</pre>

Setup Go and build the packages.
<pre>
$ mkdir $HOME/go
$ export GOPATH="$HOME/go:/usr/share/gocode"
$ export GOBIN="$HOME/go/bin"

$ go get -d github.com/jmahler/rgm

$ go get -d github.com/libgit2/git2go
$ go get -d github.com/pborman/getopt/v2
$ go get -d golang.org/x/crypto/openpg

$ cd $HOME/go/src/github.com/libgit2/git2go/
$ make test-static
$ make install-static
</pre>

Build and install rgm.
<pre>
$ go build github.com/jmahler/rgm github.com/jmahler/rgm/rgm
$ go install github.com/jmahler/rgm/rgm
</pre>

Confirm that it can be run from the command line.
<pre>
$ ~/go/bin/rgm -h
Usage: rgm [-h] [-C value] [-c value] [-r value] [parameters ...]
 -C value  path to git repo for rpm
 -c value  config file (e.g. config.json)
 -h        help
 -r value  rpm name (e.g. patch)
</pre>

# AUTHOR

Jeremiah Mahler &lt;jmmahler@gmail.com&gt;

# LICENSE

Copyright &copy; 2020, Jeremiah Mahler.<br>
Released under the MIT License.
