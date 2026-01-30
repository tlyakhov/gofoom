package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	_ "tlyakhov/gofoom/components/behaviors"
	_ "tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/ecs"
	_ "tlyakhov/gofoom/scripting_symbols"
)

func main() {
	// Assumes we're running this from gofoom/tools/worlds
	os.Chdir("../../../gofoom-data")
	filepath.Walk("worlds/", func(path string, info fs.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		log.Printf("Processing %v", path)
		ecs.Initialize()
		if err = ecs.Load(path); err != nil {
			log.Printf("Error loading world %v", err)
			return nil
		}
		ecs.Save(path + ".updated")
		return nil
	})
	filepath.Walk("worlds/", func(path string, info fs.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".yaml.updated") {
			return nil
		}
		err = os.Rename(path, strings.TrimSuffix(path, ".updated"))
		if err != nil {
			log.Printf("Error renaming %v: %v", path, err)
		}
		return nil
	})
}
