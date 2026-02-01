package database

type User struct {
	ID       string `grom:"primaryKey"`
	Username string
}
