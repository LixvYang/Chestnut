// Package chain provides chain for chestnut.
package chain

import (
	"time"

	chestnutpb "github.com/lixvyang/chestnut/pb"
)

type Syncer struct {
	nodeName         string
	group            *Group
	trxMgr           *TrxMgr
	AskNextTimer     *time.Timer
	AskNextTimeDone  chan bool
	Status           int8
	retryCount       int8
	statusBeforeFail int8
	responses        map[string]*chestnutpb.ReqBlockResp
	groupId          string
}
