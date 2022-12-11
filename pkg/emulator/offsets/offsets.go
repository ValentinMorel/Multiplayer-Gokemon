package offsets

// Player positional data
const (
	MAP_ID    uint16 = 0xD35E
	MAP_Y     uint16 = 0xD361
	MAP_X     uint16 = 0xD362
	PLAYER_DY uint16 = 0xC103
	PLAYER_DX uint16 = 0xC105

	// The direction which the player is facing (0: down, 4: up, 8: left, 12: right)
	PLAYER_DIR uint16 = 0xC109

	// When a player moves, this value counts down from 8 to 0
	WALK_COUNTER uint16 = 0xCFC5

	// General player data
	PLAYER_NAME_START uint16 = 0xD158

	// The address of the player spritesheet encoded as 2bpp in the rom
	RED_SPRITE_ADDR  uint16 = 0x4180
	RED_SPRITE_BANK  uint32 = 5
	BLUE_SPRITE_ADDR uint16 = 0x4300
	BLUE_SPRITE_BANK uint32 = 5
	OAK_SPRITE_ADDR  uint16 = 0x4480
	OAK_SPRITE_BANK  uint32 = 5

	// The address of the main font encoded as a 1bpp sprite in the rom
	FONT_ADDR uint16 = 0x5A80
	FONT_BANK uint32 = 4

	// The address of the textbox border encoded as 2bpp sprite in the rom
	BORDER_ADDR uint16 = 0x6288 + 2*8*(4*6+1)
	BORDER_BANK uint32 = 4

	// The location of the tile map
	TILE_MAP uint16 = 0xC3A0

	// Useful addresses for hacks
	LOADED_ROM_BANK uint16 = 0xFFB8
	FRAME_COUNTER   uint16 = 0xFFD5
	BANK_SWITCH     uint16 = 0x35D6

	// Addresses for sprite check hack
	NUM_SPRITES          uint16 = 0xD4E1
	OVERWORLD_LOOP_START uint16 = 0x03FF
	SPRITE_CHECK_START   uint16 = 0x0B23
	SPRITE_CHECK_EXIT_1  uint16 = 0x0BA0
	SPRITE_CHECK_EXIT_2  uint16 = 0x0BC4
	SPRITE_INDEX         uint16 = 0xFF8C

	// Addresses for sprite update hack
	CLEAR_SPRITES   uint16 = 0x0082
	UPDATE_SPRITES  uint16 = 0x2429
	SPRITES_ENABLED uint16 = 0xCFCB

	// Addresses for display text hack
	DISPLAY_TEXT_ID            uint16 = 0x2920
	DISPLAY_TEXT_ID_AFTER_INIT uint16 = 0x292B
	DISPLAY_TEXT_SETUP_DONE    uint16 = 0x29CD
	GET_NEXT_CHAR_1            uint16 = 0x1B55
	GET_NEXT_CHAR_2            uint16 = 0x1956
	TEXT_PROCESSOR_END         uint16 = 0x1B5E

	// Addresses for battle hack
	TRAINER_CLASS       uint16 = 0xD031
	TRAINER_NAME        uint16 = 0xD04A
	TRAINER_NUM         uint16 = 0xD05D
	ACTIVE_BATTLE       uint16 = 0xD057
	CURRRENT_OPPONENT   uint16 = 0xD059
	CURRENT_ENEMY_LEVEL uint16 = 0xD127
	CURRENT_ENEMY_NICK  uint16 = 0x0000
	BATTLE_TYPE         uint16 = 0xD05A
	IS_LINK_BATTLE      uint16 = 0xD12B

	// The Prof. Oak battle is unused by the game, so it is a convenient place to replace with our
	// battle data.
	PROF_OAK_DATA_ADDR uint16 = 0x621D
	PROF_OAK_DATA_BANK uint32 = 0xE

	// Addresses for battle data
	PLAYER_BATTLE_DATA_START uint16 = 0xD163
	ENEMY_BATTLE_DATA_START  uint16 = 0xD89C
	ENEMY_NAME_START         uint16 = 0xD887

	// Addresses for specific party data
	PARTY_COUNT  uint16 = 0xD163
	PARTY_POKE_1 uint16 = 0xD16B
	PARTY_POKE_2 uint16 = 0xD197
	PARTY_POKE_3 uint16 = 0xD1C3
	PARTY_POKE_4 uint16 = 0xD1EF
	PARTY_POKE_5 uint16 = 0xD21B
	PARTY_POKE_6 uint16 = 0xD247
)
