/*
 * Copyright (c) Tim Lyakhovetskiy
 * SPDX-License-Identifier: MPL-2.0
 */

#include <stdio.h>
#include <stdlib.h>
#include <AL/al.h>
#include <AL/efx.h>

/* Filter object functions */
static LPALGENFILTERS pfnGenFilters;
static LPALDELETEFILTERS pfnDeleteFilters;
static LPALISFILTER pfnIsFilter;
static LPALFILTERI pfnFilteri;
static LPALFILTERIV pfnFilteriv;
static LPALFILTERF pfnFilterf;
static LPALFILTERFV pfnFilterfv;
static LPALGETFILTERI pfnGetFilteri;
static LPALGETFILTERIV pfnGetFilteriv;
static LPALGETFILTERF pfnGetFilterf;
static LPALGETFILTERFV pfnGetFilterfv;

/* Effect object functions */
static LPALGENEFFECTS pfnGenEffects;
static LPALDELETEEFFECTS pfnDeleteEffects;
static LPALISEFFECT pfnIsEffect;
static LPALEFFECTI pfnEffecti;
static LPALEFFECTIV pfnEffectiv;
static LPALEFFECTF pfnEffectf;
static LPALEFFECTFV pfnEffectfv;
static LPALGETEFFECTI pfnGetEffecti;
static LPALGETEFFECTIV pfnGetEffectiv;
static LPALGETEFFECTF pfnGetEffectf;
static LPALGETEFFECTFV pfnGetEffectfv;

/* Auxiliary Effect Slot object functions */
static LPALGENAUXILIARYEFFECTSLOTS pfnGenAuxiliaryEffectSlots;
static LPALDELETEAUXILIARYEFFECTSLOTS pfnDeleteAuxiliaryEffectSlots;
static LPALISAUXILIARYEFFECTSLOT pfnIsAuxiliaryEffectSlot;
static LPALAUXILIARYEFFECTSLOTI pfnAuxiliaryEffectSloti;
static LPALAUXILIARYEFFECTSLOTIV pfnAuxiliaryEffectSlotiv;
static LPALAUXILIARYEFFECTSLOTF pfnAuxiliaryEffectSlotf;
static LPALAUXILIARYEFFECTSLOTFV pfnAuxiliaryEffectSlotfv;
static LPALGETAUXILIARYEFFECTSLOTI pfnGetAuxiliaryEffectSloti;
static LPALGETAUXILIARYEFFECTSLOTIV pfnGetAuxiliaryEffectSlotiv;
static LPALGETAUXILIARYEFFECTSLOTF pfnGetAuxiliaryEffectSlotf;
static LPALGETAUXILIARYEFFECTSLOTFV pfnGetAuxiliaryEffectSlotfv;

