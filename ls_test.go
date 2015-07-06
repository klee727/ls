package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"testing"
	"time"
)

// the root directory where individual test subdirectories are stored.
var test_root string

// default terminal width
const tw = 80

const default_LSCOLORS = "exfxcxdxbxegedabagacad"
const default_LS_COLORS = "rs=0:di=01;34:ln=01;36:mh=00:pi=40;33:so=01;35:" +
	"do=01;35:bd=40;33;01:cd=40;33;01:or=40;31;01:mi=01;05;37;41:su=37;41:" +
	"sg=30;43:ca=30;41:tw=30;42:ow=34;42:st=37;44:ex=01;32:*.tar=01;31:"

// change directory to the given path
func _cd(path string) {
	err := os.Chdir(path)
	if err != nil {
		fmt.Printf("error: os.Chdir(%s)\n", path)
		fmt.Printf("\t%v\n", err)
		os.Exit(1)
	}
}

// create a directory with the given path, using default ownership and 0755
// permissions
func _mkdir(path string) {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		fmt.Printf("error: os.MkdirAll(%s, 0755)\n", path)
		fmt.Printf("\t%v\n", err)
		os.Exit(1)
	}
}

// recursively remove the file or directory at the given path
func _rm(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("error: os.RemoveAll(%s)\n", path)
		fmt.Printf("\t%v\n", err)
		os.Exit(1)
	}
}

// create an empty file at the given path with default ownership and permissions
func _mkfile(path string) {
	_, err := os.Create(path)
	if err != nil {
		fmt.Printf("error: os.Create(%s)\n", path)
		fmt.Printf("\t%v\n", err)
		os.Exit(1)
	}
}

// create a symlink named path_linkname pointing to path_target
func _mklink(path_target string, path_linkname string) {
	err := os.Symlink(path_target, path_linkname)
	if err != nil {
		fmt.Printf("error: os.Symlink(%s, %s)\n", path_target, path_linkname)
		fmt.Printf("\t%v\n", err)
		os.Exit(1)
	}
}

// Perform the necessary Chmod, Chown, and Chtimes on the given path
func _modify_path(path string,
	mode os.FileMode,
	uid int,
	gid int,
	mod_epoch_s time.Time) {

	err := os.Chmod(path, mode)
	if err != nil {
		fmt.Printf("error: os.Chmod(%s, %d)\n", path, mode)
		fmt.Printf("\t%v\n", err)
		os.Exit(1)
	}

	err = os.Chown(path, uid, gid)
	if err != nil {
		fmt.Printf("error: os.Chmod(%s, %d)\n", path, mode)
		fmt.Printf("\t%v\n", err)
		os.Exit(1)
	}

	err = os.Chtimes(path, mod_epoch_s, mod_epoch_s)
	if err != nil {
		fmt.Printf("error: os.Chtimes(%s, %v, %v)\n", path, mod_epoch_s,
			mod_epoch_s)
		fmt.Printf("\t%v\n", err)
		os.Exit(1)
	}
}

// create a file at the given path with specific permissions, ownership, size,
// and modified time
func _mkfile2(path string,
	//num_hard_links int,
	mode os.FileMode,
	uid int,
	gid int,
	size_bytes int,
	mod_epoch_s time.Time) {

	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("error: os.Create(%s)\n", path)
		fmt.Printf("\t%v\n", err)
		os.Exit(1)
	}

	writer := bufio.NewWriter(f)
	for i := 0; i < size_bytes; i++ {
		_, err := writer.WriteString(" ")
		if err != nil {
			fmt.Printf("error: writer.WriteString(\" \")\n")
			fmt.Printf("\t%v\n", err)
			os.Exit(1)
		}
	}
	err = writer.Flush()
	if err != nil {
		fmt.Printf("error: writer.Flush()\n")
		fmt.Printf("\t%v\n", err)
		os.Exit(1)
	}

	_modify_path(path, mode, uid, gid, mod_epoch_s)
}

// create a directory at the given path with specific permissions, ownership,
// and modified time
func _mkdir2(path string,
	mode os.FileMode,
	uid int,
	gid int,
	mod_epoch_s time.Time) {

	err := os.Mkdir(path, mode)
	if err != nil {
		fmt.Printf("error: os.Mkdir(%s, %v)\n", path, mode)
		fmt.Printf("\t%v\n", err)
		os.Exit(1)
	}

	_modify_path(path, mode, uid, gid, mod_epoch_s)
}

// change to the test_root, create a directory for the test, and change to that
// directory
func setup_test_dir(path string) {
	os.Setenv("LSCOLORS", "")
	os.Setenv("LS_COLORS", "")
	_cd(test_root)
	_mkdir(path)
	_cd(path)
}

// fail the given test if the output and expected strings do not match
func check_output(t *testing.T, output, expected string) {
	if output != expected {
		t.Logf("\nexpected:\n\"%s\"\n\nbut got:\n\"%s\"\n", expected, output)
		t.Fail()
	}
}

