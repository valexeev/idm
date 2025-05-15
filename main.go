package main

import (
	"fmt"

	"github.com/google/uuid"
)

func main() {
	fmt.Println("Hello, Go.")

	id := uuid.New()
	fmt.Println("Generated UUID:", id.String())
}