void alLoadEAXProcs()
{
    // This code is copied from openal-soft:

    /* C doesn't allow casting between function and non-function pointer types, so
     * with C99 we need to use a union to reinterpret the pointer type. Pre-C99
     * still needs to use a normal cast and live with the warning (C++ is fine with
     * a regular reinterpret_cast).
     */
#if __STDC_VERSION__ >= 199901L
#define FUNCTION_CAST(T, ptr) (union {void *p; T f; }){ptr}.f
#else
#define FUNCTION_CAST(T, ptr) (T)(ptr)
#endif
    /* Define a macro to help load the function pointers. */
#define STRINGIFY(x) #x
#define LOAD_PROC(T, x) ((pfn##x) = FUNCTION_CAST(T, alGetProcAddress(STRINGIFY(al##x))))
    LOAD_PROC(LPALGENFILTERS, GenFilters);
    LOAD_PROC(LPALDELETEFILTERS, DeleteFilters);
    LOAD_PROC(LPALISFILTER, IsFilter);
    LOAD_PROC(LPALFILTERI, Filteri);
    LOAD_PROC(LPALFILTERIV, Filteriv);
    LOAD_PROC(LPALFILTERF, Filterf);
    LOAD_PROC(LPALFILTERFV, Filterfv);
    LOAD_PROC(LPALGETFILTERI, GetFilteri);
    LOAD_PROC(LPALGETFILTERIV, GetFilteriv);
    LOAD_PROC(LPALGETFILTERF, GetFilterf);
    LOAD_PROC(LPALGETFILTERFV, GetFilterfv);

    LOAD_PROC(LPALGENEFFECTS, GenEffects);
    LOAD_PROC(LPALDELETEEFFECTS, DeleteEffects);
    LOAD_PROC(LPALISEFFECT, IsEffect);
    LOAD_PROC(LPALEFFECTI, Effecti);
    LOAD_PROC(LPALEFFECTIV, Effectiv);
    LOAD_PROC(LPALEFFECTF, Effectf);
    LOAD_PROC(LPALEFFECTFV, Effectfv);
    LOAD_PROC(LPALGETEFFECTI, GetEffecti);
    LOAD_PROC(LPALGETEFFECTIV, GetEffectiv);
    LOAD_PROC(LPALGETEFFECTF, GetEffectf);
    LOAD_PROC(LPALGETEFFECTFV, GetEffectfv);

    LOAD_PROC(LPALGENAUXILIARYEFFECTSLOTS, GenAuxiliaryEffectSlots);
    LOAD_PROC(LPALDELETEAUXILIARYEFFECTSLOTS, DeleteAuxiliaryEffectSlots);
    LOAD_PROC(LPALISAUXILIARYEFFECTSLOT, IsAuxiliaryEffectSlot);
    LOAD_PROC(LPALAUXILIARYEFFECTSLOTI, AuxiliaryEffectSloti);
    LOAD_PROC(LPALAUXILIARYEFFECTSLOTIV, AuxiliaryEffectSlotiv);
    LOAD_PROC(LPALAUXILIARYEFFECTSLOTF, AuxiliaryEffectSlotf);
    LOAD_PROC(LPALAUXILIARYEFFECTSLOTFV, AuxiliaryEffectSlotfv);
    LOAD_PROC(LPALGETAUXILIARYEFFECTSLOTI, GetAuxiliaryEffectSloti);
    LOAD_PROC(LPALGETAUXILIARYEFFECTSLOTIV, GetAuxiliaryEffectSlotiv);
    LOAD_PROC(LPALGETAUXILIARYEFFECTSLOTF, GetAuxiliaryEffectSlotf);
    LOAD_PROC(LPALGETAUXILIARYEFFECTSLOTFV, GetAuxiliaryEffectSlotfv);
#undef LOAD_PROC
}

/*

Ideally we would get this extension function pointers from Go, but unfortunately
CGO can't call C functions via pointers, so we have to wrap all of this in C.

It could be good to use some kind of pre-processor/code-generation here, but
it wasn't obvious how to do this quickly.

There is also no error checking here, so devices that don't support these
functions will just crash.

*/

void alGenEffects(ALsizei n, ALuint *effects)
{
    pfnGenEffects(n, effects);
}
void alDeleteEffects(ALsizei n, const ALuint *effects)
{
    pfnDeleteEffects(n, effects);
}
ALboolean alIsEffect(ALuint effect)
{
    return pfnIsEffect(effect);
}
void alEffecti(ALuint effect, ALenum param, ALint iValue)
{
    pfnEffecti(effect, param, iValue);
}
void alEffectiv(ALuint effect, ALenum param, const ALint *piValues)
{
    pfnEffectiv(effect, param, piValues);
}
void alEffectf(ALuint effect, ALenum param, ALfloat flValue)
{
    pfnEffectf(effect, param, flValue);
}
void alEffectfv(ALuint effect, ALenum param, const ALfloat *pflValues)
{
    pfnEffectfv(effect, param, pflValues);
}
void alGetEffecti(ALuint effect, ALenum param, ALint *piValue)
{
    pfnGetEffecti(effect, param, piValue);
}
void alGetEffectiv(ALuint effect, ALenum param, ALint *piValues)
{
    pfnGetEffectiv(effect, param, piValues);
}
void alGetEffectf(ALuint effect, ALenum param, ALfloat *pflValue)
{
    pfnGetEffectf(effect, param, pflValue);
}
void alGetEffectfv(ALuint effect, ALenum param, ALfloat *pflValues)
{
    pfnGetEffectfv(effect, param, pflValues);
}

