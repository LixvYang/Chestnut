// Package chain provides chain for chestnut.
package chain

type Molasses struct {
	name string
	producer Producer
	user User
}

func NewMolasses(p Producer, u User) *Molasses {
	return &Molasses{name:"molasses", producer: p, user: u}
}

func (m *Molasses) Name() string {
	return m.name
}

func (m *Molasses) Producer() Producer {
	return m.producer
}

func (m *Molasses) User() User {
	return m.user
}
