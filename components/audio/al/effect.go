// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package al

import "fmt"

type Effect uint32

// GenEffects generates n new effects. The generated effects should be deleted
// once they are no longer in use.
func GenEffects(n int) []Effect {
	return alGenEffects(n)
}

// DeleteEffects deletes the effects.
func DeleteEffects(effects ...Effect) {
	if len(effects) == 0 {
		return
	}
	alDeleteEffects(effects)
}

// Load the given initial reverb properties into the given OpenAL effect. See
// reverb_presets for more
func (effect Effect) Load(params *EfxReverbParams) error {
	/* Prepare the effect for EAX Reverb (standard reverb doesn't contain
	 * the needed panning vectors).
	 */
	alEffecti(effect, EffectType, EffectEAXReverb)

	if err := Error(); err != NoError {
		return fmt.Errorf("failed to set EAX reverb: %s", alGetString(err))
	}

	/* Load the reverb properties. */
	alEffectf(effect, EAXReverbDensity, params.Density)
	alEffectf(effect, EAXReverbDiffusion, params.Diffusion)
	alEffectf(effect, EAXReverbGain, params.Gain)
	alEffectf(effect, EAXReverbGainhf, params.GainHF)
	alEffectf(effect, EAXReverbGainlf, params.GainLF)
	alEffectf(effect, EAXReverbDecayTime, params.DecayTime)
	alEffectf(effect, EAXReverbDecayHfratio, params.DecayHFRatio)
	alEffectf(effect, EAXReverbDecayLfratio, params.DecayLFRatio)
	alEffectf(effect, EAXReverbReflectionsGain, params.ReflectionsGain)
	alEffectf(effect, EAXReverbReflectionsDelay, params.ReflectionsDelay)
	alEffectfv(effect, EAXReverbReflectionsPan, params.ReflectionsPan)
	alEffectf(effect, EAXReverbLateReverbGain, params.LateReverbGain)
	alEffectf(effect, EAXReverbLateReverbDelay, params.LateReverbDelay)
	alEffectfv(effect, EAXReverbLateReverbPan, params.LateReverbPan)
	alEffectf(effect, EAXReverbEchoTime, params.EchoTime)
	alEffectf(effect, EAXReverbEchoDepth, params.EchoDepth)
	alEffectf(effect, EAXReverbModulationTime, params.ModulationTime)
	alEffectf(effect, EAXReverbModulationDepth, params.ModulationDepth)
	alEffectf(effect, EAXReverbAirAbsorptionGainhf, params.AirAbsorptionGainHF)
	alEffectf(effect, EAXReverbHfreference, params.HFReference)
	alEffectf(effect, EAXReverbLfreference, params.LFReference)
	alEffectf(effect, EAXReverbRoomRolloffFactor, params.RoomRolloffFactor)
	alEffecti(effect, EAXReverbDecayHflimit, params.DecayHFLimit)

	if err := Error(); err != NoError {
		return fmt.Errorf("Error setting up reverb: %s", alGetString(err))
	}
	return nil
}
