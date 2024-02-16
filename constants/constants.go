package constants

const (
	// Simulation constants
	TimeStep          = 1000.0 / 128.0 // calibrate to 128 frames per second
	TimeStepS         = TimeStep / 1000.0
	MinMillisPerFrame = 1000.0 / 10.0 // Ensure we don't do a spiral of death at framerates < 10

	// Rendering constants
	MaxPortals              = 100 // avoid infinite portal traversal
	IntersectEpsilon        = 1e-10
	LightGrid               = 5.0
	LightSafety             = 2
	VelocityEpsilon         = 1e-15
	LightAttenuationEpsilon = 0.001
	MaxViewDistance         = 10000.0
	FieldOfView             = 90
	DebugLevel              = 1
	CollisionSteps          = 10
	MaxLightmapAge          = 4 // in frames
	LightmapRefreshDither   = 8 // in frames
	RenderMultiThreaded     = true
	DebugLighting           = false

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
	SwimDamping           = 2.0
	CollisionCheck        = 2.0
	LiquidChurnSpeed      = 2.0
	LiquidChurnSize       = 0.03
	DoorSpeed             = 0.04

	// Player constants
	// Some of these are loosely based on Doom constants, see https://doomwiki.org/wiki/Player
	PlayerMass           = 80.0 // kg
	PlayerBoundingRadius = 10.0
	PlayerHeight         = 40.0
	PlayerCrouchHeight   = 16.0
	PlayerWalkForce      = 10.0 * PlayerMass * TimeStep // Newtons (we work backwards and aim for X meters/time step)
	PlayerTurnSpeed      = 180.0                        // Degrees per second
	PlayerJumpForce      = 30.0 * PlayerMass * TimeStep // Newtons
	PlayerSwimStrength   = 0.005
	PlayerHurtTime       = 30
	PlayerMountHeight    = 15.0
	PlayerMaxHealth      = 100

	// Game constants
	MaxGameTextTime          = 15 * 1000
	MaxGameText              = 25
	GameTextFadeTime         = 1000
	InteractionDistance      = 70
	WanderSectorProbability  = 0.25
	InventoryGatherDistance  = 30
	InventoryGatherTextStyle = "#666"
	InfoBarSrc               = "/data/game/infobar.png"
	AvatarSrc                = "/data/game/avatar.png"
	FirstMap                 = "testMap"
)
