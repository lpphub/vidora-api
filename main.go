package main

import (
	"vidora-api/app/server"
)

func main() {
	app := server.New()
	app.Run()
}
