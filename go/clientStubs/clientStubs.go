package clientStubs

import (
	"Roulette/rpc"
	"Roulette/transport"
)

func Summary(message string) {
	transport.ClientStub[any](rpc.Summary, message)
}

func Action(message string) {
	transport.ClientStub[any](rpc.Action, message)
}

func Damage(damage int, target string) {
	transport.ClientStub[any](rpc.Damage, rpc.DamageArgs{Damage: damage, Target: target})
}

func GameOver(message string) {
	transport.ClientStub[any](rpc.GameOver, message)
}

func MoreItems() {
	transport.ClientStub[any](rpc.MoreItems, nil)
}

func YourTurn(player, opponent string) {
	transport.ClientStub[any](rpc.YourTurn, rpc.YourTurnArgs{Player: player, Opponent: opponent})
}

func Reload(shells []rpc.Shell) {
	transport.ClientStub[any](rpc.Reload, shells)
}
