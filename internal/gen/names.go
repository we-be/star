package gen

import "math/rand/v2"

// Procedural star system name generator.
// Produces names like "Alpherion", "Velanthos", "Cygnaris".

var starPrefixes = []string{
	"Alph", "Ald", "Ant", "Arc", "Aur", "Bel", "Can",
	"Cap", "Cas", "Cen", "Cyg", "Del", "Den", "Dra",
	"Erid", "Fom", "Gem", "Had", "Hyd", "Ind", "Kep",
	"Lyr", "Mim", "Min", "Mir", "Nav", "Nix", "Ori",
	"Per", "Pol", "Pro", "Rig", "Sag", "Ser", "Sir",
	"Sol", "Spi", "Tau", "Vel", "Veg", "Vol", "Zan",
}

var starMids = []string{
	"ar", "er", "an", "en", "al", "or", "eth", "ant",
	"oph", "eon", "ion", "ax", "el", "os", "at", "ir",
}

var starSuffixes = []string{
	"is", "os", "us", "a", "on", "ax", "ar", "ius",
	"ion", "eon", "oth", "ine", "ira", "ora", "ais",
}

func SystemName(rng *rand.Rand) string {
	pre := starPrefixes[rng.IntN(len(starPrefixes))]
	mid := ""
	if rng.IntN(10) < 5 {
		mid = starMids[rng.IntN(len(starMids))]
	}
	suf := starSuffixes[rng.IntN(len(starSuffixes))]
	return pre + mid + suf
}
