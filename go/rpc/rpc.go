package rpc

const (
	Summary = "summary"
	Action  = "action"
	Damage  = "damage"
)

type DamageArgs struct {
	Damage int
	Target string
}
