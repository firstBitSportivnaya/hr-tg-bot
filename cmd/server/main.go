package main

import (
	"fmt"
	app2 "github.com/IT-Nick/internal/app"
	"os"
)

func main() {
	fmt.Println("app starting")

	app, err := app2.NewApp(os.Getenv("CONFIG_PATH"))
	if err != nil {
		panic(err)
	}

	if err := app.ListenAndServe(); err != nil {
		panic(err)
	}
}
