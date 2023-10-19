package constants

const (
	// Simulation constants
	TimeStep          = 1000.0 / 120.0 // calibrate to 120 frames per second
	MinMillisPerFrame = 1000.0 / 15.0  // Ensure we don't do a spiral of death at framerates < 15

	// Rendering constants
	MaxPortals              = 30 // avoid infinite portal traversal
	IntersectEpsilon        = 1e-10
	LightGrid               = 6.0
	LightSafety             = 2
	VelocityEpsilon         = 1e-15
	LightAttenuationEpsilon = 0.001
	MaxViewDistance         = 1000000.0
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
	Gravity          = 9.8 * 2 / (TimeStep * TimeStep) //0.005
	GravitySwim      = 0.1 * 2 / (TimeStep * TimeStep)
	SwimDamping      = 2.0
	CollisionCheck   = 2.0
	LiquidChurnSpeed = 2.0
	LiquidChurnSize  = 0.03
	DoorSpeed        = 0.1

	// Player constants
	PlayerBoundingRadius = 10.0
	PlayerHeight         = 32.0
	PlayerCrouchHeight   = 16.0
	PlayerSpeed          = 0.015
	PlayerTurnSpeed      = 0.17
	PlayerJumpStrength   = 0.02
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
