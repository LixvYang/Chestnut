// Package api provides API for chestnut.
package api


//const name
const (
	GROUP_NAME string = "group_name"
	GROUP_ID string = "group_id"
	GROUP_OWNER_PUBKEY string = "owner_pubkey"
	GROUP_ITEMS string = "group_items"
	GROUP_ITEM string = "group_item"
	GENESIS_BLOCK string = "genesis_block"
	NODE_VERSION string = "node_version"
	NODE_PUBKEY string = "node_publickey"
	NODE_STATUS string = "node_status"
	NODE_ID string = "node_id"
	SIGNATURE string = "signature"
	TRX_ID string = "trx_id"
	PEERS string = "peers"
	NODETYPE string = "node_type"
)

//config
const GROUP_NAME_MIN_LENGTH int = 5

//error
const ERROR_INFO string = "error"

const (
	Add      = "Add"
	Update   = "Update"
	Remove   = "Remove"
	Group    = "Group"
	User     = "User"
	Auth     = "Auth"
	Note     = "Note"
	Producer = "Producer"
	Announce = "Announce"
	App      = "App"
)

const BLACK_LIST_OP_PREFIX string = "blklistop_"