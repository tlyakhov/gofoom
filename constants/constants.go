package constants

const (
	// Rendering constants
	IntersectEpsilon        = 1e-10
	LightGrid               = 4
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
	Gravity          = 0.1
	GravitySwim      = 0.01
	SwimDamping      = 2.0
	CollisionCheck   = 2.0
	LiquidChurnSpeed = 2.0
	LiquidChurnSize  = 0.03
	DoorSpeed        = 3.0

	// Player constants
	PlayerBoundingRadius = 10.0
	PlayerHeight         = 32.0
	PlayerCrouchHeight   = 16.0
	PlayerSpeed          = 1
	PlayerTurnSpeed      = 4.0
	PlayerJumpStrength   = 1.5
	PlayerSwimStrength   = 0.6
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
