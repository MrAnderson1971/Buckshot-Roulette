package main

import (
	"Roulette/clientStubs"
	"Roulette/game"
	"Roulette/transport"
	"bufio"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var connection net.Conn
var numberToItem = []Item{
	&MagnifyingGlass{},
	&Cigarette{},
	&Beer{},
	&Handsaw{},
	&Handcuffs{},
	&Phone{},
	&Medicine{},
	&Inverter{},
}
var items = map[Item]int{
	numberToItem[0]: 0,
	numberToItem[1]: 0,
	numberToItem[2]: 0,
	numberToItem[3]: 0,
	numberToItem[4]: 0,
	numberToItem[5]: 0,
	numberToItem[6]: 0,
	numberToItem[7]: 0,
}

func splitLine(buffer string) (line, rest string) {
	idx := strings.Index(buffer, "\n")
	if idx == -1 {
		return buffer, ""
	}
	return buffer[:idx], buffer[idx+1:]
}

func moreItems() {
	tempItems := make([]Item, 0, len(numberToItem))
	for _, stuff := range numberToItem {
		tempItems = append(tempItems, stuff)
	}
	chosenItems := make([]Item, 0)
	for i := 0; i < 2; i++ {
		selectedItem := rand.Intn(len(tempItems))
		chosenItems = append(chosenItems, tempItems[selectedItem])
		items[numberToItem[selectedItem]]++
	}
	sb := ""
	for _, chosenItem := range chosenItems {
		sb += chosenItem.Name() + ", "
	}
	sb += "."
	fmt.Println("You get " + sb)
	clientStubs.Summary("Opponent gets " + sb)
}

func handleIncomingMessages(ctx context.Context, player string, opponent string) {
	reader := bufio.NewReader(connection)
	buffer := ""
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := connection.SetReadDeadline(time.Now().Add(1 * time.Second))
			if err != nil {
				panic(fmt.Sprintf("Error %s", err))
			}
			data, err := reader.ReadString('\n')
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					// timeout loop back again
					continue
				}
				panic(fmt.Sprintf("Error %s", err))
			}
			buffer += data

			for strings.Contains(buffer, "\n") {
				var line string
				line, buffer = splitLine(buffer)
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				switch {
				case strings.HasPrefix(line, "control:"):
					msg := strings.TrimPrefix(line, "control:")
					switch {
					case strings.HasPrefix(msg, "continue"):
						fmt.Println(opponent + " got a blank! It's still their turn.")
					case strings.HasPrefix(msg, "your_turn"):
						currentTurn(player, opponent)
					default:
						fmt.Println(msg)
					}
				case strings.HasPrefix(line, "game_over:"):
					transport.GameOver <- strings.TrimPrefix(line, "game_over:")
				case strings.HasPrefix(line, "summary:"):
					fmt.Println(strings.TrimPrefix(line, "summary:"))
				case strings.HasPrefix(line, "action:"):
					game.RemoveFirst(&game.Shells)
					fmt.Println(opponent + "'s move: " + strings.TrimPrefix(line, "action:"))
				case strings.HasPrefix(line, "reload:"):
					game.Shells = game.Shells[:0]
					liveCount := 0
					blankCount := 0
					msg := strings.TrimPrefix(line, "reload:")
					for i := 0; i < len(msg); i++ {
						atoi, err := strconv.Atoi(string(msg[i]))
						if err != nil {
							fmt.Printf("Error converting %s to int: %s\n", msg[i], err)
							os.Exit(1)
						}
						game.Shells = append(game.Shells, game.Shell{atoi})
						if atoi == 0 {
							liveCount++
						} else if atoi == 1 {
							blankCount++
						} else {
							fmt.Printf("Unknown shell %d\n", atoi)
							os.Exit(1)
						}
					}
					fmt.Printf("[INFO] Shotgun loaded with %d live game.Shells and %d blank game.Shells (order is hidden).\n",
						liveCount, blankCount)
				case strings.HasPrefix(line, "damage:"):
					msg := strings.TrimPrefix(line, "damage:")
					parts := strings.Split(msg, ",")
					newHp, _ := strconv.Atoi(parts[0])
					target := parts[1]
					game.Hp[target] -= newHp
				case strings.HasPrefix(line, "moreitems:"):
					moreItems()
				case strings.HasPrefix(line, "heal:"):
					msg := strings.Split(strings.TrimPrefix(line, "heal:"), ",")
					newHp, _ := strconv.Atoi(msg[1])
					game.Hp[msg[0]] += newHp
					fmt.Println(msg[2])
				case strings.HasPrefix(line, "eject:"):
					if len(game.Shells) > 0 {
						game.RemoveFirst(&game.Shells)
					}
					fmt.Println(strings.TrimPrefix(line, "eject:"))
				case strings.HasPrefix(line, "invert:"):
					if len(game.Shells) > 0 {
						if game.Shells[0].Value == 0 {
							game.Shells[0] = game.Shell{1}
						} else {
							game.Shells[0] = game.Shell{0}
						}
					}
					fmt.Println("Opponent inverted shell.")
				default:
					fmt.Println(line)
				}
			}
		}
	}
}

