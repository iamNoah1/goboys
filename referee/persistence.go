package main

import (
	"github.com/jinzhu/gorm"
)

type Cowboy struct {
	ID     uint   `gorm:"primary_key"`
	Name   string `json:"name"`
	Health int    `json:"health"`
	Damage int    `json:"damage"`
}

func (c *Cowboy) Create(db *gorm.DB) error {
	if err := db.Create(c).Error; err != nil {
		return err
	}
	return nil
}

func (c *Cowboy) Update(db *gorm.DB) error {
	if err := db.Model(&Cowboy{}).Where("name = ?", c.Name).Updates(c).Error; err != nil {
		return err
	}
	return nil
}

func (c *Cowboy) Delete(db *gorm.DB) error {
	if err := db.Where("name = ?", c.Name).Delete(&Cowboy{}).Error; err != nil {
		return err
	}
	return nil
}

func ClearCowboys(db *gorm.DB) error {
	if err := db.Delete(&Cowboy{}, "").Error; err != nil {
		return err
	}
	return nil
}

func GetAllCowboys(db *gorm.DB) ([]Cowboy, error) {
	var cowboys []Cowboy
	if err := db.Find(&cowboys).Error; err != nil {
		return nil, err
	}
	return cowboys, nil
}
