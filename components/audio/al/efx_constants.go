package al

const ExtEfxName = "ALC_EXT_EFX"

const EfxMajorVersion = 0x20001
const EfxMinorVersion = 0x20002
const MaxAuxiliarySends = 0x20003

/* Listener properties. */
const MetersPerUnit = 0x20004

/* Source properties. */
const DirectFilter = 0x20005
const AuxiliarySendFilter = 0x20006
const AirAbsorptionFactor = 0x20007
const RoomRolloffFactor = 0x20008
const ConeOuterGainhf = 0x20009
const DirectFilterGainhfAuto = 0x2000A
const AuxiliarySendFilterGainAuto = 0x2000B
const AuxiliarySendFilterGainhfAuto = 0x2000C

/* Effect properties. */

/* Reverb effect parameters */
const ReverbDensity = 0x0001
const ReverbDiffusion = 0x0002
const ReverbGain = 0x0003
const ReverbGainhf = 0x0004
const ReverbDecayTime = 0x0005
const ReverbDecayHfratio = 0x0006
const ReverbReflectionsGain = 0x0007
const ReverbReflectionsDelay = 0x0008
const ReverbLateReverbGain = 0x0009
const ReverbLateReverbDelay = 0x000A
const ReverbAirAbsorptionGainhf = 0x000B
const ReverbRoomRolloffFactor = 0x000C
const ReverbDecayHflimit = 0x000D

/* EAX Reverb effect parameters */
const EAXReverbDensity = 0x0001
const EAXReverbDiffusion = 0x0002
const EAXReverbGain = 0x0003
const EAXReverbGainhf = 0x0004
const EAXReverbGainlf = 0x0005
const EAXReverbDecayTime = 0x0006
const EAXReverbDecayHfratio = 0x0007
const EAXReverbDecayLfratio = 0x0008
const EAXReverbReflectionsGain = 0x0009
const EAXReverbReflectionsDelay = 0x000A
const EAXReverbReflectionsPan = 0x000B
const EAXReverbLateReverbGain = 0x000C
const EAXReverbLateReverbDelay = 0x000D
const EAXReverbLateReverbPan = 0x000E
const EAXReverbEchoTime = 0x000F
const EAXReverbEchoDepth = 0x0010
const EAXReverbModulationTime = 0x0011
const EAXReverbModulationDepth = 0x0012
const EAXReverbAirAbsorptionGainhf = 0x0013
const EAXReverbHfreference = 0x0014
const EAXReverbLfreference = 0x0015
const EAXReverbRoomRolloffFactor = 0x0016
const EAXReverbDecayHflimit = 0x0017

/* Chorus effect parameters */
const ChorusWaveform = 0x0001
const ChorusPhase = 0x0002
const ChorusRate = 0x0003
const ChorusDepth = 0x0004
const ChorusFeedback = 0x0005
const ChorusDelay = 0x0006

/* Distortion effect parameters */
const DistortionEdge = 0x0001
const DistortionGain = 0x0002
const DistortionLowpassCutoff = 0x0003
const DistortionEqcenter = 0x0004
const DistortionEqbandwidth = 0x0005

/* Echo effect parameters */
const EchoDelay = 0x0001
const EchoLrdelay = 0x0002
const EchoDamping = 0x0003
const EchoFeedback = 0x0004
const EchoSpread = 0x0005

/* Flanger effect parameters */
const FlangerWaveform = 0x0001
const FlangerPhase = 0x0002
const FlangerRate = 0x0003
const FlangerDepth = 0x0004
const FlangerFeedback = 0x0005
const FlangerDelay = 0x0006

/* Frequency shifter effect parameters */
const FrequencyShifterFrequency = 0x0001
const FrequencyShifterLeftDirection = 0x0002
const FrequencyShifterRightDirection = 0x0003

/* Vocal morpher effect parameters */
const VocalMorpherPhonemea = 0x0001
const VocalMorpherPhonemeaCoarseTuning = 0x0002
const VocalMorpherPhonemeb = 0x0003
const VocalMorpherPhonemebCoarseTuning = 0x0004
const VocalMorpherWaveform = 0x0005
const VocalMorpherRate = 0x0006

/* Pitchshifter effect parameters */
const PitchShifterCoarseTune = 0x0001
const PitchShifterFineTune = 0x0002

/* Ringmodulator effect parameters */
const RingModulatorFrequency = 0x0001
const RingModulatorHighpassCutoff = 0x0002
const RingModulatorWaveform = 0x0003

/* Autowah effect parameters */
const AutowahAttackTime = 0x0001
const AutowahReleaseTime = 0x0002
const AutowahResonance = 0x0003
const AutowahPeakGain = 0x0004

/* Compressor effect parameters */
const CompressorOnoff = 0x0001

/* Equalizer effect parameters */
const EqualizerLowGain = 0x0001
const EqualizerLowCutoff = 0x0002
const EqualizerMid1Gain = 0x0003
const EqualizerMid1Center = 0x0004
const EqualizerMid1Width = 0x0005
const EqualizerMid2Gain = 0x0006
const EqualizerMid2Center = 0x0007
const EqualizerMid2Width = 0x0008
const EqualizerHighGain = 0x0009
const EqualizerHighCutoff = 0x000A

/* Effect type */
const EffectFirstParameter = 0x0000
const EffectLastParameter = 0x8000
const EffectType = 0x8001

/* Effect types, used with the AL_EFFECT_TYPE property */
const EffectNull = 0x0000
const EffectReverb = 0x0001
const EffectChorus = 0x0002
const EffectDistortion = 0x0003
const EffectEcho = 0x0004
const EffectFlanger = 0x0005
const EffectFrequencyShifter = 0x0006
const EffectVocalMorpher = 0x0007
const EffectPitchShifter = 0x0008
const EffectRingModulator = 0x0009
const EffectAutowah = 0x000A
const EffectCompressor = 0x000B
const EffectEqualizer = 0x000C
const EffectEAXReverb = 0x8000

/* Auxiliary Effect Slot properties. */
const EffectSlotEffect = 0x0001
const EffectSlotGain = 0x0002
const EffectSlotAuxiliarySendAuto = 0x0003

/* NULL Auxiliary Slot ID to disable a source send. */
const EffectSlotNull = 0x0000

/* Filter properties. */

/* Lowpass filter parameters */
const LowpassGain = 0x0001
const LowpassGainhf = 0x0002

/* Highpass filter parameters */
const HighpassGain = 0x0001
const HighpassGainlf = 0x0002

/* Bandpass filter parameters */
const BandpassGain = 0x0001
const BandpassGainlf = 0x0002
const BandpassGainhf = 0x0003

/* Filter type */
const FilterFirstParameter = 0x0000
const FilterLastParameter = 0x8000
const FilterType = 0x8001

/* Filter types, used with the AlFilterType property */
const FilterNull = 0x0000
const FilterLowpass = 0x0001
const FilterHighpass = 0x0002
const FilterBandpass = 0x0003
