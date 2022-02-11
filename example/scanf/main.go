// Golang program to illustrate the usage of
// fmt.Scanf() function

// Including the main package
package main

// Importing fmt
import (
	"fmt"
	"log"
)

// Calling main
func main() {

	// Declaring some variables
	var name string
	var alphabet_count int
	var float_value float32
	var bool_value bool

	// Calling Scanf() function for
	// scanning and reading the input
	// texts given in standard input
	fmt.Print("sdfasdf:")
	_, err := fmt.Scanf("%s", &name)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Print("sdfasdf:")
	fmt.Scanf("%d", &alphabet_count)
	fmt.Print("sdfasdf:")
	fmt.Scanf("%g", &float_value)
	fmt.Print("sdfasdf:")
	fmt.Scanf("%t", &bool_value)

	// Printing the given texts
	fmt.Printf("%s %d %g %t", name,
		alphabet_count, float_value, bool_value)

}
