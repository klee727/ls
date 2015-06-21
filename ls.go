package main

import (
	"bufio"
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Base set of color codes for colorized output
const (
	color_black  = "\x1b[0;30m"
	color_red    = "\x1b[0;31m"
	color_green  = "\x1b[0;32m"
	color_brown  = "\x1b[0;33m"
	color_blue   = "\x1b[0;34m"
	color_purple = "\x1b[0;35m"
	color_cyan   = "\x1b[0;36m"
	color_none   = "\x1b[0m"
)

// This a FileInfo paired with the original path as passed in to the program.
// Unfortunately, the Name() in FileInfo is only the basename, so the associated
// path must be manually recorded as well.
type FileInfoPath struct {
	path string
	info os.FileInfo
}

// This struct wraps all the option settings for the program into a single
// object.
type Options struct {
	all          bool
	long         bool
	human        bool
	one          bool
	dir          bool
	color        bool
	sort_reverse bool
	sort_time    bool
	sort_size    bool
}

// Listings contain all the information about a file or directory in a printable
// form.
type Listing struct {
	permissions    string
	num_hard_links string
	owner          string
	group          string
	size           string
	epoch_nano     int64
	month          string
	day            string
	time           string
	name           string
	linkname       string
}

// Global variables used by multiple functions
var (
	user_map  map[int]string // matches uid to username
	group_map map[int]string // matches gid to groupname
	options   Options        // the state of all program options
)

// Write the given Listing's name to the output buffer, with the appropriate
// formatting based on the current options.
func write_listing_name(output_buffer *bytes.Buffer, l Listing) {
	if options.color {
		applied_color := false
		if l.permissions[0] == 'd' {
			output_buffer.WriteString(color_blue)
			applied_color = true
		} else if l.permissions[0] == 'l' {
			output_buffer.WriteString(color_purple)
			applied_color = true
		} else if strings.Contains(l.permissions, "x") {
			output_buffer.WriteString(color_red)
			applied_color = true
		}

		output_buffer.WriteString(l.name)
		if applied_color {
			output_buffer.WriteString(color_none)
		}
	} else {
		output_buffer.WriteString(l.name)
	}

	if l.permissions[0] == 'l' && options.long {
		output_buffer.WriteString(fmt.Sprintf(" -> %s", l.linkname))
	}
}

// Convert a FileInfoPath object to a Listing.  The dirname is passed for
// following symlinks.
func create_listing(dirname string, fip FileInfoPath) (Listing, error) {
	var current_listing Listing

	// permissions string
	current_listing.permissions = fip.info.Mode().String()
	if fip.info.Mode()&os.ModeSymlink == os.ModeSymlink {
		current_listing.permissions = strings.Replace(
			current_listing.permissions, "L", "l", 1)

		var _pathstr string
		if dirname == "" {
			_pathstr = fmt.Sprintf("%s", fip.path)
		} else {
			_pathstr = fmt.Sprintf("%s/%s", dirname, fip.path)
		}
		link, err := os.Readlink(fmt.Sprintf(_pathstr))
		if err != nil {
			return current_listing, err
		}
		current_listing.linkname = link
	}

	sys := fip.info.Sys()

	stat, ok := sys.(*syscall.Stat_t)
	if !ok {
		return current_listing, fmt.Errorf("syscall failed\n")
	}

	// number of hard links
	num_hard_links := uint64(stat.Nlink)
	current_listing.num_hard_links = fmt.Sprintf("%d", num_hard_links)

	// owner
	owner, err := user.LookupId(fmt.Sprintf("%d", stat.Uid))
	if err != nil {
		// if this causes an error, use the manual user_map
		//
		// this can happen if go is built using cross-compilation for multiple
		// architectures (such as with Fedora Linux), in which case these
		// OS-specific features aren't implemented
		_owner := user_map[int(stat.Uid)]
		if _owner == "" {
			// if the user isn't in the map, just use the uid number
			current_listing.owner = fmt.Sprintf("%d", stat.Uid)
		} else {
			current_listing.owner = _owner
		}
	} else {
		current_listing.owner = owner.Username
	}

	// group
	_group := group_map[int(stat.Gid)]
	if _group == "" {
		// if the group isn't in the map, just use the gid number
		current_listing.group = fmt.Sprintf("%d", stat.Gid)
	} else {
		current_listing.group = _group
	}

	// size
	if options.human {
		size := float64(fip.info.Size())

		count := 0
		for size >= 1.0 {
			size /= 1024
			count++
		}

		if count < 0 {
			count = 0
		} else if count > 0 {
			size *= 1024
			count--
		}

		var suffix string
		if count == 0 {
			suffix = "B"
		} else if count == 1 {
			suffix = "K"
		} else if count == 2 {
			suffix = "M"
		} else if count == 3 {
			suffix = "G"
		} else if count == 4 {
			suffix = "T"
		} else if count == 5 {
			suffix = "P"
		} else if count == 6 {
			suffix = "E"
		} else {
			suffix = "?"
		}

		size_str := ""
		if count == 0 {
			size_b := int64(size)
			size_str = fmt.Sprintf("%d%s", size_b, suffix)
		} else {
			// looks like the printf formatting automatically rounds up
			size_str = fmt.Sprintf("%.1f%s", size, suffix)
		}

		// drop the trailing .0 if it exists in the size
		// e.g. 14.0K -> 14K
		if len(size_str) > 3 &&
			size_str[len(size_str)-3:len(size_str)-1] == ".0" {
			size_str = size_str[0:len(size_str)-3] + suffix
		}

		current_listing.size = size_str

	} else {
		current_listing.size = fmt.Sprintf("%d", fip.info.Size())
	}

	// epoch_nano
	current_listing.epoch_nano = fip.info.ModTime().UnixNano()

	// month
	current_listing.month = fip.info.ModTime().Month().String()[0:3]

	// day
	current_listing.day = fmt.Sprintf("%02d", fip.info.ModTime().Day())

	// time
	// if older than six months, print the year
	// otherwise, print hour:minute
	epoch_now := time.Now().Unix()
	var seconds_in_six_months int64 = 182 * 24 * 60 * 60
	epoch_six_months_ago := epoch_now - seconds_in_six_months
	epoch_modified := fip.info.ModTime().Unix()

	var time_str string
	if epoch_modified <= epoch_six_months_ago ||
		epoch_modified >= (epoch_now+5) {
		time_str = fmt.Sprintf("%d", fip.info.ModTime().Year())
	} else {
		time_str = fmt.Sprintf("%02d:%02d",
			fip.info.ModTime().Hour(),
			fip.info.ModTime().Minute())
	}

	current_listing.time = time_str

	current_listing.name = fip.path

	return current_listing, nil
}

// Comparison function used for sorting Listings by name.
func compare_name(a, b Listing) int {
	a_name_lower := strings.ToLower(a.name)
	b_name_lower := strings.ToLower(b.name)

	var smaller_len int
	if len(a.name) < len(b.name) {
		smaller_len = len(a.name)
	} else {
		smaller_len = len(b.name)
	}

	for i := 0; i < smaller_len; i++ {
		if a_name_lower[i] < b_name_lower[i] {
			return -1
		} else if a_name_lower[i] > b_name_lower[i] {
			return 1
		}
	}

	if len(a.name) < len(b.name) {
		return -1
	} else if len(b.name) < len(a.name) {
		return 1
	} else {
		return 0
	}
}

// Comparison function used for sorting Listings by modification time, from most
// recent to oldest.
func compare_time(a, b Listing) int {
	if a.epoch_nano >= b.epoch_nano {
		return -1
	}

	return 1
}

// Comparison function used for sorting Listings by size, from largest to
// smallest.
func compare_size(a, b Listing) int {
	a_size, _ := strconv.Atoi(a.size)
	b_size, _ := strconv.Atoi(b.size)

	if a_size >= b_size {
		return -1
	}

	return 1
}

// Sort the given listings, taking into account the current program options.
func sort_listings(listings []Listing) {
	comparison_function := compare_name
	if options.sort_time {
		comparison_function = compare_time
	} else if options.sort_size {
		comparison_function = compare_size
	}

	for {
		done := true
		for i := 0; i < len(listings)-1; i++ {
			a := listings[i]
			b := listings[i+1]

			if comparison_function(a, b) > -1 {
				tmp := a
				listings[i] = listings[i+1]
				listings[i+1] = tmp
				done = false
			}
		}
		if done {
			break
		}
	}

	if options.sort_reverse {
		middle_index := (len(listings) / 2)
		if len(listings)%2 == 0 {
			middle_index--
		}

		for i := 0; i <= middle_index; i++ {
			front_index := i
			rear_index := len(listings) - 1 - i

			if front_index == rear_index {
				break
			}

			tmp := listings[front_index]
			listings[front_index] = listings[rear_index]
			listings[rear_index] = tmp
		}
	}
}

// Create a set of Listings, comprised of the files and directories currently in
// the given directory.
func list_files_in_dir(dir Listing) ([]Listing, error) {
	l := make([]Listing, 0)

	if options.all {
		//info_dot, err := os.Stat(dir.path)
		info_dot, err := os.Stat(dir.name)
		if err != nil {
			return l, err
		}

		listing_dot, err := create_listing(dir.name,
			FileInfoPath{".", info_dot})
		if err != nil {
			return l, err
		}

		info_dotdot, err := os.Stat(dir.name + "/..")
		if err != nil {
			return l, err
		}

		listing_dotdot, err := create_listing(dir.name,
			FileInfoPath{"..", info_dotdot})
		if err != nil {
			return l, err
		}

		l = append(l, listing_dot)
		l = append(l, listing_dotdot)
	}

	files_in_dir, err := ioutil.ReadDir(dir.name)
	if err != nil {
		return l, err
	}

	for _, f := range files_in_dir {
		// if this is a .dotfile and '-a' is not specified, skip it
		if []rune(f.Name())[0] == rune('.') && !options.all {
			continue
		}

		_l, err := create_listing(dir.name,
			FileInfoPath{f.Name(), f})
		if err != nil {
			return l, err
		}
		l = append(l, _l)
	}

	sort_listings(l)

	return l, nil
}

// Given a set of Listings, print them to the output buffer, taking into account
// the current program arguments and terminal width as necessary.
func write_listings_to_buffer(output_buffer *bytes.Buffer,
	listings []Listing,
	terminal_width int) {

	if options.long {
		var (
			width_permissions    int = 0
			width_num_hard_links int = 0
			width_owner          int = 0
			width_group          int = 0
			width_size           int = 0
			width_time           int = 0
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
			write_listing_name(output_buffer, l)
			output_buffer.WriteString("\n")
		}
		if output_buffer.Len() > 0 {
			output_buffer.Truncate(output_buffer.Len() - 1)
		}
	} else if options.one {
		separator := "\n"

		for _, l := range listings {
			write_listing_name(output_buffer, l)
			output_buffer.WriteString(separator)
		}
		if output_buffer.Len() > 0 {
			output_buffer.Truncate(output_buffer.Len() - 1)
		}
	} else {
		separator := "  "

		// calculate the number of rows needed for column output
		num_rows := 1
		var col_widths []int
		for {
			num_cols := len(listings)/num_rows + len(listings)%num_rows

			col_widths = make([]int, num_cols)
			for i, _ := range col_widths {
				col_widths[i] = 0
			}

			// calculate necessary column widths
			for i := 0; i < len(listings); i++ {
				col := i / num_rows
				if col_widths[col] < len(listings[i].name) {
					col_widths[col] = len(listings[i].name)
				}
			}

			// calculate the maximum width of each row
			max_row_length := 0
			for i := 0; i < num_cols; i++ {
				max_row_length += col_widths[i]
			}
			max_row_length += len(separator) * (num_cols - 1)

			if max_row_length > terminal_width && num_rows >= len(listings) {
				break
			} else if max_row_length > terminal_width {
				num_rows++
			} else {
				break
			}
		}

		for r := 0; r < num_rows; r++ {
			for i, l := range listings {
				if i%num_rows == r {
					write_listing_name(output_buffer, l)
					for s := 0; s < col_widths[i/num_rows]-len(l.name); s++ {
						output_buffer.WriteString(" ")
					}
					output_buffer.WriteString(separator)
				}
			}
			if len(listings) > 0 {
				output_buffer.Truncate(output_buffer.Len() - len(separator))
			}
			output_buffer.WriteString("\n")
		}
		output_buffer.Truncate(output_buffer.Len() - 1)
	}
}

// Parse the program arguments and write the appropriate listings to the output
// buffer.
func ls(output_buffer *bytes.Buffer, args []string, width int) error {
	args_options := make([]string, 0)
	args_files := make([]string, 0)
	list_dirs := make([]Listing, 0)
	list_files := make([]Listing, 0)

	//
	// read in all the information from /etc/groups
	//
	group_map = make(map[int]string)

	group_file, err := os.Open("/etc/group")
	if err != nil {
		return fmt.Errorf("could not open /etc/group for reading\n")
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

		gid, err := strconv.ParseInt(line_split[2], 10, 0)
		if err != nil {
			return err
		}
		group_name := line_split[0]
		group_map[int(gid)] = group_name
	}

	//
	// read in all information from /etc/passwd for user lookup
	//
	user_map = make(map[int]string)

	user_file, err := os.Open("/etc/passwd")
	if err != nil {
		return fmt.Errorf("could not open /etc/passwd for reading\n")
	}

	reader = bufio.NewReader(user_file)
	scanner = bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Trim(line, " \t")

		if line[0] == '#' || line == "" {
			continue
		}

		line_split := strings.Split(line, ":")

		uid, err := strconv.ParseInt(line_split[2], 10, 0)
		if err != nil {
			return err
		}
		user_name := line_split[0]
		user_map[int(uid)] = user_name
	}

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
	options = Options{}
	options.color = true // use color by default
	for _, o := range args_options {

		// is it a short option '-' or a long option '--'?
		if strings.Contains(o, "--") {
			if strings.Contains(o, "--nocolor") {
				options.color = false
			}
		} else {
			if strings.Contains(o, "1") {
				options.one = true
			}
			if strings.Contains(o, "a") {
				options.all = true
			}
			if strings.Contains(o, "d") {
				options.dir = true
			}
			if strings.Contains(o, "h") {
				options.human = true
			}
			if strings.Contains(o, "l") {
				options.long = true
			}
			if strings.Contains(o, "r") {
				options.sort_reverse = true
			}
			if strings.Contains(o, "t") {
				options.sort_time = true
			}
			if strings.Contains(o, "S") {
				options.sort_size = true
			}
		}
	}

	// if no files are specified, list the current directory
	if len(args_files) == 0 {
		this_dir, _ := os.Lstat(".")
		//this_dir, _ := os.Stat(".")

		this_dir_listing, err := create_listing("",
			FileInfoPath{".", this_dir})
		if err != nil {
			return err
		}

		// for option_dir (-d), treat the '.' directory like a regular file
		if options.dir {
			list_files = append(list_files, this_dir_listing)
		} else { // else, treat '.' like a directory
			list_dirs = append(list_dirs, this_dir_listing)
		}
	}

	//
	// separate the files from the directories
	//
	for _, f := range args_files {
		//info, err := os.Stat(f)
		info, err := os.Lstat(f)

		if err != nil && os.IsNotExist(err) {
			return fmt.Errorf("cannot access %s: No such file or directory", f)
		} else if err != nil {
			return err
		}

		f_listing, err := create_listing("",
			FileInfoPath{f, info})
		if err != nil {
			return err
		}

		// for option_dir (-d), treat directories like regular files
		if options.dir {
			list_files = append(list_files, f_listing)
		} else { // else, separate the files and directories
			if info.IsDir() {
				list_dirs = append(list_dirs, f_listing)
			} else {
				list_files = append(list_files, f_listing)
			}
		}
	}

	num_files := len(list_files)
	num_dirs := len(list_dirs)

	// sort the lists if necessary
	sort_listings(list_files)
	sort_listings(list_dirs)

	//
	// list the files first
	//
	if num_files > 0 {
		write_listings_to_buffer(output_buffer,
			list_files,
			width)
	}

	//
	// then list the directories
	//
	if (num_files > 0 && num_dirs > 0) || (num_dirs > 1) {
		if num_files > 0 {
			output_buffer.WriteString("\n\n")
		}

		for _, d := range list_dirs {
			output_buffer.WriteString(d.name + ":\n")

			listings, err := list_files_in_dir(d)
			if err != nil {
				return err
			}

			if len(listings) > 0 {
				write_listings_to_buffer(output_buffer,
					listings,
					width)
				output_buffer.WriteString("\n\n")
			} else {
				output_buffer.WriteString("\n")
			}
		}

		output_buffer.Truncate(output_buffer.Len() - 2)
	} else if num_dirs == 1 {
		for _, d := range list_dirs {

			listings, err := list_files_in_dir(d)
			if err != nil {
				return err
			}

			write_listings_to_buffer(output_buffer,
				listings,
				width)
		}
	}

	return nil
}

// Main function
func main() {
	// capture the current terminal dimensions
	terminal_width, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Printf("error getting terminal dimensions\n")
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	var output_buffer bytes.Buffer

	err = ls(&output_buffer, os.Args[1:], terminal_width)
	if err != nil {
		fmt.Printf("ls: %v\n", err)
		os.Exit(1)
	}

	if output_buffer.String() != "" {
		fmt.Printf("%s\n", output_buffer.String())
	}
}

// vim: tabstop=4 softtabstop=4 shiftwidth=4 noexpandtab tw=80
