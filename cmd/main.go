package main

import (
	"fmt"

	"kubeguide/internal/app"
)

func main() {
	kubeguideApp := app.New()

	if err := kubeguideApp.Initialize(); err != nil {
		fmt.Printf("Error initializing application: %v\n", err)
		return
	}

	if err := kubeguideApp.Run(); err != nil {
		panic(err)
	}
}
