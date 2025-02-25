package game

import (
	"Roulette/clientStubs"
	"fmt"
	"math/rand"
)

var Shells = make([]Shell, 0, 8)

func RemoveFirst[T any](s *[]T) {
	*s = (*s)[1:]
}

var Settings = settings{1, false}
var Hp = make(map[string]int)

type Shell struct {
	Value int
}

func (s Shell) String() string {
	if s.Value == 0 {
		return "live"
	}
	return "blank"
}

type settings struct {
	Damage         int
	CuffedOpponent bool
}

func MoreItems() {
	tempItems := make([]Item, 0, len(NumberToItem))
	for _, stuff := range NumberToItem {
		tempItems = append(tempItems, stuff)
	}
	chosenItems := make([]Item, 0)
	for i := 0; i < 2; i++ {
		selectedItem := rand.Intn(len(tempItems))
		chosenItems = append(chosenItems, tempItems[selectedItem])
		Items[NumberToItem[selectedItem]]++
	}
	sb := ""
	for _, chosenItem := range chosenItems {
		sb += chosenItem.Name() + ", "
	}
	sb += "."
	fmt.Println("You get " + sb)
	clientStubs.Summary("Opponent gets " + sb)
}

var NumberToItem = []Item{
	&MagnifyingGlass{},
	&Cigarette{},
	&Beer{},
	&Handsaw{},
	&Handcuffs{},
	&Phone{},
	&Medicine{},
	&Inverter{},
}
var Items = map[Item]int{
	NumberToItem[0]: 0,
	NumberToItem[1]: 0,
	NumberToItem[2]: 0,
	NumberToItem[3]: 0,
	NumberToItem[4]: 0,
	NumberToItem[5]: 0,
	NumberToItem[6]: 0,
	NumberToItem[7]: 0,
}
