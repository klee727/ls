package main

// option to list directories first

//if fi.Mode() & os.ModeSymlink == os.ModeSymlink {

import (
	"bytes"
	"fmt"
	"io/ioutil"
    "os"
    "regexp"
)

//
// main
//
func main() {

    args_options := make([]string, 0)
    args_files   := make([]string, 0)
    list_dirs    := make([]os.FileInfo, 0)
    list_files   := make([]os.FileInfo, 0)

    //
    // parse arguments
    //
    for _, a := range os.Args[1:] {
        option, err := regexp.MatchString( "^-", a )

        if err != nil {
            fmt.Printf( "error: %v\n", err )
            os.Exit( 1 )
        } else if option {
            // add to the options list
            args_options = append( args_options, a )
        } else {
            // add to the files/directories list
            args_files = append( args_files, a )
        }
    }

	var output_buffer bytes.Buffer

    // if no files are specified, list the current directory
    if len( args_files ) == 0 {
        this_dir, _ := os.Lstat( "." )
        list_dirs = append( list_dirs, this_dir )
    }

    //
    // separate the files from the directories
    //
    for _, f := range args_files {
        info, err := os.Lstat( f )

        if err != nil {
            fmt.Printf( "error: %v\n", err )
            os.Exit( 1 )
        }

        if info.IsDir() {
            list_dirs = append( list_dirs, info )
        } else {
            list_files = append( list_files, info )
        }
    }

    num_files := len( list_files )
    num_dirs  := len( list_dirs )

    //
    // list the files first
    //
    if num_files > 0 {
        for _, f := range list_files {
            output_buffer.WriteString( f.Name() )
            output_buffer.WriteString( " " )
        }
    }

    //
    // then list the directories
    //
    if ( num_files > 0 && num_dirs > 0 ) || ( num_dirs > 1 ) {
        if num_files > 0 {
            output_buffer.WriteString( "\n" )
        }

        for _, d := range list_dirs {
            output_buffer.WriteString( d.Name() + ":\n" )

            files_in_dir, err := ioutil.ReadDir( d.Name() )
            if err != nil {
                fmt.Printf( "error: %v\n", err )
                os.Exit( 1 )
            }

            for _, _f := range files_in_dir {
                output_buffer.WriteString( _f.Name() + " " )
            }
            output_buffer.Truncate( output_buffer.Len() - 1)
            output_buffer.WriteString( "\n\n" )
        }
        output_buffer.Truncate( output_buffer.Len() - 2 )
    } else if num_dirs == 1 {
        for _, d := range list_dirs {

            files_in_dir, err := ioutil.ReadDir( d.Name() )
            if err != nil {
                fmt.Printf( "error: %v\n", err )
                os.Exit( 1 )
            }

            for _, _f := range files_in_dir {
                output_buffer.WriteString( _f.Name() + " " )
            }
            output_buffer.Truncate( output_buffer.Len() - 1)
        }
    } else {
        fmt.Printf( "nothing to list?\n" )
        os.Exit( 1 )
    }

	fmt.Printf("%s\n", output_buffer.String())
}
