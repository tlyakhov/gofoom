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
		db := ecs.NewECS()
		//controllers.CreateTestWorld3(db)
		if err = db.Load(path); err != nil {
			log.Printf("Error loading world %v", err)
			return nil
		}
		//controllers.Respawn(db, true)
		//archetypes.CreateFont(db, "data/vga-font-8x8.png", "Default Font")
		db.Save(path)
		return nil
	})
}
