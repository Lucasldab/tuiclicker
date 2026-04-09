package ui

// creatureArt holds 12 ASCII art strings for the creature panel.
// Dimensions: [tier 0..3][branch 0=blood, 1=flesh, 2=bones]
// Each art string is exactly 5 lines (4 newlines), all lines <= 30 chars, pure ASCII.
var creatureArt [4][3]string

func init() {
	// --- Tier 0: ABERRANT SEED (same art for all branches — Phase 1 placeholder) ---
	// Rendered dim per UI-SPEC Creature Panel Contract.
	tier0 := "" +
		"    .  .  .\n" +
		"   ( o  o )\n" +
		"    >  --  <\n" +
		"   /|      |\\\n" +
		"  * |      | *"

	creatureArt[0][0] = tier0
	creatureArt[0][1] = tier0
	creatureArt[0][2] = tier0

	// --- Tier 1: NASCENT FORM ---

	// Blood: liquid, flowing, vein-like
	creatureArt[1][0] = "" +
		"    ~  ~  ~\n" +
		"  ~~( o  o)~~\n" +
		"   ~>  --  <~\n" +
		"  /|~ ~~~ ~|\\\n" +
		" * |~      |~*"

	// Flesh: bulging, layered, mass-like
	creatureArt[1][1] = "" +
		"    o  o  o\n" +
		"  .( O  O ).\n" +
		"   >  __  <\n" +
		"  /|.    .|\\\n" +
		" * |  ()  | *"

	// Bones: angular, spiked, skeletal
	creatureArt[1][2] = "" +
		"    +  +  +\n" +
		"  -( /  \\ )-\n" +
		"   >  ==  <\n" +
		"  /|  /\\  |\\\n" +
		" * |  \\/  | *"

	// --- Tier 2: GROTESQUE ---

	// Blood: spreading veins, more aggressive
	creatureArt[2][0] = "" +
		" ~~v  ~~~  v~~\n" +
		"~~ ( @  @ ) ~~\n" +
		"   ~~~>--<~~~\n" +
		"~~/|~~vvv~~|\\\n" +
		"* ~|v~~~~~v|~ *"

	// Flesh: grotesque, layered organs
	creatureArt[2][1] = "" +
		"  oOo  oOo  oOo\n" +
		" O( OO  OO )O\n" +
		"  >  oo  <\n" +
		" /|O .||. O|\\\n" +
		"* |O (  ) O| *"

	// Bones: angular, multi-spined
	creatureArt[2][2] = "" +
		" /+ /+  +\\ +\\\n" +
		"-(  /\\  /\\  )-\n" +
		"  > ==== <\n" +
		" /|/+  +\\|\\\n" +
		"* |/ \\/ \\ | *"

	// --- Tier 3: ABOMINATION ---

	// Blood: full vein coverage, dripping
	creatureArt[3][0] = "" +
		"vv~~vvv~vvv~~vv\n" +
		"~(@@@ ~~ @@@)~\n" +
		"~~ >v--v< ~~\n" +
		"~/|vvvvvvv|\\~\n" +
		"*~|vv~~~vv|~*"

	// Flesh: massive, layered, organ-covered
	creatureArt[3][1] = "" +
		"OoOoO OoO OoOoO\n" +
		"O( OOO  OOO )O\n" +
		" >  ooOoo  <\n" +
		"/|OoO .  . OoO|\\\n" +
		"*|OO(  )(  )OO|*"

	// Bones: maximum spines, fully skeletal
	creatureArt[3][2] = "" +
		"/+/+/+  +\\+\\+\\\n" +
		"-(  //\\\\//\\\\  )-\n" +
		"  >========<\n" +
		" /|/+/+ +\\+\\|\\\n" +
		"* |//\\/ \\/\\\\| *"
}

// GetCreatureArt returns the art string for the given tier and branch.
// tier: 0-3, branch: 0=blood, 1=flesh, 2=bones. Clamps to valid range.
func GetCreatureArt(tier, branch int) string {
	if tier < 0 || tier > 3 {
		tier = 0
	}
	if branch < 0 || branch > 2 {
		branch = 0
	}
	return creatureArt[tier][branch]
}
