package main

import "github.com/jt00721/meeting-notes-manager/config"

func main() {
	application := config.NewApp()

	application.Run()
}
