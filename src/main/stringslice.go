// Docker Image Authorization Plugin.
// Allows docker images to be fetched from a list of authorized registries only.
// AUTHOR: Chaitanya Prakash N <cpdevws@gmail.com>
package main

import "fmt"

// Supports multiple cmdline options to be parsed into a string array
// Usage: --<flag> <value> --<flag> <value> ...
type stringslice []string

func (str *stringslice) String() string {
	return fmt.Sprintf("%s", *str)
}

func (str *stringslice) Set(value string) error {
	*str = append(*str, value)
	return nil
}
