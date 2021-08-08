package models

import (
	"math"
	"strings"

	"gorm.io/gorm"
)

type PriceIdr struct {
	gorm.Model
	Date string
	Type string
	Idr  float64
}

type Nisab struct {
	GetNisab float64
	IdrPrice float64
}

func (idr *PriceIdr) GetIDR(metal string, db *gorm.DB) (*Nisab, error) {
	var logam string

	metal = strings.ToLower(metal)
	if metal == "emas" || metal == "dagang" {
		logam = "XAU"
	}
	if metal == "perak" {
		logam = "XAG"
	}

	err := db.Debug().Model(&PriceIdr{}).Where("type = ?", logam).Take(idr).Error
	if err != nil {
		return &Nisab{}, err
	}

	if idr.Type == "XAU" {
		getNisab := float64(80) * idr.Idr
		getNisab = math.Ceil(getNisab*100) / 100
		result := Nisab{
			GetNisab: getNisab,
			IdrPrice: idr.Idr,
		}

		return &result, nil
	}
	if idr.Type == "XAG" {
		getNisab := float64(543) * idr.Idr
		getNisab = math.Ceil(getNisab*100) / 100
		result := Nisab{
			GetNisab: getNisab,
			IdrPrice: idr.Idr,
		}

		return &result, nil
	}

	return &Nisab{}, err
}
