package rpc

const (
	Summary   = "summary"
	Action    = "action"
	Damage    = "damage"
	GameOver  = "gameOver"
	MoreItems = "moreItems"
)

type DamageArgs struct {
	Damage int
	Target string
}