func takeTurn(target string, other string, shooter string) string {
	shell := game.Shells[0]
	game.RemoveFirst(&game.Shells)
	var action string
	fmt.Printf("%s pulls the trigger, it's a %s shell!\n", shooter, shell.String())
	clientStubs.Action(fmt.Sprintf("%s fired a %s shell at %s!", shooter, shell.String(), target))

	if shell.Value == 0 { // live
		game.Hp[target] -= game.Settings.Damage
		fmt.Printf("%s's HP: %d\n", target, game.Hp[target])
		clientStubs.Summary(fmt.Sprintf("%s lost %d HP. Remaining HP: %d\n", target, game.Settings.Damage, game.Hp[target]))

		SendMessage(fmt.Sprintf("damage:%d,%s\n", game.Settings.Damage, target))
		game.Settings.Damage = 1
		if game.Hp[target] == 0 {
			message := fmt.Sprintf("Game over! %s wins.\n", other)
			SendMessage("game_over:" + message)
			transport.GameOver <- message
		}

		if game.Settings.CuffedOpponent {
			fmt.Println("Your opponent is cuffed!")
			action = "continue"
		} else {
			action = "switch"
		}
	} else {
		if shooter == target {
			action = "continue"
		} else {
			if game.Settings.CuffedOpponent {
				fmt.Println("Your opponent is cuffed!")
				action = "continue"
			} else {
				action = "switch"
			}
		}
	}
	return action
}

func currentTurn(player string, opponent string) {
	for {
		if len(game.Shells) == 0 {
			SendMessage("moreitems:\n")
			fmt.Println("Reloading the shotgun!")
			loadShotgun()
		}
		fmt.Println("Options:")
		fmt.Println("1. Shoot yourself")
		fmt.Println("2. Shoot your opponent")
		for i, item := range numberToItem {
			fmt.Printf("%d: %s (%d)\n", i+3, item.Description(), items[item])
		}
		fmt.Print("Choose an option: ")
		var choice string

		select {
		case message := <-transport.GameOver:
			fmt.Println(message)
			return
		default:
		}
		fmt.Scanln(&choice)
		if choice == "1" || choice == "2" {
			var action string
			if choice == "1" {
				action = takeTurn(player, opponent, player)
			} else {
				action = takeTurn(opponent, player, player)
			}
			if action == "switch" {
				SendMessage("control:your_turn\n")
				break
			}
			if action == "continue" {
				fmt.Printf("%s gets another turn!\n", player)
			}
		} else if choice == "cheat" {
			fmt.Println(game.Shells)
		} else if choiceToInt, err := strconv.Atoi(choice); err == nil && choiceToInt >= 3 &&
			choiceToInt < len(numberToItem)+3 && items[numberToItem[choiceToInt-3]] > 0 {
			numberToItem[choiceToInt-3].Use(player)
			items[numberToItem[choiceToInt-3]]--
		} else if choiceToInt, err := strconv.Atoi(choice); err == nil && choiceToInt >= 3 &&
			choiceToInt < len(numberToItem)+3 {
			fmt.Println("You don't have " + numberToItem[choiceToInt-3].Name())
		} else {
			fmt.Println("Invalid choice. Please enter 1 or 2")
		}
	}
}

