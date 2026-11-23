// Package store is a stand-in for a user's GORM code under test.
package store

import "gorm.io/gorm"

type Note struct {
	ID   uint `gorm:"primaryKey"`
	Text string
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Note{})
}

func Add(db *gorm.DB, text string) (uint, error) {
	n := Note{Text: text}
	err := db.Create(&n).Error
	return n.ID, err
}

func Get(db *gorm.DB, id uint) (Note, error) {
	var n Note
	err := db.First(&n, id).Error
	return n, err
}
