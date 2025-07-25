package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	_ "tlyakhov/gofoom/archetypes"
	_ "tlyakhov/gofoom/components/behaviors"
	_ "tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/ecs"
)

func main() {
	os.Chdir("../..")
	filepath.Walk("./data/worlds/", func(path string, info fs.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		log.Printf("%v", path)
		//controllers.CreateTestWorld3(u)
		if err = ecs.Load(path); err != nil {
			log.Printf("Error loading world %v", err)
			return nil
		}
		//controllers.Respawn(u, true)
		//archetypes.CreateFont(u, "data/vga-font-8x8.png", "Default Font")
		ecs.Save(path)
		return nil
	})
}
