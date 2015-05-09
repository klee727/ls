package main

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
	"time"
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

func gid_string_to_int(gid_str string) int {
	gid_num, err := strconv.ParseInt(gid_str, 10, 0)
	if err != nil {
		fmt.Printf("error:  couldn't convert %s to int\n", gid_str)
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	return int(gid_num)
}

func create_listing(info os.FileInfo, group_map map[int]string) Listing {

	var current_listing Listing

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
	gid := gid_string_to_int(owner.Gid)
	current_listing.group = group_map[gid]

	// size
	current_listing.size = fmt.Sprintf("%d", info.Size())

	// month
	current_listing.month = info.ModTime().Month().String()[0:3]

	// day
	current_listing.day = fmt.Sprintf("%02d", info.ModTime().Day())

	// time
	// if older than a year, print the year -- TODO:  revisit this
	// otherwise, print hour:minute

	epoch_now := time.Now().Unix()
	var seconds_in_a_year int64 = 365 * 24 * 60 * 60
	epoch_a_year_ago := epoch_now - seconds_in_a_year
	epoch_modified := info.ModTime().Unix()

	var time_str string
	if epoch_modified <= epoch_a_year_ago {
		time_str = fmt.Sprintf("%d", info.ModTime().Year())
	} else {
		time_str = fmt.Sprintf("%02d:%02d",
			info.ModTime().Hour(),
			info.ModTime().Minute())
	}

	current_listing.time = time_str

	current_listing.name = info.Name()

	return current_listing
}

func write_listings_to_buffer(output_buffer *bytes.Buffer,
	listings []Listing,
	option_all bool,
	option_long bool) {

	if option_long {
		var (
			width_permissions    int = 0
			width_num_hard_links int = 0
			width_owner          int = 0
			width_group          int = 0
			width_size           int = 0
			//width_month          int = 3
			//width_day            int = 2
			width_time int = 0
		)
		// check max widths for each field
		for _, l := range listings {
			if len(l.permissions) > width_permissions {
				width_permissions = len(l.permissions)
			}
			if len(l.num_hard_links) > width_num_hard_links {
				width_num_hard_links = len(l.num_hard_links)
			}
			if len(l.owner) > width_owner {
				width_owner = len(l.owner)
			}
			if len(l.group) > width_group {
				width_group = len(l.group)
			}
			if len(l.size) > width_size {
				width_size = len(l.size)
			}
			if len(l.time) > width_time {
				width_time = len(l.time)
			}
		}

		// now print the listings
		for _, l := range listings {
			// permissions
			output_buffer.WriteString(l.permissions)
			for i := 0; i < width_permissions-len(l.permissions); i++ {
				output_buffer.WriteString(" ")
			}
			output_buffer.WriteString(" ")

			// number of hard links (right justified)
			for i := 0; i < width_num_hard_links-len(l.num_hard_links); i++ {
				output_buffer.WriteString(" ")
			}
			for i := 0; i < 2-width_num_hard_links; i++ {
				output_buffer.WriteString(" ")
			}
			output_buffer.WriteString(l.num_hard_links)
			output_buffer.WriteString(" ")

			// owner
			output_buffer.WriteString(l.owner)
			for i := 0; i < width_owner-len(l.owner); i++ {
				output_buffer.WriteString(" ")
			}
			output_buffer.WriteString(" ")

			// group
			output_buffer.WriteString(l.group)
			for i := 0; i < width_group-len(l.group); i++ {
				output_buffer.WriteString(" ")
			}
			output_buffer.WriteString(" ")

			// size
			for i := 0; i < width_size-len(l.size); i++ {
				output_buffer.WriteString(" ")
			}
			output_buffer.WriteString(l.size)
			output_buffer.WriteString(" ")

			// month
			output_buffer.WriteString(l.month)
			output_buffer.WriteString(" ")

			// day
			output_buffer.WriteString(l.day)
			output_buffer.WriteString(" ")

			// time
			for i := 0; i < width_time-len(l.time); i++ {
				output_buffer.WriteString(" ")
			}
			output_buffer.WriteString(l.time)
			output_buffer.WriteString(" ")

			// name
			output_buffer.WriteString(l.name)
			output_buffer.WriteString("\n")
		}
		if output_buffer.Len() > 0 {
			output_buffer.Truncate(output_buffer.Len() - 1)
		}
	} else {
		if option_all {
			output_buffer.WriteString(". .. ")
		}

		for _, l := range listings {
			output_buffer.WriteString(l.name)
			output_buffer.WriteString(" ")
		}
		if output_buffer.Len() > 0 {
			output_buffer.Truncate(output_buffer.Len() - 1)
		}
	}
}

func ls(output_buffer *bytes.Buffer, args []string) {
	args_options := make([]string, 0)
	args_files := make([]string, 0)
	list_dirs := make([]os.FileInfo, 0)
	list_files := make([]os.FileInfo, 0)

	listings := make([]Listing, 0)

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

		gid := gid_string_to_int(line_split[2])
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
			l := create_listing(f, group_map)
			listings = append(listings, l)
		}
		write_listings_to_buffer(output_buffer,
			listings,
			option_all,
			option_long)
		listings = make([]Listing, 0)
	}

	//
	// then list the directories
	//
	if (num_files > 0 && num_dirs > 0) || (num_dirs > 1) {
		if num_files > 0 {
			output_buffer.WriteString("\n\n")
		}

		for _, d := range list_dirs {
			output_buffer.WriteString(d.Name() + ":\n")

			files_in_dir, err := ioutil.ReadDir(d.Name())
			if err != nil {
				fmt.Printf("error: %v\n", err)
				os.Exit(1)
			}

			for _, _f := range files_in_dir {
				if is_dot_name(_f) && !option_all {
					continue
				}

				l := create_listing(_f, group_map)
				listings = append(listings, l)
			}

			write_listings_to_buffer(output_buffer,
				listings,
				option_all,
				option_long)
			output_buffer.WriteString("\n\n")
			listings = make([]Listing, 0)
		}
		output_buffer.Truncate(output_buffer.Len() - 2)
	} else if num_dirs == 1 {
		for _, d := range list_dirs {
			files_in_dir, err := ioutil.ReadDir(d.Name())
			if err != nil {
				fmt.Printf("error: %v\n", err)
				os.Exit(1)
			}

			for _, _f := range files_in_dir {
				if is_dot_name(_f) && !option_all {
					continue
				}

				l := create_listing(_f, group_map)
				listings = append(listings, l)
			}
			write_listings_to_buffer(output_buffer,
				listings,
				option_all,
				option_long)
		}
	}
}

//
// main
//
func main() {
	var output_buffer bytes.Buffer

	ls(&output_buffer, os.Args[1:])

	fmt.Printf("%s\n", output_buffer.String())
}
