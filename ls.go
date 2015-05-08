package main

// option to list directories first
// handle environment variables:  ls ${HOME}; ls $HOME

//if fi.Mode() & os.ModeSymlink == os.ModeSymlink {

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
)

func is_dot_name(info os.FileInfo) bool {
	info_name_rune := []rune(info.Name())
	return (info_name_rune[0] == rune('.'))
}

func add_to_output(output_buffer *bytes.Buffer, long bool, info os.FileInfo) {
	if long {
		var long_buffer bytes.Buffer
		// permissions string
		long_buffer.WriteString(info.Mode().String())
		long_buffer.WriteString(" ")

		// number of hard links
		sys := info.Sys()
		if sys != nil {
			stat, ok := sys.(*syscall.Stat_t)
			if ok {
				num_hard_links := uint64(stat.Nlink)
				str_hard_links := fmt.Sprintf("%2d ", num_hard_links)
				long_buffer.WriteString(str_hard_links)
			}
		}

		output_buffer.WriteString(long_buffer.String() + "\n")
	} else {
		output_buffer.WriteString(info.Name())
		output_buffer.WriteString(" ")
	}
}

func ls(output_buffer *bytes.Buffer, args []string) {
	args_options := make([]string, 0)
	args_files := make([]string, 0)
	list_dirs := make([]os.FileInfo, 0)
	list_files := make([]os.FileInfo, 0)

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
			add_to_output(output_buffer, option_long, f)
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
				add_to_output(output_buffer, option_long, _f)
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
				add_to_output(output_buffer, option_long, _f)
			}
			if output_buffer.Len() > 0 {
				output_buffer.Truncate(output_buffer.Len() - 1)
			}
		}
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
