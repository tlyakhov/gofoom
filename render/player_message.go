package render

import (
	"bytes"
	"fmt"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/ecs"
)

type PlayerMessageParams struct {
	TargetableEntity ecs.Entity
	PlayerTargetable *behaviors.PlayerTargetable
	Player           *character.Player
	Carrier          *inventory.Carrier
}

func ApplyPlayerMessage(pt *behaviors.PlayerTargetable, params *PlayerMessageParams) string {
	if pt.MessageTemplate == nil {
		return pt.Message
	}

	var buf bytes.Buffer
	err := pt.MessageTemplate.Execute(&buf, params)
	if err != nil {
		return fmt.Sprintf("Error in message template %v: %v", pt.Message, err)
	}
	return buf.String()
}
