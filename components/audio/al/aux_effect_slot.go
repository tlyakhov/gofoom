// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package al

type AuxEffectSlot uint32

func (slot AuxEffectSlot) AuxiliaryEffectSloti(key Enum, v int32) {
	alAuxiliaryEffectSloti(slot, key, v)
}

// GenAuxEffectSlotss generates n new effects. The generated effects should be deleted
// once they are no longer in use.
func GenAuxEffectSlots(n int) []AuxEffectSlot {
	return alGenAuxiliaryEffectSlots(n)
}

func DeleteAuxiliaryEffectSlots(slots ...AuxEffectSlot) {
	alDeleteAuxiliaryEffectSlots(slots)
}
