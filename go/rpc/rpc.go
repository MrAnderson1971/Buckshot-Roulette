package rpc

const (
	Summary   = "summary"
	Action    = "action"
	Damage    = "damage"
	GameOver  = "gameOver"
	MoreItems = "moreItems"
	YourTurn  = "yourTurn"
	Reload    = "reload"
)

type DamageArgs struct {
	Damage int
	Target string
}

type YourTurnArgs struct {
	Player   string
	Opponent string
}

type Shell struct {
	Value int
}

func (s Shell) String() string {
	if s.Value == 0 {
		return "live"
	}
	return "blank"
}
