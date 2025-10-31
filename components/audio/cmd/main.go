package main

import (
	"fmt"
	"time"
	"tlyakhov/gofoom/components/audio"
	"tlyakhov/gofoom/ecs"
)

func main() {
	ecs.Initialize()

	audio.Mixer.Initialize()
	if audio.Mixer.Error != nil {
		fmt.Printf("Failed to initialize audio engine: %v\n", audio.Mixer.Error)
		return
	}
	defer audio.Mixer.Close()

	/*eMusic := ecs.NewEntity()
	sndMusic := ecs.NewAttachedComponent(eMusic, audio.SoundCID).(*audio.Sound)
	sndMusic.Source = "data/sounds/r2beepboop.mp3"
	if err := sndMusic.Load(); err != nil {
		fmt.Printf("Failed to load music: %v\n", err)
	}

	if _, err := audio.PlaySound(sndMusic.Entity, 0); err != nil {
		fmt.Printf("Failed to play music: %v\n", err)
	}*/

	eIR := ecs.NewEntity()
	sndIR := ecs.NewAttachedComponent(eIR, audio.SoundCID).(*audio.Sound)
	sndIR.Source = "data/sounds/impulses/empty_corridor.wav"
	if err := sndIR.Load(); err != nil {
		fmt.Printf("Failed to load sound: %v\n", err)
	}

	eCollect := ecs.NewEntity()
	sndCollect := ecs.NewAttachedComponent(eCollect, audio.SoundCID).(*audio.Sound)
	sndCollect.Source = "data/sounds/collect.wav"
	if err := sndCollect.Load(); err != nil {
		fmt.Printf("Failed to load sound: %v\n", err)
	}

	event, err := audio.PlaySound(sndCollect.Entity, 0, "test", false)
	//event, err := mixer.Play(sndCollect, []audio.Effect{&audio.BitCrush{Bits: 8}}, 0.5)
	//event, err := mixer.Play(sndCollect, []audio.Effect{audio.NewDelay(12000, 0.5, 0.5)}, 0.5)
	//event.Effects = []audio.Effect{&audio.DistortionEffect{Mix: 1, Drive:
	//300000}}
	if err != nil {
		fmt.Printf("Failed to play sound: %v\n", err)
	}

	// Let the sounds play for a bit
	time.Sleep(45 * time.Second)

	// Stop the sound effect
	event.Stop()

	// Let the music play for another 5 seconds
	time.Sleep(35 * time.Second)
}
