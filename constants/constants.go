// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package constants

const (
	// Simulation constants
	TimeStepNS        = 1_000_000_000.0 / 128.0 // calibrate to 128 frames per second
	TimeStepMS        = TimeStepNS / 1_000_000.0
	TimeStepS         = TimeStepMS / 1_000.0
	MinMillisPerFrame = 1000.0 / 10.0 // Ensure we don't do a spiral of death at framerates < 10

	// Rendering constants
	MaxPortals              = 300 // avoid infinite portal traversal
	IntersectEpsilon        = 1e-8
	VelocityEpsilon         = 1e-15
	LightAttenuationEpsilon = 0.1
	MaxViewDistance         = 10000.0
	QuadtreeInitDim         = 256.0
	DebugLevel              = 1
	CollisionSteps          = 10
	MaxLightmapAge          = 3 // in frames
	LightmapRefreshDither   = 4 // in frames
	DebugLighting           = false
	MaxWeaponMarks          = 30

	// Rendering defaults
	FieldOfView         = 90
	RenderMultiThreaded = true
	RenderBlocks        = 32 // When multi-threaded, each block will have its own goroutine
	// Decrease this value for more detailed shadows. 2 looks nice, uses lots of
	// memory and is very slow.
	LightGrid = 4.0

	// Misc constants

	HumanQuantityEpsilon = 1e-10

	// Audio constants
	SourcesPerSound  = 8
	AudioUnitsFactor = 12

	// World constants
	UnitsPerMeter         = 32.0 // 20 units = 1m
	MetersPerUnit         = 1.0 / UnitsPerMeter
	Gravity               = 9.8 //0.005
	GravitySwim           = 0.1
	AirDensity            = 1.293
	SphereDragCoefficient = 0.47
	SwimDamping           = 1.5
	CollisionCheck        = 2.0
	LiquidChurnSpeed      = 2.0
	LiquidChurnSize       = 0.03

	// Player constants
	// Some of these are loosely based on Doom constants, see https://doomwiki.org/wiki/Player
	PlayerMass           = 80.0 // kg
	PlayerBoundingRadius = 10.0
	PlayerHeight         = 40.0
	PlayerCrouchHeight   = 16.0
	PlayerWalkForce      = 10.0 * PlayerMass * TimeStepMS // Newtons (we work backwards and aim for X meters/time step)
	PlayerTurnSpeed      = 180.0                          // Degrees per second
	PlayerJumpForce      = 60.0 * PlayerMass * TimeStepMS // Newtons
	PlayerSwimStrength   = 20.0 * PlayerMass * TimeStepMS
	PlayerHurtTime       = 30
	PlayerMountHeight    = 15.0
	PlayerMaxHealth      = 100

	// Game constants
	UserSettings    = "settings.json"
	DefaultFontPath = "data/fonts/vga-font-8x8.png"
	TestWorldPath   = "data/worlds/pursuer-test.yaml"

	// Pursuit
	PursuitNpcAvoidDistance    = 50.0
	PursuitWallAvoidDistance   = 50.0
	PursuitEnemyTargetDistance = 1000.0
	PursuitMaxBreadcrumbs      = 20
	PursuitBreadcrumbRateNs    = 250 * 1_000_000 // 250 ms
)
