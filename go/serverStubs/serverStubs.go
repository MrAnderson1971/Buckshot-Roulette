package serverStubs

import (
	"Roulette/game"
	"Roulette/rpc"
	"Roulette/transport"
	"fmt"
)

func init() {
	transport.Register(rpc.Summary, summary)
	transport.Register(rpc.Action, action)
	transport.Register(rpc.Damage, damage)
	transport.Register(rpc.GameOver, gameOver)
	transport.Register(rpc.MoreItems, moreItems)
	transport.Register(rpc.YourTurn, yourTurn)
	transport.Register(rpc.Reload, reload)
}

func summary(argData []byte) (out []byte, err error) {
	return transport.ServerStub(argData, func(message string) any {
		fmt.Println(message)
		return nil
	})
}

func action(argData []byte) (out []byte, err error) {
	return transport.ServerStub(argData, func(message string) any {
		game.RemoveFirst(&game.Shells)
		fmt.Println(message)
		return nil
	})
}

func damage(argData []byte) (out []byte, err error) {
	return transport.ServerStub(argData, func(args rpc.DamageArgs) any {
		game.Hp[args.Target] -= args.Damage
		return nil
	})
}

func gameOver(argData []byte) (out []byte, err error) {
	return transport.ServerStub(argData, func(message string) any {
		transport.GameOver <- message
		return nil
	})
}

func moreItems(argData []byte) (out []byte, err error) {
	return transport.ServerStub(argData, func(any) any {
		game.MoreItems()
		return nil
	})
}

func yourTurn(argData []byte) (out []byte, err error) {
	return transport.ServerStub(argData, func(args rpc.YourTurnArgs) any {
		game.CurrentTurn(args.Opponent, args.Player) // reverse
		return nil
	})
}

func reload(argData []byte) (out []byte, err error) {
	return transport.ServerStub(argData, func(shells []rpc.Shell) any {
		game.Shells = game.Shells[:0]
		copy(game.Shells, shells)
		return nil
	})
}
