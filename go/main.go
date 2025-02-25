package main

import (
	"Roulette/game"
	"Roulette/rpc"
	"Roulette/transport"
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var connection net.Conn

func splitLine(buffer string) (line, rest string) {
	idx := strings.Index(buffer, "\n")
	if idx == -1 {
		return buffer, ""
	}
	return buffer[:idx], buffer[idx+1:]
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
						game.CurrentTurn(player, opponent)
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
						game.Shells = append(game.Shells, rpc.Shell{atoi})
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
					game.MoreItems()
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
							game.Shells[0] = rpc.Shell{1}
						} else {
							game.Shells[0] = rpc.Shell{0}
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
		game.CurrentTurn(playerName, opponentName)
		fmt.Println("Waiting for your opponent's turn...")
	} else if mode == "join" {
		fmt.Println("Waiting for your turn...")
	}
}
