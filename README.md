# gofoom

gofoom is a very WIP golang 2.5D sector/portal-based software raycasting game
engine. It continues on from my older project
https://github.com/tlyakhov/jsfoom.

The included game assets are programmer art & creative commons things I'm using
for testing. As the engine gets closer, I hope to make something actually fun
with it :)

## Motivation

Constraints are key! This project has the following goals:
* No GPU involvement in rendering (preferably anything else either).
* Pure dynamic raycasting for graphics.
    * No pre-processing for worlds (e.g. BSPs)
    * Rays correspond to actual rendered vertical columns - no intermediate
      translation to quads, triangles, or spans.
    * The benefit of this is that every game object or property can be dynamic -
      including the world geometry!
* High quality, efficient editing tools.
* High degree of flexibility and expandability:
    * Game code separate from engine code.
    * User scripting
    * Corrolary: lots of knobs to tweak. Examples include per-sector
      gravity/friction, entities can have different collision response, etc...
* Scalability. While the number of sectors/segments/entities on screen at once
  may have performance limitations, overall level size should be unconstrained.
* Cross-platform compatibility (Linux, Mac tested, not sure about Windows)

This project takes inspiration from Ken Silverman's BUILD engine used for Duke
Nukem 3D for the portal concept and Doom of course, but as mentioned above
unlike BUILD and Doom it aims to be a fully dynamic raycaster, avoiding
pre-processing steps or multi-pass rendering. This simplicity gives a lot of
benefits as far as dynamic sectors, lighting, etc...

## Features

* Multi-threaded rendering using goroutines.
* Sectors with non-orthogonal walls of variable height.
* Texture mapped floors, ceilings, and walls.
  * Layered texture shaders
  * Transparent walls and "stained glass"
* Sloped floors and ceilings.
* Objects represented as sprites with multiple angles.
* No limits to sector/entity counts
* Fully dynamic lighting with dynamic shadow maps.
* Bilinear filtering & mipmapping for images and shadow maps.
* Separation of game/engine code.
* Entity/Component/System architecture
* Behavioral system for entities.
* Various effect sectors (doors, underwater sectors)
* Physics and collision detection for player and objects.
* World editor:
    * Realtime 3D view, allowing live manipulation of the world by clicking.
    * Edit any sector/segment/entity property.
    * Slice sectors/split segments.
    * Automatic portal generation.
    * Undo/redo history.

### Soon:

* More interactivity: inventory, weapons/projectiles, NPC movement
* Sprites as world geometry
* "Pre-fabs" and instancing to enable more complex construction

## Repository Structure

* [/archetypes](/archetypes/) - these are preset or template combinations of
  components to build up more complex entities.
* [/concepts](/concepts/) - general types, interfaces, and functions.
  Math, serialization, etc... Includes the ECS database and associated storage
  methods for holding components.
* [/constants](/constants/) - self-explanatory. Currently a mix of game and
  engine constants.
* [/controllers](/controllers/) - the "systems" in the ECS architecture.
  Basically all the logic, and simulation. I prefer the name "controllers".
* [/components](/components/) - the  engine models and interfaces. As a general
  rule, code that changes model state based on another model belongs in
  [/controllers](/controllers/) rather than here.
  * [/behaviors](/components/behaviors/) - these modify how other components
    behave (for example, adding some movement logic to a Body)
  * [/core](/components/core/) - The simplest components like sectors, bodies,
    and lights.
  * [/materials](/components/materials/) - shaders, textures that affect the look of world
    geometry and bodies.
  * [/sectors](/components/sectors/) - these modify the behavior of basic
    sectors (e.g. underwater, doors, etc...)
* [/data](/data/) - a bunch of test data.
* [/editor](/editor/) - all the code for the world editor.
* [/game](/game/) - the game executable.
* [/render](/render/) - the renderer and its state.