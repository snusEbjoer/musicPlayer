// Sample Go code for user authorization

package main

import (
	"main/youtube"
)

func main() {
	// client := auth.C{}
	// ctx := client.ClientContext()
	youtube := youtube.C{}
	youtube.Search("квинка")
	// client.FetchToken()
	// context, err := json.Marshal(ctx.Context)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(string(context))
}