func loadShotgun() {
	liveShells := rand.Intn(3) + 1
	blankShells := rand.Intn(3) + 1
	game.Shells = append(make([]game.Shell, liveShells), make([]game.Shell, blankShells)...)
	rand.Shuffle(len(game.Shells), func(i, j int) {
		game.Shells[i], game.Shells[j] = game.Shells[j], game.Shells[i]
	})
	for i := 0; i < liveShells; i++ {
		game.Shells[i] = game.Shell{Value: 0} // Live shell
	}
	for i := liveShells; i < liveShells+blankShells; i++ {
		game.Shells[i] = game.Shell{Value: 1} // Blank shell
	}
	fmt.Printf("[INFO] Shotgun loaded with %d live game.Shells and %d blank game.Shells (order is hidden).\n", liveShells, blankShells)
	shellValues := make([]string, len(game.Shells))
	for i, shell := range game.Shells {
		shellValues[i] = strconv.Itoa(shell.Value)
	}
	msg := "reload:" + strings.Join(shellValues, "") + "\n"
	_, err := connection.Write([]byte(msg))
	if err != nil {
		panic(fmt.Sprintf("Error %s", err))
	}
	moreItems()
}

func main() {
	var playerName string
	var opponentName string
	defer func() {
		if connection != nil {
			connection.Close()
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Print("Enter your name: ")
	if _, err := fmt.Scanln(&playerName); err != nil {
		return
	}
	fmt.Print("Do you want to host or join? (host/join): ")
	var mode string
	if _, err := fmt.Scanln(&mode); err != nil {
		panic(err)
	}
	for mode != "host" && mode != "join" {
		fmt.Print("Do you want to host or join? (host/join): ")
		var mode string
		if _, err := fmt.Scanln(&mode); err != nil {
			panic(err)
		}
	}

	if mode == "join" {
		var ipAddr string
		ipAddr, _, opponentName = transport.DiscoverHost()
		connection, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ipAddr, transport.PORT))
		if err != nil {
			panic(fmt.Sprintf("Error %s", err))
		}
		fmt.Printf("Connected to %s\n", ipAddr)
		for playerName == opponentName {
			fmt.Print("Name cannot be opponent playerName")
			if _, err = fmt.Scanln(&playerName); err != nil {
				panic(fmt.Sprintf("Error %s", err))
			}
		}
		if _, err = connection.Write([]byte(playerName + "\n")); err != nil {
			panic(fmt.Sprintf("Error %s", err))
		}
	} else if mode == "host" {
		func() {
			discoveryBroadcast := &transport.DiscoveryBroadcast{}
			discoveryBroadcast.Start(playerName)
			defer discoveryBroadcast.Close()
			listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", transport.PORT))
			if err != nil {
				panic(fmt.Sprintf("Error %s", err))
			}
			defer listener.Close()
			connection, err = listener.Accept()
			if err != nil {
				panic(fmt.Sprintf("Error %s", err))
			}
			addr := connection.LocalAddr().String()
			fmt.Printf("Connected to %s\n", addr)

			if _, err = connection.Write([]byte(playerName + "\n")); err != nil {
				panic(fmt.Sprintf("Error %s", err))
			}

			reader := bufio.NewReader(connection)
			opponentName, err = reader.ReadString('\n')
			if err != nil {
				panic(fmt.Sprintf("Error %s", err))
			}
			opponentName = strings.TrimSpace(opponentName)
		}()
	}
	game.Hp[playerName] = 5
	game.Hp[opponentName] = 5
	fmt.Printf("%s's HP: %d\n", playerName, game.Hp[playerName])
	fmt.Printf("%s's HP: %d\n", opponentName, game.Hp[opponentName])

	go handleIncomingMessages(ctx, playerName, opponentName)
	if mode == "host" {
		currentTurn(playerName, opponentName)
		fmt.Println("Waiting for your opponent's turn...")
	} else if mode == "join" {
		fmt.Println("Waiting for your turn...")
	}
}
