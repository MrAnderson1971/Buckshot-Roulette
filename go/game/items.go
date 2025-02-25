package game

import (
	"Roulette/clientStubs"
	"fmt"
	"math/rand"
)

type Item interface {
	Name() string
	Description() string
	Use(player string)
}

type MagnifyingGlass struct{}

func (*MagnifyingGlass) Name() string {
	return "Magnifying Glass"
}

func (*MagnifyingGlass) Description() string {
	return "🔍 Reveals the next shell."
}

func (*MagnifyingGlass) Use(player string) {
	fmt.Printf("The next item is a %s shell.\n", Shells[0])
	clientStubs.Summary("Opponent used magnifying glass (very interesting)...")
}

type Cigarette struct{}

func (*Cigarette) Name() string {
	return "Cigarette"
}

func (*Cigarette) Description() string {
	return "🚬 Restore one HP."
}

func (*Cigarette) Use(player string) {
	Hp[player]++
	fmt.Println("Smoked one HP back.")
	clientStubs.Summary(fmt.Sprintf("heal:%s,1,Opponened smoked 1 HP.\n", player))
}

type Handsaw struct{}

func (*Handsaw) Name() string {
	return "Handsaw"
}

func (*Handsaw) Description() string {
	return "🪚 Next shot does double damage."
}

func (*Handsaw) Use(player string) {
	Settings.Damage = 2
	fmt.Println("Sawed off shotgun...")
	clientStubs.Summary("summary:Opponent used handsaw...\n")
}

type Beer struct{}

func (*Beer) Name() string {
	return "Beer"
}

func (*Beer) Description() string {
	return "🍺 Ejects the current shell."
}

func (*Beer) Use(player string) {
	first := Shells[0]
	RemoveFirst(&Shells)
	fmt.Printf("Ejected a %s shell.\n", first)
	SendMessage(fmt.Sprintf("eject:Opponent ejected a %s shell."))
}

type Handcuffs struct{}

func (*Handcuffs) Name() string {
	return "Handcuffs"
}

func (*Handcuffs) Description() string {
	return "🔗 Skips your opponent's turn."
}

func (*Handcuffs) Use(player string) {
	Settings.CuffedOpponent = true
	fmt.Println("Cuffed your opponent.")
	clientStubs.Summary("Opponent cuffed you!")
}

type Phone struct{}

func (*Phone) Name() string {
	return "Phone"
}

func (*Phone) Description() string {
	return "📱 A mysterious voice reveals insights from the future"
}

func (*Phone) Use(player string) {
	if len(Shells) <= 1 {
		fmt.Println("How unfortunate...")
	} else {
		selected := rand.Intn(len(Shells)-1) + 1
		fmt.Printf("Shell #%d, %s", selected+1, Shells[selected])
	}
	clientStubs.Summary("Opponent used phone...")
}

type Medicine struct{}

func (*Medicine) Name() string {
	return "Medicine"
}

func (*Medicine) Description() string {
	return "💊 50% chance to gain 2 HP. If not, lose 1 HP."
}

func (*Medicine) Use(player string) {
	if rand.Intn(2) == 1 {
		Hp[player] += 2
		fmt.Println("You gained 2 HP!")
		SendMessage(fmt.Sprintf("heal:%s,2,Opponent gained 2 HP!\n", player))
	} else {
		Hp[player]--
		fmt.Println("You collapsed! -1 HP.")
		SendMessage(fmt.Sprintf("heal:%s,-1,Opponent collapsed! They lose 1 HP\n", player))
	}
}

type Inverter struct{}

func (*Inverter) Name() string {
	return "Inverter"
}

func (*Inverter) Description() string {
	return "🪫 Reverses polarity of current shell."
}

func (*Inverter) Use(player string) {
	if len(Shells) > 0 {
		if Shells[0].Value == 0 {
			Shells[0] = Shell{1}
		} else {
			Shells[0] = Shell{0}
		}
	}
	fmt.Println("Inverted shell.")
	SendMessage("invert:\n")
}
