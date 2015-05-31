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

// recursively remove the directory at the given path
func _rmdir(path string) {
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
	_rmdir(test_root)

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

// Test running 'ls' in a directory containing a directory, and ensure that the
// directory's name is in color
func Test_None_None_Dir(t *testing.T) {
	setup_test_dir("None_None_Dir")

	_mkdir("test_dir")

	var output_buffer bytes.Buffer
	args := []string{}
	err := ls(&output_buffer, args, tw)
	output := clean_output_buffer(output_buffer)

	expected := fmt.Sprintf("%stest_dir%s", color_blue, color_none)

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
	expected_err := "cannot access a: No such file or directory"

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

	expected := fmt.Sprintf("-rw------- 1 %s %s %d %s %d %02d:%02d %s",
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

	_mkdir2("dir0", 0755, os.Getuid(), os.Getgid(), time_now)
	_mkfile2("dir0/minus_one", 0600, os.Getuid(), os.Getgid(), 0,
		time_now.Add(-1*time.Second))
	_mkfile2("dir0/zero", 0600, os.Getuid(), os.Getgid(), 0, time_now)

	_mkdir2("dir2", 0755, os.Getuid(), os.Getgid(), time_now.Add(-2*time.Second))
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

	_mkdir2("dir1", 0755, os.Getuid(), os.Getgid(), time_now.Add(-5*time.Second))

	var output_buffer bytes.Buffer
	args := []string{"-t", "a", "b", "dir0", "dir1", "dir2"}
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

// vim: tabstop=4 softtabstop=4 shiftwidth=4 noexpandtab tw=80
