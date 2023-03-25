package main

import (
	"github.com/jinzhu/gorm"
)

type Cowboy struct {
	gorm.Model
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
	if err := db.Save(c).Error; err != nil {
		return err
	}
	return nil
}

func (c *Cowboy) Delete(db *gorm.DB) error {
	if err := db.Delete(c).Error; err != nil {
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

func GetCowboyById(db *gorm.DB, id uint) (*Cowboy, error) {
	var cowboy Cowboy
	if err := db.First(&cowboy, id).Error; err != nil {
		return nil, err
	}
	return &cowboy, nil
}
