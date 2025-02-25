package main

import (
	"Roulette/game"
	"Roulette/serverStubs"
	"Roulette/transport"
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
)

var connection net.Conn

func main() {
	serverStubs.Register()
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

	var rpcListener net.Listener
	if mode == "join" {
		ipAddr, _, opponentName := transport.DiscoverHost()
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

		// Start joiner's RPC server
		rpcListener, err = net.Listen("tcp", "0.0.0.0:0")
		if err != nil {
			panic(err)
		}
		joinerRPCPort := rpcListener.Addr().(*net.TCPAddr).Port

		// Send RPC port to host
		fmt.Fprintf(connection, "%d\n", joinerRPCPort)

		// Read joiner's RPC port
		hostRPCPortString, err := bufio.NewReader(connection).ReadString('\n')
		hostRPCPort := strings.TrimSpace(hostRPCPortString)
		// Set host's RPC address
		hostIP, _, _ := net.SplitHostPort(connection.RemoteAddr().String())
		transport.Bind(fmt.Sprintf("%s:%s", hostIP, hostRPCPort))

		// Start RPC listener
		connection.Close()
	} else if mode == "host" {
		func() {
			listener, err := net.Listen("tcp", "0.0.0.0:0")
			if err != nil {
				panic(err)
			}
			port := listener.Addr().(*net.TCPAddr).Port
			discoveryBroadcast := &transport.DiscoveryBroadcast{}
			discoveryBroadcast.Start(playerName, port)
			defer discoveryBroadcast.Close()
			listener, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", transport.PORT))
			if err != nil {
				panic(err)
			}
			defer listener.Close()
			connection, err = listener.Accept()
			if err != nil {
				panic(err)
			}
			addr := connection.LocalAddr().String()
			fmt.Printf("Connected to %s\n", addr)

			reader := bufio.NewReader(connection)
			opponentName, err = reader.ReadString('\n')
			if err != nil {
				panic(err)
			}
			opponentName = strings.TrimSpace(opponentName)

			// After accepting connection and exchanging names
			rpcListener, err = net.Listen("tcp", "0.0.0.0:0")
			if err != nil {
				panic(err)
			}
			hostRPCPort := rpcListener.Addr().(*net.TCPAddr).Port

			// Send RPC port to joiner
			fmt.Fprintf(connection, "%d\n", hostRPCPort)

			// Read joiner's RPC port
			joinerRPCPortStr, err := bufio.NewReader(connection).ReadString('\n')
			joinerRPCPort := strings.TrimSpace(joinerRPCPortStr)

			// Determine joiner's IP and set RPC address
			joinerIP, _, _ := net.SplitHostPort(connection.RemoteAddr().String())
			transport.Bind(fmt.Sprintf("%s:%s", joinerIP, joinerRPCPort))

			// Start RPC listener
			connection.Close()
		}()
	}
	game.Hp[playerName] = 5
	game.Hp[opponentName] = 5
	fmt.Printf("%s's HP: %d\n", playerName, game.Hp[playerName])
	fmt.Printf("%s's HP: %d\n", opponentName, game.Hp[opponentName])

	game.Wg.Add(1)
	go transport.Listen(ctx, rpcListener)
	if mode == "host" {
		game.CurrentTurn(playerName, opponentName)
		fmt.Println("Waiting for your opponent's turn...")
	} else if mode == "join" {
		fmt.Println("Waiting for your turn...")
	}
	game.Wg.Wait()
	fmt.Scanln()
}
