package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	_ "tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/components/audio"
	_ "tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/materials"
	_ "tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/ecs"
	_ "tlyakhov/gofoom/scripting_symbols"
)

const baseDataPath = "../data/"
const baseDataPathInGame = "data/"

var outputFile = filepath.Join(baseDataPath, "worlds/common-assets.yaml")
var existingAssets = make(map[string]ecs.Entity)

// generateName processes the input string to replace non-alphanumeric characters
// with a single space, and removes leading/trailing whitespace.
func generateName(input string) string {
	reNonAlphanum := regexp.MustCompile(`[^a-zA-Z0-9_ /]+`)
	substituted := reNonAlphanum.ReplaceAllString(input, " ")
	reMultipleSpaces := regexp.MustCompile(`\s{2,}`)
	singleSpace := reMultipleSpaces.ReplaceAllString(substituted, " ")
	result := strings.TrimSpace(singleSpace)
	split := strings.Split(result, "/")
	if len(split) > 1 {
		return split[len(split)-1] + " (" + strings.Join(split[:len(split)-1], " ") + ")"
	}

	return result
}

func hashAsset(cid ecs.ComponentID, src string) string {
	cidString := strconv.FormatUint(uint64(cid), 10)
	return cidString + "|" + src
}

func cacheExistingAssets(cid ecs.ComponentID) {
	// images
	arena := ecs.ArenaByID(cid)
	for i := range arena.Cap() {
		asset := arena.Component(i)
		if asset == nil {
			continue
		}

		switch c := asset.(type) {
		case *materials.Image:
			if c.Source != "" {
				existingAssets[hashAsset(cid, c.Source)] = c.Entity
			}
		case *audio.Sound:
			if c.Source != "" {
				existingAssets[hashAsset(cid, c.Source)] = c.Entity
			}
		}
	}
}

func assetEntity(cid ecs.ComponentID, src string) ecs.Entity {
	hash := hashAsset(cid, src)
	if e, ok := existingAssets[hash]; ok {
		ecs.Delete(e)
		ecs.CreateEntity(e)
		return e
	}

	return ecs.NewEntity()
}

func main() {
	ecs.Initialize()

	if _, err := os.Stat(outputFile); err == nil {
		err := ecs.Load(outputFile)
		if err != nil {
			log.Printf("Error loading file: %v", err)
			return
		}
		cacheExistingAssets(materials.ImageCID)
		cacheExistingAssets(audio.SoundCID)
	}

	baseDataPathAbs, _ := filepath.Abs(baseDataPath)
	split := strings.Split(baseDataPathAbs, string(filepath.Separator))
	toTrim := len(split)
	filepath.WalkDir(baseDataPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("Error walking path: %v", err)
			return err
		}
		if d.IsDir() {
			return nil
		}
		dir, file := filepath.Split(path)
		ext := filepath.Ext(file)
		abs, _ := filepath.Abs(path)

		gamePath := filepath.Join(strings.Split(abs, string(filepath.Separator))[toTrim:]...)
		gamePath = filepath.Join(baseDataPathInGame, gamePath)
		log.Printf("%v", gamePath)
		// Generate name
		name, _ := filepath.Rel(baseDataPath, path)
		name = strings.TrimSuffix(name, ext)
		name = generateName(name)

		var e ecs.Entity
		switch ext {
		case ".jpg", ".jpeg", ".png":
			// Image asset
			log.Printf("%v is an image. Name: %v", path, name)
			e = assetEntity(materials.ImageCID, gamePath)
			img := ecs.NewAttachedComponent(e, materials.ImageCID).(*materials.Image)
			img.Source = gamePath

			if strings.Contains(dir, "fonts") {
				img.GenerateMipMaps = false
				img.Filter = false
				sprite := ecs.NewAttachedComponent(e, materials.SpriteSheetCID).(*materials.SpriteSheet)
				sprite.Rows = 16
				sprite.Cols = 16
				sprite.Material = e
				sprite.Angles = 0
			}
		case ".wav", ".mp3", ".ogg":
			if strings.Contains(path, "music") {
				// Don't preload music
				return nil
			}
			// Sound asset
			log.Printf("%v is a sound. Name: %v", path, name)
			e = assetEntity(audio.SoundCID, gamePath)
			snd := ecs.NewAttachedComponent(e, audio.SoundCID).(*audio.Sound)
			snd.Source = gamePath
		default:
			// Unknown
			return nil
		}
		named := ecs.NewAttachedComponent(e, ecs.NamedCID).(*ecs.Named)
		named.Name = name

		return nil
	})
	ecs.Save(outputFile)
}
