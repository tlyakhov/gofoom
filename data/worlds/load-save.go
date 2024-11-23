package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/ecs"
)

func main() {
	os.Chdir("../..")
	filepath.Walk("./data/worlds/", func(path string, info fs.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".json") {
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
		archetypes.CreateFont(db, "data/RDE_8x8.png", "Default Font")
		db.Save(strings.Replace(path, ".json", ".yaml", -1))
		return nil
	})
}
