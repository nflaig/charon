// Copyright © 2022 Obol Labs Inc.
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of  MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along with
// this program.  If not, see <http://www.gnu.org/licenses/>.

package cluster

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
)

var (
	// list of 144 nouns.
	nouns = []string{
		"adult",
		"age",
		"amount",
		"area",
		"back",
		"bed",
		"blood",
		"body",
		"book",
		"box",
		"boy",
		"bulb",
		"bunch",
		"business",
		"camera",
		"chicken",
		"child",
		"chocolates",
		"city",
		"clothes",
		"colony",
		"colors",
		"company",
		"computer",
		"continent",
		"council",
		"country",
		"course",
		"cycle",
		"dates",
		"day",
		"death",
		"desk",
		"door",
		"egg",
		"face",
		"fact",
		"factory",
		"family",
		"farm",
		"farmer",
		"father",
		"fish",
		"floor",
		"flowers",
		"food",
		"fridge",
		"future",
		"game",
		"garden",
		"gas",
		"glass",
		"group",
		"health",
		"hill",
		"hospital",
		"idea",
		"image",
		"industry",
		"island",
		"jewelry",
		"job",
		"kitchen",
		"land",
		"law",
		"leaves",
		"leg",
		"letter",
		"life",
		"magazine",
		"market",
		"metal",
		"mirror",
		"mobile",
		"money",
		"morning",
		"mother",
		"mountain",
		"movie",
		"name",
		"nest",
		"news",
		"ocean",
		"oil",
		"painter",
		"park",
		"party",
		"pen",
		"pen",
		"pencil",
		"person",
		"picture",
		"pillow",
		"place",
		"plant",
		"pond",
		"rain",
		"rate",
		"result",
		"ring",
		"road",
		"rock",
		"rocket",
		"room",
		"rope",
		"rule",
		"sale",
		"school",
		"shape",
		"shapes",
		"ship",
		"shop",
		"sister",
		"site",
		"skin",
		"snacks",
		"son",
		"song",
		"sort",
		"sound",
		"soup",
		"sports",
		"state",
		"stone",
		"street",
		"system",
		"taxi",
		"tea",
		"teacher",
		"team",
		"toy",
		"tractor",
		"trade",
		"train",
		"video",
		"view",
		"water",
		"waterfall",
		"week",
		"women",
		"wood",
		"word",
		"year",
		"yesterday",
	}

	// list of 225 adjectives.
	adjectives = []string{
		"adorable",
		"adventurous",
		"aggressive",
		"agreeable",
		"alert",
		"alive",
		"amused",
		"angry",
		"annoyed",
		"annoying",
		"anxious",
		"arrogant",
		"ashamed",
		"attractive",
		"average",
		"awful",
		"bad",
		"beautiful",
		"better",
		"bewildered",
		"black",
		"bloody",
		"blue",
		"blushing",
		"bored",
		"brainy",
		"brave",
		"breakable",
		"bright",
		"busy",
		"calm",
		"careful",
		"cautious",
		"charming",
		"cheerful",
		"clean",
		"clear",
		"clever",
		"cloudy",
		"clumsy",
		"colorful",
		"combative",
		"comfortable",
		"concerned",
		"condemned",
		"confused",
		"cooperative",
		"courageous",
		"crazy",
		"creepy",
		"crowded",
		"cruel",
		"curious",
		"cute",
		"dangerous",
		"dark",
		"dead",
		"defeated",
		"defiant",
		"delightful",
		"depressed",
		"determined",
		"different",
		"difficult",
		"disgusted",
		"distinct",
		"disturbed",
		"dizzy",
		"doubtful",
		"drab",
		"dull",
		"eager",
		"easy",
		"elated",
		"elegant",
		"embarrassed",
		"enchanting",
		"encouraging",
		"energetic",
		"enthusiastic",
		"envious",
		"evil",
		"excited",
		"expensive",
		"exuberant",
		"fair",
		"faithful",
		"famous",
		"fancy",
		"fantastic",
		"fierce",
		"filthy",
		"fine",
		"foolish",
		"fragile",
		"frail",
		"frantic",
		"friendly",
		"frightened",
		"funny",
		"gentle",
		"gifted",
		"glamorous",
		"gleaming",
		"glorious",
		"good",
		"gorgeous",
		"graceful",
		"grieving",
		"grotesque",
		"grumpy",
		"handsome",
		"happy",
		"healthy",
		"helpful",
		"helpless",
		"hilarious",
		"homeless",
		"homely",
		"horrible",
		"hungry",
		"hurt",
		"ill",
		"important",
		"impossible",
		"inexpensive",
		"innocent",
		"inquisitive",
		"itchy",
		"jealous",
		"jittery",
		"jolly",
		"joyous",
		"kind",
		"lazy",
		"light",
		"lively",
		"lonely",
		"long",
		"lovely",
		"lucky",
		"magnificent",
		"misty",
		"modern",
		"motionless",
		"muddy",
		"mushy",
		"mysterious",
		"nasty",
		"naughty",
		"nervous",
		"nice",
		"nutty",
		"obedient",
		"obnoxious",
		"odd",
		"open",
		"outrageous",
		"outstanding",
		"panicky",
		"perfect",
		"plain",
		"pleasant",
		"poised",
		"poor",
		"powerful",
		"precious",
		"prickly",
		"proud",
		"putrid",
		"puzzled",
		"quaint",
		"real",
		"relieved",
		"repulsive",
		"rich",
		"scary",
		"selfish",
		"shiny",
		"shy",
		"silly",
		"sleepy",
		"smiling",
		"smoggy",
		"sore",
		"sparkling",
		"splendid",
		"spotless",
		"stormy",
		"strange",
		"stupid",
		"successful",
		"super",
		"talented",
		"tame",
		"tasty",
		"tender",
		"tense",
		"terrible",
		"thankful",
		"thoughtful",
		"thoughtless",
		"tired",
		"tough",
		"troubled",
		"ugliest",
		"ugly",
		"uninterested",
		"unsightly",
		"unusual",
		"upset",
		"uptight",
		"vast",
		"victorious",
		"vivacious",
		"wandering",
		"weary",
		"wicked",
		"wild",
		"witty",
		"worried",
		"worrisome",
		"wrong",
		"zany",
		"zealous",
	}
)

// randomName returns a deterministic name for an ecdsa public key. The name consists of a noun
// and an adjective separated by a hyphen. The noun is calculated using PublicKey's X coordinate
// while the adjective is calculated using PublicKey's Y coordinate.
func randomName(pk ecdsa.PublicKey) string {
	// calculate the index of the adjective using X % ADJ_LEN
	adjLen := big.NewInt(int64(len(adjectives)))
	adjIdx := new(big.Int).Rem(pk.X, adjLen).Uint64()

	// similarly, calculate the index of the noun using Y % NOUN_LEN
	nounLen := big.NewInt(int64(len(nouns)))
	nounIdx := new(big.Int).Rem(pk.Y, nounLen).Uint64()

	return fmt.Sprintf("%s-%s", adjectives[adjIdx], nouns[nounIdx])
}
