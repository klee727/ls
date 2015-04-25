package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

//
// main
//
func main() {

	var output_buffer bytes.Buffer

	files, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Printf("woops\n")
	}

	for _, f := range files {
		output_buffer.WriteString(f.Name())
		output_buffer.WriteString(" ")
	}

	output_buffer.Truncate(output_buffer.Len() - 1) // remove the last space
	fmt.Printf("%s\n", output_buffer.String())
}
