package game

import (
	"Roulette/clientStubs"
	"Roulette/rpc"
	"Roulette/transport"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
)

var Shells = make([]rpc.Shell, 0, 8)

func RemoveFirst[T any](s *[]T) {
	*s = (*s)[1:]
}

var Settings = settings{1, false}
var Hp = make(map[string]int)
var Wg sync.WaitGroup

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

func TakeTurn(target string, other string, shooter string) string {
	shell := Shells[0]
	RemoveFirst(&Shells)
	var action string
	fmt.Printf("%s pulls the trigger, it's a %s shell!\n", shooter, shell.String())
	clientStubs.Action(fmt.Sprintf("%s fired a %s shell at %s!", shooter, shell.String(), target))

	if shell.Value == 0 { // live
		Hp[target] -= Settings.Damage
		fmt.Printf("%s's HP: %d\n", target, Hp[target])
		clientStubs.Summary(fmt.Sprintf("%s lost %d HP. Remaining HP: %d\n", target, Settings.Damage, Hp[target]))
		clientStubs.Damage(Settings.Damage, target)
		Settings.Damage = 1
		if Hp[target] == 0 {
			message := fmt.Sprintf("Game over! %s wins.\n", other)
			clientStubs.GameOver(message)
			transport.GameOver <- message
		}

		if Settings.CuffedOpponent {
			fmt.Println("Your opponent is cuffed!")
			action = "continue"
		} else {
			action = "switch"
		}
	} else {
		if shooter == target {
			action = "continue"
		} else {
			if Settings.CuffedOpponent {
				fmt.Println("Your opponent is cuffed!")
				action = "continue"
			} else {
				action = "switch"
			}
		}
	}
	return action
}

func CurrentTurn(player string, opponent string) {
	for {
		if len(Shells) == 0 {
			clientStubs.MoreItems()
			fmt.Println("Reloading the shotgun!")
			LoadShotgun()
		}
		fmt.Println("Options:")
		fmt.Println("1. Shoot yourself")
		fmt.Println("2. Shoot your opponent")
		for i, item := range NumberToItem {
			fmt.Printf("%d: %s (%d)\n", i+3, item.Description(), Items[item])
		}
		fmt.Print("Choose an option: ")
		var choice string

		select {
		case message := <-transport.GameOver:
			fmt.Println(message)
			Wg.Done()
			return
		default:
		}
		fmt.Scanln(&choice)
		if choice == "1" || choice == "2" {
			var action string
			if choice == "1" {
				action = TakeTurn(player, opponent, player)
			} else {
				action = TakeTurn(opponent, player, player)
			}
			if action == "switch" {
				clientStubs.YourTurn(player, opponent)
				break
			}
			if action == "continue" {
				fmt.Printf("%s gets another turn!\n", player)
			}
		} else if choice == "cheat" {
			fmt.Println(Shells)
		} else if choiceToInt, err := strconv.Atoi(choice); err == nil && choiceToInt >= 3 &&
			choiceToInt < len(NumberToItem)+3 && Items[NumberToItem[choiceToInt-3]] > 0 {
			NumberToItem[choiceToInt-3].Use(player)
			Items[NumberToItem[choiceToInt-3]]--
		} else if choiceToInt, err := strconv.Atoi(choice); err == nil && choiceToInt >= 3 &&
			choiceToInt < len(NumberToItem)+3 {
			fmt.Println("You don't have " + NumberToItem[choiceToInt-3].Name())
		} else {
			fmt.Println("Invalid choice. Please enter 1 or 2")
		}
	}
}

func LoadShotgun() {
	liveShells := rand.Intn(3) + 1
	blankShells := rand.Intn(3) + 1
	Shells = append(make([]rpc.Shell, liveShells), make([]rpc.Shell, blankShells)...)
	rand.Shuffle(len(Shells), func(i, j int) {
		Shells[i], Shells[j] = Shells[j], Shells[i]
	})
	for i := 0; i < liveShells; i++ {
		Shells[i] = rpc.Shell{Value: 0} // Live shell
	}
	for i := liveShells; i < liveShells+blankShells; i++ {
		Shells[i] = rpc.Shell{Value: 1} // Blank shell
	}
	fmt.Printf("[INFO] Shotgun loaded with %d live game.Shells and %d blank game.Shells (order is hidden).\n", liveShells, blankShells)
	shellValues := make([]string, len(Shells))
	for i, shell := range Shells {
		shellValues[i] = strconv.Itoa(shell.Value)
	}
	clientStubs.Reload(Shells)
	MoreItems()
}
