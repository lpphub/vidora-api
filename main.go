package main

import "vidora-api/server"

func main() {
	app := server.New()
	app.Run()
}