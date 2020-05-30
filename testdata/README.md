
# FAQ

## What are these patch.\* directories!?

These patch.\* directories are bare Git repos created
using commands like the following.

    $ mkdir patch.empty
    $ cd patch.empty
	$ git init --bare
	$ cd ..
	$ git clone patch.empty patch.empty.tmp
	$ cd patch.empty.tmp
	$ git commit --allow-empty -m start
	$ git push -u origin master

And to keep the repo minimal, since it is being committed,
the samples are removed.

    $ rm -f $(find ./patch.empty -name *.sample)