void alGenFilters(ALsizei n, ALuint *filters)
{
    pfnGenFilters(n, filters);
}
void alDeleteFilters(ALsizei n, const ALuint *filters)
{
    pfnDeleteFilters(n, filters);
}
ALboolean alIsFilter(ALuint filter)
{
    return pfnIsFilter(filter);
}
void alFilteri(ALuint filter, ALenum param, ALint iValue)
{
    pfnFilteri(filter, param, iValue);
}
void alFilteriv(ALuint filter, ALenum param, const ALint *piValues)
{
    pfnFilteriv(filter, param, piValues);
}
void alFilterf(ALuint filter, ALenum param, ALfloat flValue)
{
    pfnFilterf(filter, param, flValue);
}
void alFilterfv(ALuint filter, ALenum param, const ALfloat *pflValues)
{
    pfnFilterfv(filter, param, pflValues);
}
void alGetFilteri(ALuint filter, ALenum param, ALint *piValue)
{
    pfnGetFilteri(filter, param, piValue);
}
void alGetFilteriv(ALuint filter, ALenum param, ALint *piValues)
{
    pfnGetFilteriv(filter, param, piValues);
}
void alGetFilterf(ALuint filter, ALenum param, ALfloat *pflValue)
{
    pfnGetFilterf(filter, param, pflValue);
}
void alGetFilterfv(ALuint filter, ALenum param, ALfloat *pflValues)
{
    pfnGetFilterfv(filter, param, pflValues);
}

void alGenAuxiliaryEffectSlots(ALsizei n, ALuint *effectslots)
{
    pfnGenAuxiliaryEffectSlots(n, effectslots);
}
void alDeleteAuxiliaryEffectSlots(ALsizei n, const ALuint *effectslots)
{
    pfnDeleteAuxiliaryEffectSlots(n, effectslots);
}
ALboolean alIsAuxiliaryEffectSlot(ALuint effectslot)
{
    return pfnIsAuxiliaryEffectSlot(effectslot);
}
void alAuxiliaryEffectSloti(ALuint effectslot, ALenum param, ALint iValue)
{
    pfnAuxiliaryEffectSloti(effectslot, param, iValue);
}
void alAuxiliaryEffectSlotiv(ALuint effectslot, ALenum param, const ALint *piValues)
{
    pfnAuxiliaryEffectSlotiv(effectslot, param, piValues);
}
void alAuxiliaryEffectSlotf(ALuint effectslot, ALenum param, ALfloat flValue)
{
    pfnAuxiliaryEffectSlotf(effectslot, param, flValue);
}
void alAuxiliaryEffectSlotfv(ALuint effectslot, ALenum param, const ALfloat *pflValues)
{
    pfnAuxiliaryEffectSlotfv(effectslot, param, pflValues);
}
void alGetAuxiliaryEffectSloti(ALuint effectslot, ALenum param, ALint *piValue)
{
    pfnGetAuxiliaryEffectSloti(effectslot, param, piValue);
}
void alGetAuxiliaryEffectSlotiv(ALuint effectslot, ALenum param, ALint *piValues)
{
    pfnGetAuxiliaryEffectSlotiv(effectslot, param, piValues);
}
void alGetAuxiliaryEffectSlotf(ALuint effectslot, ALenum param, ALfloat *pflValue)
{
    pfnGetAuxiliaryEffectSlotf(effectslot, param, pflValue);
}
void alGetAuxiliaryEffectSlotfv(ALuint effectslot, ALenum param, ALfloat *pflValues)
{
    pfnGetAuxiliaryEffectSlotfv(effectslot, param, pflValues);
}
