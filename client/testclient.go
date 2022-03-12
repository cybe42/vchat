package main

import "fmt"

func main() {
	user := Client{Name: "cybebot", IP: "ws://localhost:80/echo"}
	err := user.Connect()
	if err != nil {
		panic(err)
	}

	fmt.Println("channel: ", user.GetChannel())
	erree := user.Send("Hello World!", user.GetChannel())
	if erree != nil {
		panic(erree)
	}
	user.Listen(func(msg Msg, err error) {
		if err != nil {
			panic(err)
		}
		fmt.Println(msg.Message)
	})
}
