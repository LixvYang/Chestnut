// Package appdata provides storage for chestnut.
package appdata

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	logging "github.com/ipfs/go-log/v2"
	"github.com/lixvyang/chestnut/storage"
	chestnutpb "github.com/lixvyang/chestnut/pb"
	"github.com/google/orderedcode"
)

type AppDb struct {
	Db storage.ChestnutStorage
	seq map[string]storage.Sequence
	DataPath string
}

var appdatalog = logging.Logger("appdata")

const (
	GRP_PREFIX = "grp_"
	CNT_PREFIX string = "cnt_"
	SDR_PREFIX string = "sdr_"
	SEQ_PREFIX string = "seq_"
	TRX_PREFIX string = "trx_"
	STATUS_PREFIX string = "stu_"
	term = "\x00\x01"
)

func NewAppDb() *AppDb {
	app := AppDb{}
	app.seq = make(map[string]storage.Sequence)
	return &app
}

func (appdb *AppDb) GetGroupStatus(groupid, name string) (string, error) {
	key := fmt.Sprintf("%s%s_%s", STATUS_PREFIX, groupid, name)
	exist, err := appdb.Db.IsExist([]byte(key))
	if err != nil {
		return "", err
	}
	if !exist {
		return "", err
	}
	value, _ := appdb.Db.Get([]byte(key))
	return string(value), nil
}

func (appdb *AppDb) GetSeqId(seqkey string) (uint64, error) {
	if appdb.seq[seqkey] == nil {
		seq, err := appdb.Db.GetSequence([]byte(seqkey), 100)
		if err != nil {
			return 0, err
		}
		appdb.seq[seqkey] = seq
	}
	return appdb.seq[seqkey].Next()
}

func (appdb *AppDb) Rebuild(vertag string, chainDb *badger.DB) error {
	return nil
}

func (appdb *AppDb) GetGroupContentBySenders(groupid string, senders []string, starttrx string, num int, reverse bool) ([]string, error) {
	prefix := fmt.Sprintf("%s%s-%s", CNT_PREFIX, GRP_PREFIX, groupid)
	sendermap := make(map[string]bool)
	for _, s := range senders {
		sendermap[s] = true
	}
	trxids := []string{}

	p := []byte(prefix)
	if reverse == true {
		p = append(p, 0xff, 0xff, 0xff, 0xff) // add the postfix 0xffffffff, badger will search the seqid <= 4294967295, it's big enough?
	}

	runcollector := false
	if starttrx == "" {
		runcollector = true //no trxid, start collecting from the first item
	}

	err := appdb.Db.PrefixForeachKey(p, []byte(prefix), reverse, func(k []byte, err error) error {
		if err != nil {
			return err
		}

		dataidx := bytes.LastIndexByte(k, byte('_'))
		trxid := string(k[len(k)-37-1 : len(k)-1-1])
		if runcollector == true {
			sender := string(k[dataidx+1+2 : len(k)-37-2]) //+2/-2 for remove the term, len(term)=2
			if len(senders) == 0 || sendermap[sender] == true {
				trxids = append(trxids, trxid)
			}
		}
		if trxid == starttrx { //start collecting after this item
			runcollector = true
		}
		if len(trxids) == num {
			// use this to break loop
			return errors.New("OK")
		}
		return nil
	})

	if err != nil && err.Error() == "OK" {
		err = nil
	}

	return trxids, err
}

func getKey(prefix string, seqid uint64, tailing string) ([]byte, error) {
	return orderedcode.Append(nil, prefix, "-", orderedcode.Infinity, uint64(seqid), "_", tailing)
}

func (appdb *AppDb) AddMetaByTrx(blockId string, groupid string, trxs []*chestnutpb.Trx) error {
	var err error

	seqkey := SEQ_PREFIX + CNT_PREFIX + GRP_PREFIX + groupid

	keylist := [][]byte{}
	for _, trx := range trxs {
		if trx.Type == chestnutpb.TrxType_POST {
			seqid, err := appdb.GetSeqId(seqkey)
			if err != nil {
				return err
			}

			key, err := getKey(fmt.Sprintf("%s%s-%s", CNT_PREFIX, GRP_PREFIX, groupid), seqid, fmt.Sprintf("%s:%s", trx.SenderPubkey, trx.TrxId))
			if err != nil {
				return err
			}
			keylist = append(keylist, key)
		}
	}

	keys := [][]byte{}
	values := [][]byte{}

	for _, key := range keylist {
		keys = append(keys, key)
		values = append(values, nil)
	}

	valuename := "HighestBlockId"
	groupLastestBlockidkey := fmt.Sprintf("%s%s_%s", STATUS_PREFIX, groupid, valuename)
	keys = append(keys, []byte(groupLastestBlockidkey))
	values = append(values, []byte(blockId))

	err = appdb.Db.BatchWrite(keys, values)

	return err
}

func (appdb *AppDb) Release() error {
	for seqkey := range appdb.seq {
		err := appdb.seq[seqkey].Release()
		if err != nil {
			return err
		}
	}
	return nil
}

func (appdb *AppDb) Close() {
	appdb.Release()
	appdb.Db.Close()
}


