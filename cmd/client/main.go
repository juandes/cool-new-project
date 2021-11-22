package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	token = flag.String("token", "", "Access token")
)

func main() {
	flag.Parse()
	resp, err := http.Get(fmt.Sprintf("http://localhost:8080/averages?token=%s", *token))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}
