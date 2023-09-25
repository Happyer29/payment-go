package main

import (
	"log"
	"payment-go/internal/app"
)

func main() {
	application := app.GetApp()

	if err := application.Prepare(); err != nil {
		log.Fatal(err)
	}

	if err := application.Launch(); err != nil {
		log.Fatal(err)
	}
}
