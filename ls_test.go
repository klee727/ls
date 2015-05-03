package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

const (
	test_root = "/tmp/ls_test"
)

func _cd(path string) {
	err := os.Chdir(path)
	if err != nil {
		fmt.Printf("error: os.Chdir(%s)\n", path)
		fmt.Printf("\t%v\n", err)
		os.Exit(1)
	}
}

func _mkdir(path string) {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		fmt.Printf("error: os.MkdirAll(%s, 0755)\n", path)
		fmt.Printf("\t%v\n", err)
		os.Exit(1)
	}
}

func _rmdir(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("error: os.Removeall(%s)\n", path)
		fmt.Printf("\t%v\n", err)
		os.Exit(1)
	}
}

func _mkfile(path string) {
	_, err := os.Create(path)
	if err != nil {
		fmt.Printf("error: os.Create(%s)\n", path)
		fmt.Printf("\t%v\n", err)
		os.Exit(1)
	}
}

func TestMain(m *testing.M) {

	//
	// setup
	//

	// create the test root directory if it does not exist
	_, err := os.Stat(test_root)
	if err != nil && os.IsNotExist(err) {
		_mkdir(test_root)
	} else if err != nil {
		fmt.Printf("error: os.Stat(%s)\n", test_root)
		fmt.Printf("\t%v\n", err)
		os.Exit(1)
	} else {
		_rmdir(test_root)
		_mkdir(test_root)
	}

	_cd(test_root)

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

	if output_buffer.String() != expected {
		t.Logf("expected \"%s\", but got \"%s\"\n",
			expected,
			output_buffer.String())
		t.Fail()
	}
}

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

	if output_buffer.String() != expected {
		t.Logf("expected \"%s\", but got \"%s\"\n",
			expected,
			output_buffer.String())
		t.Fail()
	}
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

	if output_buffer.String() != expected {
		t.Logf("expected \"%s\", but got \"%s\"\n",
			expected,
			output_buffer.String())
		t.Fail()
	}
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

	if output_buffer.String() != expected {
		t.Logf("expected \"%s\", but got \"%s\"\n",
			expected,
			output_buffer.String())
		t.Fail()
	}
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

	if output_buffer.String() != expected {
		t.Logf("expected \"%s\", but got \"%s\"\n",
			expected,
			output_buffer.String())
		t.Fail()
	}
}

// markmini:ls mark$ ls -a .. .git
// ..:
// .          ..         hello      ls         stringutil web
//
// .git:
// .              HEAD           description    info           refs
// ..             branches       hooks          logs
// COMMIT_EDITMSG config         index          objects
// markmini:ls mark$ $GOPATH/bin/ls -a .. .git
// ..:
// hello ls stringutil web
//
// .git:
// COMMIT_EDITMSG HEAD branches config description hooks index info logs objects refs

// markmini:ls mark$ ls -a ls.go
// ls.go
// markmini:ls mark$ $GOPATH/bin/ls -a ls.go
// nothing to list?
