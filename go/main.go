package main

import (
	"Roulette/game"
	"Roulette/transport"
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
)

var connection net.Conn

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

	game.Wg.Add(1)
	go transport.Listen(ctx)
	if mode == "host" {
		game.CurrentTurn(playerName, opponentName)
		fmt.Println("Waiting for your opponent's turn...")
	} else if mode == "join" {
		fmt.Println("Waiting for your turn...")
	}
	game.Wg.Wait()
	fmt.Scanln()
}
