package main

// Store structure - Redis
/*
(KEY)board:{boardId}				(VALUE)board					Board - Redis Hash. The board details. Owned by single user.
(KEY)msg:{messageId}				(VALUE)message					Message - Redis Hash. Useful for fetch/add/update for an individual message.
(KEY)board:msg:{boardId}			(VALUE)[messageIds] 			Board-wise Messages - Redis Set. Useful for fetching list of messages.
(KEY)board:cmts:{boardId}      		(VALUE)[commentIds]      		Board-wise Comments - Redis Set. For fetching all comments.
(KEY)msg:likes:{messageId}			(VALUE)[userIds]				Likes - Redis Set. For recording likes/votes for a message.
(KEY)board:user:{boardId}:{userId}	(VALUE)User						User - Redis Hash. User master. Keeping as board specific.
(KEY)board:user:xid:seq:{boardId}	(VALUE)last_xid					Last generated sequential xid for Board - Redis INCR. Used to generate sequential Xids.
(KEY)board:presence:{boardId}		(VALUE)[userIds]				Board-wise Live(Connected) Users - Redis Set.
(KEY)board:col:{boardId}:{colId}	(VALUE)column					Column - Redis Hash. Column definition for a Board.
(KEY)board:col:{boardId}			(VALUE)[colIds]					Board-wise columns - Redis Set. Just a list of colIds for a board.
*/

// Base prefixes
const (
	keyBoard              = "board:"
	keyBoardMsgs          = "board:msg:"
	keyBoardCmts          = "board:cmts:"
	keyBoardUsersPresence = "board:presence:"
	keyBoardUser          = "board:user:"
	keyBoardUserXid       = "board:user:xid:seq:"
	keyBoardCols          = "board:col:"
	keyMsg                = "msg:"
	keyMsgLikes           = "msg:likes:"
)

// board:{boardId}.
// Board - Redis HASH.
func boardKey(boardId string) string {
	return keyBoard + boardId
}

// board:msg:{boardId}.
// Board-wise Messages - Redis SET.
func boardMsgsKey(boardId string) string {
	return keyBoardMsgs + boardId
}

// board:cmts:{boardId}
// Board-wise Comments - Redis SET.
func boardCmtsKey(boardId string) string {
	return keyBoardCmts + boardId
}

// board:presence:{boardId}.
// Board-wise Live(Connected) users - Redis SET.
func boardUsersPresenceKey(boardId string) string {
	return keyBoardUsersPresence + boardId
}

// board:user:xid:seq:{boardId}.
// Key for last generated sequential Xid - Redis INCR.
func boardUserXidKey(boardId string) string {
	return keyBoardUserXid + boardId
}

// board:col:{boardId}.
// Board-wise columns - Redis SET.
func boardColsKey(boardId string) string {
	return keyBoardCols + boardId
}

// msg:{messageId}.
// Message - Redis HASH.
func msgKey(msgId string) string {
	return keyMsg + msgId
}

// msg:likes:{messageId}.
// Likes - Redis SET.
func msgLikesKey(msgId string) string {
	return keyMsgLikes + msgId
}

// board:user:{boardId}:{userId}.
// User - Redis HASH.
func boardUserKey(boardId, userId string) string {
	return keyBoardUser + boardId + ":" + userId
}

// board:col:{boardId}:{colId}.
// Column - Redis HASH.
func boardColKey(boardId, colId string) string {
	return keyBoardCols + boardId + ":" + colId
}
