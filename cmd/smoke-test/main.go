package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Drakmyth/safecracker-fix/internal/installer"
	"github.com/Drakmyth/safecracker-fix/internal/safecracker"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("usage: smoke-test <inspect|apply|reverse|inspect-shim|apply-shim|restore-shim> <game-directory>")
	}

	action := os.Args[1]
	directory := os.Args[2]

	installation, err := installer.Inspect(directory)
	if err != nil {
		return err
	}

	switch action {
	case "inspect":
		// No modifications.
	case "apply":
		installation, err = installation.Apply(safecracker.PATCH_FPU_PRESERVE)
	case "reverse":
		installation, err = installation.Reverse(safecracker.PATCH_FPU_PRESERVE)
	case "inspect-shim":
		inspection, err := installation.InspectGraphicsShim()
		if err != nil {
			log.Fatal(err)
		}
		printGraphicsShimInspection(inspection)
	case "install-shim":
		inspection, err := installation.InstallGraphicsShim()
		if err != nil {
			log.Fatal(err)
		}
		printGraphicsShimInspection(inspection)
	case "remove-shim":
		inspection, err := installation.RemoveGraphicsShim()
		if err != nil {
			log.Fatal(err)
		}
		printGraphicsShimInspection(inspection)
	default:
		return fmt.Errorf("unknown action: %q", action)
	}

	if err != nil {
		return err
	}

	fmt.Println("Directory:", installation.Directory)
	fmt.Println("Executable:", installation.ExecutablePath)
	fmt.Println("Target:", installation.Target.Name)
	fmt.Println("Backup:", installation.BackupPath())

	for _, patch := range installation.Target.Patches {
		state := installation.Inspection.Patches[patch.ID]
		fmt.Printf("%s: %s\n", patch.Name, state)
	}

	return nil
}

func printGraphicsShimInspection(inspection *installer.GraphicsShimInspection) {
	fmt.Println("Graphics shim:", inspection.State)
	fmt.Println("DLL exists:", inspection.DLLExists)
	fmt.Println("DLL matches:", inspection.DLLMatches)
	fmt.Println("Config exists:", inspection.ConfigExists)
	fmt.Println("Config managed:", inspection.ConfigManaged)
	fmt.Println("Config matches:", inspection.ConfigMatches)
}
