package main

import (
	"fmt"

	. "github.com/ya-breeze/httpscripter/pkg"
)

func main() {
	body := JSON(map[string]interface{}{
		"name": "Jo\"hn",
		"o": map[string]interface{}{
			"a": 1,
			"b": 2,
			"c": true,
		},
	})
	POST("https://jsonplaceholder.typicode.com/todos", body,
		"Authorization:Bearer 1234",
		"username==john",
		"password==1234\"&",
	)

	if Succeed(Last.Response.StatusCode) {
		fmt.Println("TADA", Value("id").String())
	} else {
		fmt.Println("Oops", Last.Response.StatusCode)
	}
}
