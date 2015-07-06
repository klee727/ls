# ls

This is an implementation of the `ls` command in golang.

## Installation

With golang installed, simply use `go get` to download and build the program.

`$ go get github.com/reganm/ls`

Optionally, test cases can be run with `go test`.

`$ go test github.com/reganm/ls`

## Usage

`ls` will be installed at `${GOPATH}/bin/ls`.  Run the program with the `--help`
option to show usage information.

```
$ ${GOPATH}/bin/ls --help
usage:  ls [OPTIONS] [FILES]

OPTIONS:
    --dirs-first  list directories first
    --help        display usage information
    --nocolor     remove color formatting
    -1            one entry per line
    -a            include entries starting with '.'
    -d            list directories like files
    -h            list sizes with human-readable units
    -l            long listing
    -r            reverse any sorting
    -t            sort entries by modify time
    -S            sort entries by size
```

Only a commonly-used subset of the typical GNU or BSD `ls` options are
available.

## Color Output

Color output is enabled by default.  Use the `--nocolor` option to disable
colors.

This version of `ls` accepts either the BSD `LSCOLORS` environment variable 

`LSCOLORS=exfxcxdxbxegedabagacad`

or the GNU `LS_COLORS` environment variable

`LS_COLORS=rs=0:di=01;34:ln=01;36: ... and so on`

to determine the listings' color codes.  The variables are checked in the
following order:

1.  Use `LSCOLORS` if it is defined.
2.  Use `LS_COLORS` if it is defined.
3.  If neither are defined, use `LSCOLORS` with a default setting of
`exfxcxdxbxegedabagacad`.
