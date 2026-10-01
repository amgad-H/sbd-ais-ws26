package main

import "fmt"

type Greeting struct {
	Words string
}

func (g *Greeting) Shout() {
	fmt.Println(g.Words)
}

func main() {
	g := Greeting{"Hello WOld"}
	g.Shout()
}
