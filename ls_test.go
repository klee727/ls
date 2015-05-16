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
	"testing"
	"time"
)

// the root directory where individual test subdirectories are stored.
var test_root string

// key-value map for converting GIDs to their string representations
var group_map map[int]string

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

	err = os.Chmod(path, mode)
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

	err = os.Chtimes(path, mod_epoch_s, mod_epoch_s)
	if err != nil {
		fmt.Printf("error: os.Chtimes(%s, %v, %v)\n", path, mod_epoch_s,
			mod_epoch_s)
		fmt.Printf("\t%v\n", err)
		os.Exit(1)
	}
}

/*
	if output_buffer.String() != expected {
		t.Logf("expected \"%s\", but got \"%s\"\n",
			expected,
			output_buffer.String())
		t.Fail()
	}
	*/
func check_output( t *testing.T, output, expected string ) {
	if output != expected {
		t.Logf("\nexpected:\n\"%s\"\n\nbut got:\n\"%s\"\n", expected, output)
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

	// read in all the information from /etc/groups
	group_map = make(map[int]string)

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

		gid, err := strconv.ParseInt(line_split[2], 10, 0)
		if err != nil {
			fmt.Printf("error:  couldn't convert %s to int\n", line_split[2])
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		group_name := line_split[0]
		group_map[int(gid)] = group_name
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
func Test_NoArgsNoFiles(t *testing.T) {
	_cd(test_root)

	dir := "NoArgsNoFiles"

	_mkdir(dir)
	_cd(dir)

	var output_buffer bytes.Buffer
	var args []string
	ls(&output_buffer, args)

	expected := ""

	check_output(t, output_buffer.String(), expected)
}

// Test running 'ls' in a directory with files
func Test_NoArgsFiles(t *testing.T) {
	_cd(test_root)

	dir := "NoArgsFiles"

	_mkdir(dir)
	_cd(dir)
	_mkfile("a")
	_mkfile("b")
	_mkfile("c")

	var output_buffer bytes.Buffer
	var args []string
	ls(&output_buffer, args)

	expected := "a b c"

	check_output(t, output_buffer.String(), expected)
}

// Test running 'ls' in a directory with .files
func Test_NoArgsDotFiles(t *testing.T) {
	_cd(test_root)

	dir := "NoArgsDotFiles"

	_mkdir(dir)
	_cd(dir)
	_mkfile(".a")
	_mkfile(".b")
	_mkfile(".c")

	var output_buffer bytes.Buffer
	var args []string
	ls(&output_buffer, args)

	expected := ""

	check_output(t, output_buffer.String(), expected)
}

func Test_NoArgsDotFilesInDir(t *testing.T) {
	_cd(test_root)

	dir := "NoArgsDotFilesInDir"
	dir2 := "dir"

	_mkdir(dir)
	_cd(dir)
	_mkdir(dir2)
	_mkfile(dir2 + "/.a")
	_mkfile(dir2 + "/.b")
	_mkfile(dir2 + "/.c")

	var output_buffer bytes.Buffer
	var args []string
	args = append(args, dir2)
	ls(&output_buffer, args)

	expected := ""

	check_output(t, output_buffer.String(), expected)
}

func Test_AllDotFiles(t *testing.T) {
	_cd(test_root)

	dir := "AllDotFiles"

	_mkdir(dir)
	_cd(dir)
	_mkfile(".a")
	_mkfile(".b")
	_mkfile(".c")

	var output_buffer bytes.Buffer
	var args []string
	args = append(args, "-a")
	ls(&output_buffer, args)

	expected := ". .. .a .b .c"

	check_output(t, output_buffer.String(), expected)
}

func Test_AllUpDir(t *testing.T) {
	_cd(test_root)

	dir := "AllUpDir"
	dir2 := "AllUpDir2"

	_mkdir(dir)
	_cd(dir)
	_mkfile(".a")
	_mkfile(".b")
	_mkfile(".c")
	_mkdir(dir2)
	_cd(dir2)

	var output_buffer bytes.Buffer
	args := []string{"-a", ".."}
	ls(&output_buffer, args)

	expected := ". .. .a .b .c AllUpDir2"

	check_output(t, output_buffer.String(), expected)
}

func Test_AllUpDir2(t *testing.T) {
	_cd(test_root)

	dir := "AllUpDir2"
	dir2 := "AllUpDir2_1"
	dir3 := "AllUpDir2_2"

	_mkdir(dir)
	_cd(dir)
	_mkfile(".a")
	_mkfile(".b")
	_mkfile(".c")
	_mkdir(dir2)
	_cd(dir2)
	_mkfile(".e")
	_mkfile(".f")
	_mkfile(".g")
	_mkdir(dir3)
	_cd(dir3)
	_mkfile(".h")
	_mkfile(".i")
	_mkfile(".j")
	_cd("..")

	var output_buffer bytes.Buffer
	args := []string{"-a", ".", "..", dir3}
	ls(&output_buffer, args)

	expected := ".:\n" +
		". .. .e .f .g AllUpDir2_2\n" +
		"\n" +
		"..:\n" +
		". .. .a .b .c AllUpDir2_1\n" +
		"\n" +
		"AllUpDir2_2:\n" +
		". .. .h .i .j"

	check_output(t, output_buffer.String(), expected)
}

func Test_OneFile(t *testing.T) {
	_cd(test_root)

	dir := "OneFile"

	_mkdir(dir)
	_cd(dir)
	_mkfile("a")

	var output_buffer bytes.Buffer
	args := []string{"a"}
	ls(&output_buffer, args)

	expected := "a"

	check_output(t, output_buffer.String(), expected)
}

func Test_LL_OneFile(t *testing.T) {
	_cd(test_root)

	dir := "LL_OneFile"

	_mkdir(dir)
	_cd(dir)

	time_now := time.Now()
	size := 13
	path := "a"
	_mkfile2(path, 0600, os.Getuid(), os.Getgid(), size, time_now)

	var output_buffer bytes.Buffer
	args := []string{"-l", "a"}
	ls(&output_buffer, args)

	output := clean_output_buffer(output_buffer)

	owner, err := user.LookupId(fmt.Sprintf("%d", os.Getuid()))
	if err != nil {
		fmt.Printf("error: user.LookupId(%d)\n", os.Getuid())
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	group := group_map[os.Getgid()]

	expected := fmt.Sprintf("-rw------- 1 %s %s %d %s %d %02d:%02d %s",
		owner.Username,
		group,
		size,
		time_now.Month().String()[0:3],
		time_now.Day(),
		time_now.Hour(),
		time_now.Minute(),
		path)

	check_output(t, output, expected)
}

func Test_option1(t *testing.T) {
	_cd(test_root)

	dir := "option1"

	_mkdir(dir)
	_cd(dir)

	_mkfile("a")
	_mkfile("b")
	_mkfile("c")
	_mkfile("d")
	_mkfile("e")

	var output_buffer bytes.Buffer
	args := []string{"-1"}
	ls(&output_buffer, args)

	expected := "a\nb\nc\nd\ne"

	check_output(t, output_buffer.String(), expected)
}

// vim: tabstop=4 softtabstop=4 shiftwidth=4 noexpandtab tw=80
