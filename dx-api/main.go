package main

import (
	"dx-api/bootstrap"
)

func main() {
	app := bootstrap.Boot()

	app.Start()
}
