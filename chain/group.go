// Package chain provides chain for chestnut.
package chain

import (
	chestnutpb "github.com/lixvyang/chestnut/pb"
)

type Group struct {
	// Group Item
	Item *chestnutpb.GroupItem
	ChainCtx *Chain
}