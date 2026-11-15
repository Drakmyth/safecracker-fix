package main

import (
	"log"

	safecrackerfix "github.com/Drakmyth/safecracker-fix"
	"github.com/Drakmyth/safecracker-fix/internal/gui"
)

func main() {
	if err := gui.Run(safecrackerfix.Version()); err != nil {
		log.Fatal(err)
	}
}
