package serverStubs

import (
	"Roulette/game"
	"Roulette/rpc"
	"Roulette/transport"
	"fmt"
)

func init() {
	transport.Register(rpc.Summary, summary)
}

func summary(argData []byte) (out []byte, err error) {
	return transport.ServerStub[string](out, func(message string) any {
		fmt.Println(message)
		return nil
	})
}

func action(argData []byte) (out []byte, err error) {
	return transport.ServerStub[string](out, func(message string) any {
		game.RemoveFirst(&game.Shells)
		fmt.Println(message)
		return nil
	})
}

func damage(argData []byte) (out []byte, err error) {
	return transport.ServerStub(out, func(args rpc.DamageArgs) any {

	})
}
