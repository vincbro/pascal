package database

import (
	"log/slog"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	Client *gorm.DB
}

func NewDatabase(dsn string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&User{})
	db.AutoMigrate(&Trip{})

	return &Database{
		Client: db,
	}, nil
}

func (d *Database) GetOrCreateUser(userID, username string) (*User, error) {
	user := &User{}
	result := d.Client.FirstOrCreate(user, User{ID: userID, Username: username})
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (d *Database) GetAllUsers() ([]*User, error) {
	users := []*User{}
	result := d.Client.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	if len(users) == 0 {
		slog.Warn("warning 0 users returned from GetAllUsers")
	}
	return users, nil
}

func (d *Database) UpdateUser(user *User) error {
	result := d.Client.Save(user)
	return result.Error
}

func (d *Database) AddTrip(trip *Trip) error {
	result := d.Client.Create(trip)
	return result.Error
}
