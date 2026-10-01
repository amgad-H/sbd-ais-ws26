package main

import "fmt"

type Greeting struct {
	Words string
}

func (g Greeting) Shout() string {
	return g.Words
}

func main() {
	g := Greeting{"Hello WOld"}
	fmt.Println(g.Shout())
}
