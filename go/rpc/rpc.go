package rpc

const (
	Summary    = "summary"
	Action     = "action"
	Damage     = "damage"
	GameOver   = "gameOver"
	MoreItems  = "moreItems"
	YourTurn   = "yourTurn"
	Reload     = "reload"
	Eject      = "eject"
	Heal       = "heal"
	Invert     = "invert"
	Adrenaline = "adrenaline"
	Steal      = "steal"
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

type HealArgs struct {
	Amount  int
	Target  string
	Message string
}

func (s Shell) String() string {
	if s.Value == 0 {
		return "live"
	}
	return "blank"
}

type Item interface {
	Name() string
	Description() string
	Use(player string)
}
