package main

import (
	"fmt"
	app2 "github.com/IT-Nick/internal/app"
)

func main() {
	fmt.Println("app starting")

	//app, err := app2.NewApp(os.Getenv("CONFIG_PATH"))
	app, err := app2.NewApp("configs/values_examples.yaml")
	if err != nil {
		panic(err)
	}

	if err := app.ListenAndServe(); err != nil {
		panic(err)
	}
}
