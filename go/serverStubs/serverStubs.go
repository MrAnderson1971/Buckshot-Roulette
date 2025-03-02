package serverStubs

import (
	"Roulette/game"
	"Roulette/rpc"
	"Roulette/transport"
	"fmt"
)

func Register() {
	transport.Register(rpc.Summary, summary)
	transport.Register(rpc.Action, action)
	transport.Register(rpc.Damage, damage)
	transport.Register(rpc.GameOver, gameOver)
	transport.Register(rpc.MoreItems, moreItems)
	transport.Register(rpc.YourTurn, yourTurn)
	transport.Register(rpc.Reload, reload)
	transport.Register(rpc.Eject, eject)
	transport.Register(rpc.Heal, heal)
	transport.Register(rpc.Invert, invert)
	transport.Register(rpc.Adrenaline, adrenaline)
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
		go game.CurrentTurn(args.Opponent, args.Player) // reverse
		return nil
	})
}

func reload(argData []byte) (out []byte, err error) {
	return transport.ServerStub(argData, func(shells []rpc.Shell) any {
		game.Shells = game.Shells[:0]
		var liveCount, blankCount int
		for _, shell := range shells {
			game.Shells = append(game.Shells, shell)
			if shell.Value == 0 {
				liveCount++
			} else {
				blankCount++
			}
		}
		fmt.Printf("[INFO] Shotgun loaded with %d live game.Shells and %d blank game.Shells (order is hidden).\n",
			liveCount, blankCount)
		return nil
	})
}

func eject(argData []byte) (out []byte, err error) {
	return transport.ServerStub(argData, func(message string) any {
		if len(game.Shells) > 0 {
			game.RemoveFirst(&game.Shells)
		}
		fmt.Println(message)
		return nil
	})
}

func heal(argData []byte) (out []byte, err error) {
	return transport.ServerStub(argData, func(args rpc.HealArgs) any {
		game.Hp[args.Target] += args.Amount
		fmt.Println(args.Message)
		return nil
	})
}

func invert(argData []byte) (out []byte, err error) {
	return transport.ServerStub(argData, func(any) any {
		game.Shells[0].Value = 1 - game.Shells[0].Value
		fmt.Println("Opponent used inverter...")
		return nil
	})
}

func adrenaline(argData []byte) (out []byte, err error) {
	return transport.ServerStub[int](argData, func(num int) any {
		return game.Items[game.NumberToItem[num]]
	})
}
