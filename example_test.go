package warmed_test

import (
	"fmt"

	"github.com/koenbollen/warmed"
)

func ExampleClient() {
	client := warmed.New("https://httpbin.org/") // compatible with http.Client
	defer client.Close()

	// time passes...

	// this request should have reused a connection:
	resp, _ := client.Get("https://httpbin.org/get")

	fmt.Println(resp.Status)
	// Output: 200 OK
}

func ExampleClient_Targets() {
	client := warmed.New() // without a target url
	defer client.Close()

	_, _ = client.Get("https://httpbin.org/get")

	targets := client.Targets()
	fmt.Println(targets)
	// Output: [https://httpbin.org/]
}
