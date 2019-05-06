# warmed

[![Go Report Card](https://goreportcard.com/badge/github.com/koenbollen/warmed)](https://goreportcard.com/report/github.com/koenbollen/warmed)

This package contains a wrapper around the normal [http.Client] from Go that
proactively keeps a connection open to target hosts.

This was an experiment to see if such a stratagy would result in a performance
gain for resources/codepaths that are not hit often enough for normal 
keep-alive features to kick in. 

## Scenario

Given a service, endpoint or codepath that is rarely used, say once a week. If 
this endpoint has upstream https services normal keep-alive behavior will most 
certainly have timed out. Using the `warmed.Client` instead will make sure the 
connection is always ready. 

## Usage

```bash
go get -u github.com/koenbollen/warmed
```

```golang
package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/koenbollen/warmed"
)

func main() {
	client := warmed.New("https://www.googleapis.com")
	defer client.Close() // to cleanup go routine

	time.Sleep(10 * time.Minute)

	resp, _ := client.Get("https://www.googleapis.com/books/v1/volumes?q=Golang")
	data, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close() // make sure to always close the body

	fmt.Println(resp.StatusCode, len(data))
}
```




[http.Client]: https://golang.org/pkg/net/http/
