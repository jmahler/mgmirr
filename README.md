# NAME

rpmmirr - A tool for mirroring multiple RPM Git repos in to one.

[![GoDoc](https://godoc.org/github.com/jmahler/rpmmirr?status.svg)](http://godoc.org/github.com/jmahler/rpmmirr)
[![Test Status](https://github.com/jmahler/rpmmirr/workflows/Tests/badge.svg)](https://github.com/jmahler/rpmmirr/actions?query=workflow%3ATests)
[![Coverage Status](https://coveralls.io/repos/github/jmahler/rpmmirr/badge.svg?branch=master)](https://coveralls.io/github/jmahler/rpmmirr?branch=master)

# DESCRIPTION

A common situation with Git repos for RPM packages is that there
are multiple repos for different distros with similar contents.
CentOS is similar to Redhat is similar to Fedora, etc.

  https://src.fedoraproject.org/rpms/patch
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
    $ rpmmirr -C patch.rpm -c config.json -r patch
    $ cd patch.rpm/
    $ git branch
    [...]
    centos/c4
    centos/c5
    [...]
    fedora/f30
    fedora/f31

# AUTHOR

Jeremiah Mahler <jmmahler@gmail.com>

# LICENSE

Copyright &copy; 2020, Jeremiah Mahler.<br>
Released under the MIT License.
