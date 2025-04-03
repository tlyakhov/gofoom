# gofoom Inventory

Here's how all these components are connected:

- A `Player` or NPC has an `Carrier` component attached, which enables
  that character to have an array of `Slot`s.
- A `Slot` can have a `WeaponClass` component attached to it, which would
  indicate the traits of the weapon the character is holding.
- `WeaponClass` describes the traits of a given type of weapon. The WeaponClass
  could be unique for that particular NPC/player, or it could be a common
  component.
- `Weapon` components hold the state of a particular player/NPC's weapon. They
  should be created automatically by the game as necessary, see InventorySlotController.
