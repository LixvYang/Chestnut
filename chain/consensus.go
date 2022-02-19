// Package chain provides chain for chestnut.
package chain

type Consensus interface {
	Name() string
	Producer() Producer
	User() User
}