func check_error(t *testing.T, err error, expected string) {
	if fmt.Sprintf("%v", err) != expected {
		t.Logf("check_error:\nexpected:\n\"%s\"\n\nbut got:\n\"%v\"\n",
			expected, err)
		t.Fail()
	}
}

func check_error_nil(t *testing.T, err error) {
	if err != nil {
		t.Logf("error is not nil\n")
		t.Fail()
	}
}

// remove any consecutive spaces in the given bytes.Buffer, and return the
// sanitized string
func clean_output_buffer(buffer bytes.Buffer) string {
	output := strings.TrimSpace(buffer.String())
	output_clean := make([]uint8, 0)

	prev_space := false
	for i := 0; i < len(output); i++ {
		if output[i] == ' ' && !prev_space {
			prev_space = true
			output_clean = append(output_clean, output[i])
		} else if output[i] == ' ' && prev_space {
			continue
		} else {
			prev_space = false
			output_clean = append(output_clean, output[i])
		}
	}

	return string(output_clean)
}

// main test method that sets up the test environment and launches the tests
func TestMain(m *testing.M) {

	//
	// setup
	//

	// set up the test root temporary directory
	var err error
	test_root, err = ioutil.TempDir("", "ls_")
	if err != nil {
		fmt.Printf("error:  couldn't create test_root '%s'\n", test_root)
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	//
	// run the tests
	//
	result := m.Run()

	//
	// teardown
	//
	_rm(test_root)

	os.Exit(result)
}

// Test running 'ls' in an empty directory
func Test_None_None_Empty(t *testing.T) {
	setup_test_dir("None_None_Empty")

	var output_buffer bytes.Buffer
	args := []string{}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := ""

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls' in a directory with files
func Test_None_None_Files(t *testing.T) {
	setup_test_dir("None_None_Files")

	_mkfile("a")
	_mkfile("b")
	_mkfile("c")

	var output_buffer bytes.Buffer
	args := []string{}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := "a b c"

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls' in a directory with files names of mixed case--check that
// case does not disrupt the default alphabetic sorting
func Test_None_None_Files2(t *testing.T) {
	setup_test_dir("None_None_Files2")

	_mkfile("a")
	_mkfile("B")
	_mkfile("c")

	var output_buffer bytes.Buffer
	args := []string{}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := "a B c"

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls' in a directory with a file and a symlink
func Test_None_None_Link(t *testing.T) {
	setup_test_dir("None_None_Link")

	_mkfile("a")
	_mklink("a", "b")

	var output_buffer bytes.Buffer
	args := []string{"--nocolor"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := "a b"

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls' in a directory with .files
func Test_None_None_DotFiles(t *testing.T) {
	setup_test_dir("None_None_DotFiles")

	_mkfile(".a")
	_mkfile(".b")
	_mkfile(".c")

	var output_buffer bytes.Buffer
	args := []string{}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := ""

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls a' when the file 'a' does not exist
func Test_None_File_Empty(t *testing.T) {
	setup_test_dir("None_File_Empty")

	var output_buffer bytes.Buffer
	args := []string{"a"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected_out := ""
	expected_err := "cannot access a: no such file or directory"

	check_output(t, output, expected_out)
	check_error(t, err, expected_err)
}

// Test running 'ls a' with a single 'a' file in the current directory
func Test_None_File_Files(t *testing.T) {
	setup_test_dir("None_File_Files")

	_mkfile("a")

	var output_buffer bytes.Buffer
	args := []string{"a"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := "a"

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls dir' with .files in that dir
func Test_None_Dir_DotFilesInDir(t *testing.T) {
	setup_test_dir("None_Dir_DotFilesInDir")

	dir2 := "dir"
	_mkdir(dir2)

	_mkfile(dir2 + "/.a")
	_mkfile(dir2 + "/.b")
	_mkfile(dir2 + "/.c")

	var output_buffer bytes.Buffer
	var args []string
	args = append(args, dir2)
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := ""

	check_output(t, output, expected)
	check_error_nil(t, err)
}

func Test_None_Dir_DirPerms(t *testing.T) {
	setup_test_dir("None_Dir_DirPerms")

	//_mkdir2("test_dir", 0755, os.Getuid(), os.Getuid(), time.Now())
	_mkdir("test_dir")
	_mkfile("test_dir/a")
	_modify_path("test_dir", 0000, os.Getuid(), os.Getuid(), time.Now())

	var output_buffer bytes.Buffer
	args := []string{"test_dir/a"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	// reset test_dir permissions so the directory can be deleted
	_modify_path("test_dir", 0755, os.Getuid(), os.Getuid(), time.Now())

	expected := ""

	check_output(t, output, expected)
	check_error(t, err, "open test_dir/a: permission denied")
}

// Test running 'ls -a' in an empty directory
func Test_a_None_Empty(t *testing.T) {
	setup_test_dir("a_None_Empty")

	var output_buffer bytes.Buffer
	args := []string{"-a", "--nocolor"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := ". .."

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -a' with .files in the current directory
func Test_a_None_DotFiles(t *testing.T) {
	setup_test_dir("a_None_DotFiles")

	_mkfile(".a")
	_mkfile(".b")
	_mkfile(".c")

	var output_buffer bytes.Buffer
	args := []string{"-a", "--nocolor"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := ". .. .a .b .c"

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -a ..' with dotfiles in the parent directory
func Test_a_Dir_DotFilesInParentDir(t *testing.T) {
	setup_test_dir("a_Dir_DotFilesInParentDir")

	_mkfile(".a")
	_mkfile(".b")
	_mkfile(".c")

	dir2 := "dir2"
	_mkdir(dir2)
	_cd(dir2)

	var output_buffer bytes.Buffer
	args := []string{"-a", "..", "--nocolor"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := ". .. .a .b .c dir2"

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -a . ..' with .files in the parent, current, and child
// directories
func Test_a_Dirs_DotFilesInParentDir2(t *testing.T) {
	setup_test_dir("a_Dirs_DotFilesInParentDir2")

	_mkfile(".a")
	_mkfile(".b")
	_mkfile(".c")

	dir2 := "dir2"
	_mkdir(dir2)
	_cd(dir2)

	_mkfile(".e")
	_mkfile(".f")
	_mkfile(".g")

	dir3 := "dir3"
	_mkdir(dir3)
	_cd(dir3)

	_mkfile(".h")
	_mkfile(".i")
	_mkfile(".j")

	_cd("..")

	var output_buffer bytes.Buffer
	args := []string{"-a", ".", "..", "--nocolor", dir3}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := ".:\n" +
		". .. .e .f .g dir3\n" +
		"\n" +
		"..:\n" +
		". .. .a .b .c dir2\n" +
		"\n" +
		"dir3:\n" +
		". .. .h .i .j"

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -l' in an empty directory
func Test_l_None_Empty(t *testing.T) {
	setup_test_dir("l_None_Empty")

	var output_buffer bytes.Buffer
	args := []string{"-l"}
	//err := ls(&output_buffer, args, tw)
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := ""

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -l a' with a single 'a' file in the current directory
func Test_l_File_File(t *testing.T) {
	setup_test_dir("l_File_File")

	time_now := time.Now()
	size := 13
	path := "a"
	_mkfile2(path, 0600, os.Getuid(), os.Getgid(), size, time_now)

	var output_buffer bytes.Buffer
	args := []string{"-l", "a"}
	ls_err := ls(&output_buffer, args, tw)

	output := clean_output_buffer(output_buffer)

	var owner string
	owner_lookup, err := user.LookupId(fmt.Sprintf("%d", os.Getuid()))
	if err != nil {
		owner = user_map[int(os.Getuid())]
	} else {
		owner = owner_lookup.Username
	}

	group := group_map[os.Getgid()]

	expected := fmt.Sprintf("-rw------- 1 %s %s %d %s %02d %02d:%02d %s",
		owner,
		group,
		size,
		time_now.Month().String()[0:3],
		time_now.Day(),
		time_now.Hour(),
		time_now.Minute(),
		path)

	check_output(t, output, expected)
	check_error_nil(t, ls_err)
}

// Test running 'ls -l a' with a symlink 'b' pointing to a file 'a'
func Test_l_File_Link(t *testing.T) {
	setup_test_dir("l_File_Link")

	time_now := time.Now()
	size := 13
	path := "a"
	_mkfile2(path, 0600, os.Getuid(), os.Getgid(), size, time_now)

	_mklink(path, "b")

	var output_buffer bytes.Buffer
	args := []string{"-l", "b", "--nocolor"}
	ls_err := ls(&output_buffer, args, tw)

	output := clean_output_buffer(output_buffer)
	// remove the permissions string from the output
	output_noperms := strings.Join(strings.Split(output, " ")[1:], " ")

	var owner string
	owner_lookup, err := user.LookupId(fmt.Sprintf("%d", os.Getuid()))
	if err != nil {
		owner = user_map[int(os.Getuid())]
	} else {
		owner = owner_lookup.Username
	}

	group := group_map[os.Getgid()]

	expected := fmt.Sprintf("1 %s %s 1 %s %02d %02d:%02d b -> %s",
		owner,
		group,
		time_now.Month().String()[0:3],
		time_now.Day(),
		time_now.Hour(),
		time_now.Minute(),
		path)

	check_output(t, output_noperms, expected)
	check_error_nil(t, ls_err)
}

// Test running 'ls -l' with a symlink 'b' pointing to a file 'a'
func Test_l_File_Link2(t *testing.T) {
	setup_test_dir("l_File_Link2")

	time_now := time.Now()
	size := 13
	path := "a"
	_mkfile2(path, 0600, os.Getuid(), os.Getgid(), size, time_now)

	_mklink(path, "b")

	var output_buffer bytes.Buffer
	args := []string{"-l", "--nocolor"}
	ls_err := ls(&output_buffer, args, tw)

	output := clean_output_buffer(output_buffer)

	// remove the permissions string from each file listing
	output_line_split := strings.Split(output, "\n")
	output_noperms := ""
	for _, l := range output_line_split {
		output_line_split_noperms := strings.Split(l, " ")[1:]
		output_line_noperms := strings.Join(output_line_split_noperms, " ")
		if len(output_noperms) == 0 {
			output_noperms = output_line_noperms
		} else {
			output_noperms = output_noperms + "\n" + output_line_noperms
		}
	}

	var owner string
	owner_lookup, err := user.LookupId(fmt.Sprintf("%d", os.Getuid()))
	if err != nil {
		owner = user_map[int(os.Getuid())]
	} else {
		owner = owner_lookup.Username
	}

	group := group_map[os.Getgid()]

	expected := fmt.Sprintf(
		"1 %s %s %d %s %02d %02d:%02d a\n1 %s %s 1 %s %02d %02d:%02d b -> %s",
		owner,
		group,
		size,
		time_now.Month().String()[0:3],
		time_now.Day(),
		time_now.Hour(),
		time_now.Minute(),
		owner,
		group,
		time_now.Month().String()[0:3],
		time_now.Day(),
		time_now.Hour(),
		time_now.Minute(),
		path)

	check_output(t, output_noperms, expected)
	check_error_nil(t, ls_err)
}

// Test running 'ls -d' in an empty directory
func Test_d_None_Empty(t *testing.T) {
	setup_test_dir("d_None_Empty")

	var output_buffer bytes.Buffer
	args := []string{"-d", "--nocolor"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := "."

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -d' in a directory with files
func Test_d_None_Files(t *testing.T) {
	setup_test_dir("d_None_Files")

	_mkfile("a")
	_mkfile("b")
	_mkfile("c")

	var output_buffer bytes.Buffer
	args := []string{"-d", "--nocolor"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := "."

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -d a b c' where a and c are files, but b is a directory
func Test_d_FilesAndDirs_FilesAndDirs(t *testing.T) {
	setup_test_dir("d_FilesAndDirs_FilesAndDirs")

	_mkfile("a")
	_mkdir("b")
	_mkfile("c")

	var output_buffer bytes.Buffer
	args := []string{"-d", "a", "b", "c", "--nocolor"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := "a b c"

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -ad' in a directory with files
func Test_ad_None_Files(t *testing.T) {
	setup_test_dir("adl_None_Files")

	_mkfile("a")
	_mkdir("b")
	_mkfile("a")

	var output_buffer bytes.Buffer
	args := []string{"-ad", "--nocolor"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := "."

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -1' in an empty directory
func Test_1_None_Empty(t *testing.T) {
	setup_test_dir("1_None_Empty")

	var output_buffer bytes.Buffer
	args := []string{"-1"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := ""

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -1' in a directory with a few files
func Test_1_None_Files(t *testing.T) {
	setup_test_dir("1_None_Files")

	_mkfile("a")
	_mkfile("b")
	_mkfile("c")
	_mkfile("d")
	_mkfile("e")

	var output_buffer bytes.Buffer
	args := []string{"-1"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := "a\nb\nc\nd\ne"

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -r' in an empty directory
func Test_r_None_None(t *testing.T) {
	setup_test_dir("r_None_None")

	var output_buffer bytes.Buffer
	args := []string{"-r"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := ""

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -r' in a directory with a few files
func Test_r_None_Files(t *testing.T) {
	setup_test_dir("r_None_Files")

	_mkfile("a")
	_mkfile("b")
	_mkfile("c")
	_mkfile("d")
	_mkfile("e")

	var output_buffer bytes.Buffer
	args := []string{"-r"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := "e d c b a"

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -t' in an empty directory
func Test_t_None_None(t *testing.T) {
	setup_test_dir("t_None_None")

	var output_buffer bytes.Buffer
	args := []string{"-t"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := ""

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -t' in a directory with a few files with different modified
// times
func Test_t_None_Files(t *testing.T) {
	setup_test_dir("t_None_Files")

	time_now := time.Now()
	_mkfile2("e", 0600, os.Getuid(), os.Getgid(), 0,
		time_now)
	_mkfile2("b", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-1*time.Second))
	_mkfile2("d", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-2*time.Second))
	_mkfile2("c", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-3*time.Second))
	_mkfile2("f", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-4*time.Second))
	_mkfile2("a", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-5*time.Second))

	var output_buffer bytes.Buffer
	args := []string{"-t"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := "e b d c f a"

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -t' in a directory with files, all sharing the same
// modification time.
func Test_t_None_FilesSameModifyTime(t *testing.T) {
	setup_test_dir("t_None_FilesSameModifyTime")

	time_now := time.Now()

	_mkfile2("a", 0600, os.Getuid(), os.Getgid(), 0, time_now)
	_mkfile2("b", 0600, os.Getuid(), os.Getgid(), 0, time_now)
	_mkfile2("c", 0600, os.Getuid(), os.Getgid(), 0, time_now)

	var output_buffer bytes.Buffer
	args := []string{"-t"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := "a b c"

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -t' on multiple files and directories.
func Test_t_FilesAndDirs_FilesAndDirs(t *testing.T) {
	setup_test_dir("t_FilesAndDirs_FilesAndDirs")

	time_now := time.Now()

	_mkfile2("b", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-1*time.Second))
	_mkfile2("a", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-5*time.Second))

	_mkdir("dir0")
	_mkfile2("dir0/minus_one", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-1*time.Second))
	_mkfile2("dir0/zero", 0600, os.Getuid(), os.Getgid(), 0, time_now)

	_mkdir("dir2")
	_mkfile2("dir2/e", 0600, os.Getuid(), os.Getgid(), 0,
		time_now)
	_mkfile2("dir2/b", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-1*time.Second))
	_mkfile2("dir2/d", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-2*time.Second))
	_mkfile2("dir2/c", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-3*time.Second))
	_mkfile2("dir2/f", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-4*time.Second))
	_mkfile2("dir2/a", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-5*time.Second))

	_mkdir("dir1")

	_modify_path("dir0",
		0755,
		os.Getuid(),
		os.Getgid(),
		time_now)
	_modify_path("dir1",
		0755,
		os.Getuid(),
		os.Getgid(),
		time_now.Add(-5*time.Second))
	_modify_path("dir2",
		0755,
		os.Getuid(),
		os.Getgid(),
		time_now.Add(-2*time.Second))

	var output_buffer bytes.Buffer
	args := []string{"-t", "a", "b", "dir0", "dir1", "dir2", "--nocolor"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := "b a\n\n" +
		"dir0:\n" +
		"zero minus_one\n\n" +
		"dir2:\n" +
		"e b d c f a\n\n" +
		"dir1:"

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -rt' in a directory with a few files with different modified
// times
func Test_rt_None_Files(t *testing.T) {
	setup_test_dir("rt_None_Files")

	time_now := time.Now()
	_mkfile2("e", 0600, os.Getuid(), os.Getgid(), 0,
		time_now)
	_mkfile2("b", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-1*time.Second))
	_mkfile2("d", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-2*time.Second))
	_mkfile2("c", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-3*time.Second))
	_mkfile2("f", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-4*time.Second))
	_mkfile2("a", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-5*time.Second))

	var output_buffer bytes.Buffer
	args := []string{"-rt"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := "a f c d b e"

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -h' in an empty directory.
func Test_h_None_None(t *testing.T) {
	setup_test_dir("h_None_None")

	var output_buffer bytes.Buffer
	args := []string{"-h"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := ""

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -h' in a directory with one file.
func Test_h_None_File(t *testing.T) {
	setup_test_dir("h_None_File")

	_mkfile("a")

	var output_buffer bytes.Buffer
	args := []string{"-h"}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := "a"

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test running 'ls -lh' with a single 13-byte 'a' file in the current
// directory.
func Test_lh_None_File(t *testing.T) {
	setup_test_dir("lh_None_File")

	time_now := time.Now()
	size := 13
	path := "a"
	_mkfile2(path, 0600, os.Getuid(), os.Getgid(), size, time_now)

	var output_buffer bytes.Buffer
	args := []string{"-lh"}
	ls_err := ls(&output_buffer, args, tw)

	output := clean_output_buffer(output_buffer)

	var owner string
	owner_lookup, err := user.LookupId(fmt.Sprintf("%d", os.Getuid()))
	if err != nil {
		owner = user_map[int(os.Getuid())]
	} else {
		owner = owner_lookup.Username
	}

	group := group_map[os.Getgid()]

	expected := fmt.Sprintf("-rw------- 1 %s %s %dB %s %02d %02d:%02d %s",
		owner,
		group,
		size,
		time_now.Month().String()[0:3],
		time_now.Day(),
		time_now.Hour(),
		time_now.Minute(),
		path)

	check_output(t, output, expected)
	check_error_nil(t, ls_err)
}

// Test running 'ls -lh' with a single 1024-byte (1K) 'a' file in the current
// directory.
func Test_lh_None_File2(t *testing.T) {
	setup_test_dir("lh_None_File2")

	time_now := time.Now()
	size := 1024
	path := "a"
	_mkfile2(path, 0600, os.Getuid(), os.Getgid(), size, time_now)

	var output_buffer bytes.Buffer
	args := []string{"-lh"}
	ls_err := ls(&output_buffer, args, tw)

	output := clean_output_buffer(output_buffer)

	var owner string
	owner_lookup, err := user.LookupId(fmt.Sprintf("%d", os.Getuid()))
	if err != nil {
		owner = user_map[int(os.Getuid())]
	} else {
		owner = owner_lookup.Username
	}

	group := group_map[os.Getgid()]

	expected := fmt.Sprintf("-rw------- 1 %s %s 1K %s %02d %02d:%02d %s",
		owner,
		group,
		time_now.Month().String()[0:3],
		time_now.Day(),
		time_now.Hour(),
		time_now.Minute(),
		path)

	check_output(t, output, expected)
	check_error_nil(t, ls_err)
}

// Test running 'ls -lh' with a single 1485-byte (1.45K) 'a' file in the current
// directory.
func Test_lh_None_File3(t *testing.T) {
	setup_test_dir("lh_None_File3")

	time_now := time.Now()
	size := 1485
	path := "a"
	_mkfile2(path, 0600, os.Getuid(), os.Getgid(), size, time_now)

	var output_buffer bytes.Buffer
	args := []string{"-lh"}
	ls_err := ls(&output_buffer, args, tw)

	output := clean_output_buffer(output_buffer)

	var owner string
	owner_lookup, err := user.LookupId(fmt.Sprintf("%d", os.Getuid()))
	if err != nil {
		owner = user_map[int(os.Getuid())]
	} else {
		owner = owner_lookup.Username
	}

	group := group_map[os.Getgid()]

	expected := fmt.Sprintf("-rw------- 1 %s %s 1.5K %s %02d %02d:%02d %s",
		owner,
		group,
		time_now.Month().String()[0:3],
		time_now.Day(),
		time_now.Hour(),
		time_now.Minute(),
		path)

	check_output(t, output, expected)
	check_error_nil(t, ls_err)
}

// Test running 'ls -S' in an empty directory.
func Test_S_None_None(t *testing.T) {
	setup_test_dir("S_None_None")

	var output_buffer bytes.Buffer
	args := []string{"-S"}
	ls_err := ls(&output_buffer, args, tw)

	output := clean_output_buffer(output_buffer)

	expected := ""

	check_output(t, output, expected)
	check_error_nil(t, ls_err)
}

// Test running 'ls -S' in a directory with two files of the same size.
func Test_S_None_Files(t *testing.T) {
	setup_test_dir("S_None_Files")

	_mkfile2("a", 0600, os.Getuid(), os.Getgid(), 5, time.Now())
	_mkfile2("b", 0600, os.Getuid(), os.Getgid(), 5, time.Now())

	var output_buffer bytes.Buffer
	args := []string{"-S"}
	ls_err := ls(&output_buffer, args, tw)

	output := clean_output_buffer(output_buffer)

	expected := "a b"

	check_output(t, output, expected)
	check_error_nil(t, ls_err)
}

// Test running 'ls -S' in a directory with files of the differing sizes.
func Test_S_None_Files2(t *testing.T) {
	setup_test_dir("S_None_Files2")

	_mkfile2("a", 0600, os.Getuid(), os.Getgid(), 0, time.Now())
	_mkfile2("b", 0600, os.Getuid(), os.Getgid(), 5, time.Now())
	_mkfile2("c", 0600, os.Getuid(), os.Getgid(), 2, time.Now())

	var output_buffer bytes.Buffer
	args := []string{"-S"}
	ls_err := ls(&output_buffer, args, tw)

	output := clean_output_buffer(output_buffer)

	expected := "b c a"

	check_output(t, output, expected)
	check_error_nil(t, ls_err)
}

// Test running 'ls -rS' in a directory with files of the differing sizes.
func Test_rS_None_Files(t *testing.T) {
	setup_test_dir("rS_None_Files")

	_mkfile2("a", 0600, os.Getuid(), os.Getgid(), 0, time.Now())
	_mkfile2("b", 0600, os.Getuid(), os.Getgid(), 5, time.Now())
	_mkfile2("c", 0600, os.Getuid(), os.Getgid(), 2, time.Now())

	var output_buffer bytes.Buffer
	args := []string{"-rS"}
	ls_err := ls(&output_buffer, args, tw)

	output := clean_output_buffer(output_buffer)

	expected := "a c b"

	check_output(t, output, expected)
	check_error_nil(t, ls_err)
}

// Test running 'ls --dirs-first' in an empty directory
func Test_dirsfirst_None_None(t *testing.T) {
	setup_test_dir("dirsfirst_None_None")

	var output_buffer bytes.Buffer
	args := []string{"--dirs-first"}
	ls_err := ls(&output_buffer, args, tw)

	output := clean_output_buffer(output_buffer)

	expected := ""

	check_output(t, output, expected)
	check_error_nil(t, ls_err)
}

// Test running 'ls --dirs-first' in a directory with files
func Test_dirsfirst_None_Files(t *testing.T) {
	setup_test_dir("dirsfirst_None_Files")

	_mkfile("a")
	_mkfile("b")
	_mkfile("c")

	var output_buffer bytes.Buffer
	args := []string{"--dirs-first"}
	ls_err := ls(&output_buffer, args, tw)

	output := clean_output_buffer(output_buffer)

	expected := "a b c"

	check_output(t, output, expected)
	check_error_nil(t, ls_err)
}

// Test running 'ls --dirs-first' in a directory with files and directories
func Test_dirsfirst_None_FilesAndDirs(t *testing.T) {
	setup_test_dir("dirsfirst_None_FilesAndDirs")

	_mkfile("a")
	_mkdir("b")
	_mkfile("c")
	_mkdir("d")

	var output_buffer bytes.Buffer
	args := []string{"--dirs-first", "--nocolor"}
	ls_err := ls(&output_buffer, args, tw)

	output := clean_output_buffer(output_buffer)

	expected := "b d a c"

	check_output(t, output, expected)
	check_error_nil(t, ls_err)
}

// Test running 'ls --dirs-first *' in a directory with nested files and
// directories
func Test_dirsfirst_FilesAndDirs_FilesAndDirs(t *testing.T) {
	setup_test_dir("dirsfirst_FilesAndDirs_FilesAndDirs")

	_mkfile("a")
	_mkfile("b")
	_mkfile("c")

	_mkdir("dir1")
	_mkfile("dir1/e")
	_mkdir("dir1/f")
	_mkfile("dir1/g")

	var output_buffer bytes.Buffer
	args := []string{"--dirs-first", "--nocolor", "a", "b", "c", "dir1"}
	ls_err := ls(&output_buffer, args, tw)

	output := clean_output_buffer(output_buffer)

	expected := "dir1:\n" +
		"f e g\n\n" +
		"a b c"

	check_output(t, output, expected)
	check_error_nil(t, ls_err)
}

// -------------------------------COLOR TESTS-----------------------------------

// Test LSCOLORS directory color
func Test_LSCOLORS_Dir(t *testing.T) {
	setup_test_dir("LSCOLORS_Dir")

	_mkdir("test_dir")

	os.Setenv("LSCOLORS", default_LSCOLORS)

	var output_buffer bytes.Buffer
	args := []string{}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := fmt.Sprintf("\x1b[0;34mtest_dir\x1b[0m")

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test LS_COLORS directory color
func Test_LS_COLORS_Dir(t *testing.T) {
	setup_test_dir("LS_COLORS_Dir")

	_mkdir("test_dir")

	os.Setenv("LS_COLORS", default_LS_COLORS)

	var output_buffer bytes.Buffer
	args := []string{}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := fmt.Sprintf("\x1b[01;34mtest_dir\x1b[0m")

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test directory color with neither LSCOLORS or LS_COLORS defined--should
// default to the LSCOLORS specification
func Test_blankCOLOR_Dir(t *testing.T) {
	setup_test_dir("blankCOLOR_Dir")

	_mkdir("test_dir")

	var output_buffer bytes.Buffer
	args := []string{}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := fmt.Sprintf("\x1b[0;34mtest_dir\x1b[0m")

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test LSCOLORS world-writeable directory color
func Test_LSCOLORS_ow_Dir(t *testing.T) {
	setup_test_dir("LSCOLORS_ow_Dir")

	_mkdir2("test_dir", 0777, os.Getuid(), os.Getuid(), time.Now())

	os.Setenv("LSCOLORS", default_LSCOLORS)

	var output_buffer bytes.Buffer
	args := []string{}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := fmt.Sprintf("\x1b[0;30;43mtest_dir\x1b[0m")

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test LS_COLORS world-writeable directory color
func Test_LS_COLORS_ow_Dir(t *testing.T) {
	setup_test_dir("LS_COLORS_ow_Dir")

	_mkdir2("test_dir", 0777, os.Getuid(), os.Getuid(), time.Now())

	os.Setenv("LS_COLORS", default_LS_COLORS)

	var output_buffer bytes.Buffer
	args := []string{}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := fmt.Sprintf("\x1b[34;42mtest_dir\x1b[0m")

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test LSCOLORS world-writeable, sticky-bit directory color
func Test_LSCOLORS_ow_sticky_Dir(t *testing.T) {
	setup_test_dir("LSCOLORS_ow_sticky_Dir")

	_mkdir2("test_dir",
		0777|os.ModeSticky,
		os.Getuid(),
		os.Getuid(),
		time.Now())

	os.Setenv("LSCOLORS", default_LSCOLORS)

	var output_buffer bytes.Buffer
	args := []string{}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := fmt.Sprintf("\x1b[0;30;42mtest_dir\x1b[0m")

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test LS_COLORS world-writeable, sticky-bit directory color
func Test_LS_COLORS_ow_sticky_Dir(t *testing.T) {
	setup_test_dir("LS_COLORS_ow_sticky_Dir")

	_mkdir2("test_dir",
		0777|os.ModeSticky,
		os.Getuid(),
		os.Getuid(),
		time.Now())

	os.Setenv("LS_COLORS", default_LS_COLORS)

	var output_buffer bytes.Buffer
	args := []string{}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := fmt.Sprintf("\x1b[30;42mtest_dir\x1b[0m")

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test LSCOLORS symlink color
func Test_LSCOLORS_symlink(t *testing.T) {
	setup_test_dir("LSCOLORS_symlink")

	_mkfile("a")
	_mklink("a", "b")

	os.Setenv("LSCOLORS", default_LSCOLORS)

	var output_buffer bytes.Buffer
	args := []string{}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := fmt.Sprintf("a \x1b[0;35mb\x1b[0m")

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test LS_COLORS symlink color
func Test_LS_COLORS_symlink(t *testing.T) {
	setup_test_dir("LS_COLORS_symlink")

	_mkfile("a")
	_mklink("a", "b")

	os.Setenv("LS_COLORS", default_LS_COLORS)

	var output_buffer bytes.Buffer
	args := []string{}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := fmt.Sprintf("a \x1b[01;36mb\x1b[0m")

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test LSCOLORS executable color
func Test_LSCOLORS_executable(t *testing.T) {
	setup_test_dir("LSCOLORS_executable")

	_mkfile2("a", 0755, os.Getuid(), os.Getgid(), 0, time.Now())

	os.Setenv("LSCOLORS", default_LSCOLORS)

	var output_buffer bytes.Buffer
	args := []string{}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := fmt.Sprintf("\x1b[0;31ma\x1b[0m")

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test LS_COLORS executable color
func Test_LS_COLORS_executable(t *testing.T) {
	setup_test_dir("LS_COLORS_executable")

	_mkfile2("a", 0755, os.Getuid(), os.Getgid(), 0, time.Now())

	os.Setenv("LS_COLORS", default_LS_COLORS)

	var output_buffer bytes.Buffer
	args := []string{}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := fmt.Sprintf("\x1b[01;32ma\x1b[0m")

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// Test LS_COLORS with an orphan link
func Test_LS_COLORS_orphan_link(t *testing.T) {
	setup_test_dir("LS_COLORS_orphan_link")

	time_now := time.Now()

	_mkfile2("a", 0600, os.Getuid(), os.Getgid(), 8, time_now)
	_mklink("a", "b")
	_rm("a")

	os.Setenv("LS_COLORS", default_LS_COLORS)

	var output_buffer bytes.Buffer
	args := []string{"-l"}
	ls_err := ls(&output_buffer, args, tw)

	output := clean_output_buffer(output_buffer)
	// remove the permissions string from the output
	output_noperms := strings.Join(strings.Split(output, " ")[1:], " ")

	var owner string
	owner_lookup, err := user.LookupId(fmt.Sprintf("%d", os.Getuid()))
	if err != nil {
		owner = user_map[int(os.Getuid())]
	} else {
		owner = owner_lookup.Username
	}

	group := group_map[os.Getgid()]

	// link info
	link_info, err := os.Lstat("b")
	if err != nil {
		t.Fatalf("error checking the symlink")
	}

	expected := fmt.Sprintf("1 %s %s %d %s %02d %02d:%02d %sb%s -> %sa%s",
		owner,
		group,
		link_info.Size(),
		time_now.Month().String()[0:3],
		time_now.Day(),
		time_now.Hour(),
		time_now.Minute(),
		"\x1b[40;31;01m",
		"\x1b[0m",
		"\x1b[01;05;37;41m",
		"\x1b[0m")

	check_output(t, output_noperms, expected)
	check_error_nil(t, ls_err)
}

// Test LS_COLORS with a non-world-writeable directory with the sticky bit set
func Test_LS_COLORS_directory_sticky(t *testing.T) {
	setup_test_dir("LS_COLORS_directory_sticky")

	_mkdir2("test_dir",
		0755|os.ModeSticky,
		os.Getuid(),
		os.Getuid(),
		time.Now())

	os.Setenv("LS_COLORS", default_LS_COLORS)

	var output_buffer bytes.Buffer
	args := []string{}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := fmt.Sprintf("\x1b[37;44mtest_dir\x1b[0m")

	check_output(t, output, expected)
	check_error_nil(t, err)
}

// vim: tabstop=4 softtabstop=4 shiftwidth=4 noexpandtab tw=80
