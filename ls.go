package main

// option to list directories first
// handle environment variables:  ls ${HOME}; ls $HOME

//if fi.Mode() & os.ModeSymlink == os.ModeSymlink {

// markmini:ls mark$ ls -l /Users
// total 0
// drwxrwxrwt   9 root       wheel   306 Dec 26 23:16 Shared
// drwxr-xr-x+ 16 anastasia  staff   544 Jun 29  2014 anastasia
// drwxr-xr-x+ 73 mark       staff  2482 Apr 30 01:33 mark
//
// for formatting, looks like two spaces after the longest username/group
// need to take into account if there is an extra permission character
// username/group are left-justified
// sizes are right-justified, as are the modified times
// month is fixed with, day is right justified (when single digit)
//
// the main problem here, is that you need to full list of items before you
// can make formatting decisions.
// so do one pass to get the raw data
// then another pass, to gather the max widths of all the fields
// and then format appropriately
//
// this is pretty much the same for both long and short listings
//
// would be good to have a struct to capture all this information in strings

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

type Listing struct {
	permissions    string
	num_hard_links string
	owner          string
	group          string
	size           string
	month          string
	day            string
	time           string
	name           string
}

func is_dot_name(info os.FileInfo) bool {
	info_name_rune := []rune(info.Name())
	return (info_name_rune[0] == rune('.'))
}

func gid_string_to_int( gid_str string ) int {
    gid_num, err := strconv.ParseInt(gid_str, 10, 0)
	if err != nil {
		fmt.Printf("error:  couldn't convert %s to int\n", gid_str)
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	return int(gid_num)
}

func add_to_listings(info os.FileInfo, listings []*Listing, group_map map[int]string) []*Listing {

	current_listing := new(Listing)
	//fmt.Printf( "current_listing is type %T\n", current_listing)

	// permissions string
	current_listing.permissions = info.Mode().String()

	sys := info.Sys()
	if sys == nil {
		fmt.Printf("error:  sys is nil\n")
		os.Exit(1)
	}

	stat, ok := sys.(*syscall.Stat_t)
	if !ok {
		fmt.Printf("error:  not ok from *syscall.Stat_t\n")
		os.Exit(1)
	}

	// number of hard links
	num_hard_links := uint64(stat.Nlink)
	current_listing.num_hard_links = fmt.Sprintf("%d", num_hard_links)

	// owner
	owner, err := user.LookupId(fmt.Sprintf("%d", stat.Uid))
	if err != nil {
		fmt.Printf("error:  could not look up owner from id %d\n", stat.Uid)
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	current_listing.owner = owner.Username

	// group
    gid := gid_string_to_int( owner.Gid )
	current_listing.group = group_map[gid]

    // size
    current_listing.size = fmt.Sprintf( "%d", info.Size())

    // month
    current_listing.month = info.ModTime().Month().String()[0:3]

    // day
    current_listing.day = fmt.Sprintf( "%02d", info.ModTime().Day() )

    // time
    // if older than a year, print the year
    // otherwise, print hour:minute

	return append(listings, current_listing)
}

func ls(output_buffer *bytes.Buffer, args []string) {
	args_options := make([]string, 0)
	args_files := make([]string, 0)
	list_dirs := make([]os.FileInfo, 0)
	list_files := make([]os.FileInfo, 0)

	listings := make([]*Listing, 0)

	//
	// read in all the information from /etc/groups
	//
	group_map := make(map[int]string)

	group_file, err := os.Open("/etc/group")
	if err != nil {
		fmt.Printf("error:  couldn't open /etc/group for reading\n")
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	reader := bufio.NewReader(group_file)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Trim(line, " \t")

		if line[0] == '#' || line == "" {
			continue
		}

		line_split := strings.Split(line, ":")

        gid := gid_string_to_int( line_split[2] )
		group_name := line_split[0]
		group_map[int(gid)] = group_name
	}

	//fmt.Printf( "group_map = %v\n", group_map )

	//
	// parse arguments
	//
	for _, a := range args {
		a_rune := []rune(a)
		if a_rune[0] == '-' {
			// add to the options list
			args_options = append(args_options, a)
		} else {
			// add to the files/directories list
			args_files = append(args_files, a)
		}
	}

	//
	// parse options
	//
	option_all := false
	option_long := false
	for _, o := range args_options {
		if strings.Contains(o, "a") {
			option_all = true
		} else if strings.Contains(o, "l") {
			option_long = true
		}
	}

	// if no files are specified, list the current directory
	if len(args_files) == 0 {
		//this_dir, _ := os.Lstat(".")
		this_dir, _ := os.Stat(".")
		list_dirs = append(list_dirs, this_dir)
	}

	//
	// separate the files from the directories
	//
	for _, f := range args_files {
		info, err := os.Stat(f)

		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}

		if info.IsDir() {
			list_dirs = append(list_dirs, info)
		} else {
			list_files = append(list_files, info)
		}
	}

	num_files := len(list_files)
	num_dirs := len(list_dirs)

	//
	// list the files first
	//
	if num_files > 0 {
		for _, f := range list_files {
			listings = add_to_listings(f, listings, group_map)
		}
		output_buffer.Truncate(output_buffer.Len() - 1)
	}

	//
	// then list the directories
	//
	if (num_files > 0 && num_dirs > 0) || (num_dirs > 1) {
		if num_files > 0 {
			output_buffer.WriteString("\n")
		}

		for _, d := range list_dirs {
			output_buffer.WriteString(d.Name() + ":\n")

			// add '. .. ' to the output if -a is used
			if option_all && !option_long {
				output_buffer.WriteString(". .. ")
			}

			files_in_dir, err := ioutil.ReadDir(d.Name())
			if err != nil {
				fmt.Printf("error: %v\n", err)
				os.Exit(1)
			}

			for _, _f := range files_in_dir {
				if is_dot_name(_f) && !option_all {
					continue
				}

				//output_buffer.WriteString(_f.Name() + " ")
				listings = add_to_listings(_f, listings, group_map)
			}
			if output_buffer.Len() > 0 {
				output_buffer.Truncate(output_buffer.Len() - 1)
			}
			output_buffer.WriteString("\n\n")
		}
		output_buffer.Truncate(output_buffer.Len() - 2)
	} else if num_dirs == 1 {
		for _, d := range list_dirs {
			files_in_dir, err := ioutil.ReadDir(d.Name())
			if err != nil {
				fmt.Printf("error: %v\n", err)
				os.Exit(1)
			}

			// add '. .. ' to the output if -a is used
			if option_all {
				output_buffer.WriteString(". .. ")
				//added_dots = true
			}

			for _, _f := range files_in_dir {
				if is_dot_name(_f) && !option_all {
					continue
				}

				//output_buffer.WriteString(_f.Name() + " ")
				listings = add_to_listings(_f, listings, group_map)
			}
			if output_buffer.Len() > 0 {
				output_buffer.Truncate(output_buffer.Len() - 1)
			}
		}
	}

	//fmt.Printf( "listings = %v\n", listings )
	for _, l := range listings {
		fmt.Printf("l = %v\n", l)
	}

	//fmt.Printf("output_buffer.String() = |%s|\n", output_buffer.String())
}

//
// main
//
func main() {
	var output_buffer bytes.Buffer

	ls(&output_buffer, os.Args[1:])

	fmt.Printf("%s\n", output_buffer.String())
}
