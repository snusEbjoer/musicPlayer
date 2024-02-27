// Sample Go code for user authorization

package main

import (
	"fmt"
	"log"
	"main/youtube"
)

func main() {
	youtube := youtube.C{}
	results, err := youtube.Search("квинка")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(results)
}
