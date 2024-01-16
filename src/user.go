package main

// Store
type User struct {
	Id       string `redis:"id"`
	Xid      string `redis:"xid"`
	Nickname string `redis:"nickname"`
}
