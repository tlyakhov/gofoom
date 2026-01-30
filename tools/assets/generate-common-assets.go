package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"tlyakhov/gofoom/components/audio"
	_ "tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/materials"
	_ "tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/ecs"
	_ "tlyakhov/gofoom/scripting_symbols"
)

// Assumes we're running this from gofoom/tools/assets
const baseDataPath = "../../../gofoom-data/"
const baseDataPathInGame = ""

var outputFile = filepath.Join(baseDataPath, "worlds/common-assets.yaml")
var existingAssets = make(map[string]ecs.Entity)
var updatedAssets = make(map[ecs.Entity]string)

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

func hashFile(src string) []byte {
	f, err := os.Open(src)
	if err != nil {
		return nil
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return h.Sum(nil)
}

func hashAssetFilename(cid ecs.ComponentID, src string) string {
	cidString := strconv.FormatUint(uint64(cid), 10)
	return cidString + "|" + src
}

func hashAssetContents(cid ecs.ComponentID, hash []byte) string {
	if len(hash) == 0 {
		return ""
	}
	cidString := strconv.FormatUint(uint64(cid), 10)
	return cidString + "|" + fmt.Sprintf("%x", hash)
}

func cacheAsset(cid ecs.ComponentID, src string, e ecs.Entity) {
	if src == "" {
		return
	}
	hashContents := hashAssetContents(cid, hashFile(src))
	if hashContents != "" {
		existingAssets[hashContents] = e
	}
	existingAssets[hashAssetFilename(cid, src)] = e
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
			cacheAsset(cid, c.Source, c.Entity)
		case *audio.Sound:
			cacheAsset(cid, c.Source, c.Entity)
		}
	}
}

func assetEntity(cid ecs.ComponentID, src string) ecs.Entity {
	hash := hashAssetFilename(cid, src)
	if e, ok := existingAssets[hash]; ok {
		ecs.Delete(e)
		ecs.CreateEntity(e)
		updatedAssets[e] = src
		return e
	}
	if len(src) > 0 {
		hash = hashAssetContents(cid, hashFile(src))
		if e, ok := existingAssets[hash]; hash != "" && ok {
			ecs.Delete(e)
			ecs.CreateEntity(e)
			updatedAssets[e] = src
			return e
		}
	}

	e := ecs.NewEntity()
	updatedAssets[e] = src
	return e
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
			} else {
				ecs.NewAttachedComponent(e, materials.LitCID)
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
	ecs.Entities.Range(func(index uint32) {
		if index == 0 {
			return
		}
		entity := ecs.Entity(index)

		if _, ok := updatedAssets[entity]; !ok {
			log.Printf("%v no longer exists, removing.", entity.String())
			ecs.Delete(entity)
		}
	})
	ecs.Save(outputFile)
}
