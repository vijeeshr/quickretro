package main

// Store
type User struct {
	Id       string `redis:"id"`
	Xid      string `redis:"xid"` // Todo: Change to Int
	Nickname string `redis:"nickname"`
}
