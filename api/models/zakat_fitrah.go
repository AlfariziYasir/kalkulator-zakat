package models

import (
	"errors"
	"math"

	"gorm.io/gorm"
)

type ZakatFitrah struct {
	gorm.Model
	IdMuzakki   string  `gorm:"column:id_muzakki;unique;not null"`
	TotalPerson int     `gorm:"not null" json:"totalPerson"`
	TotalWeight float64 `gorm:"not null"`
	TotalPrice  int     `gorm:"not null"`
}

const (
	Rice_weight = 2.8
	Rice_price  = 33000
)

func (zf *ZakatFitrah) Prepare(mID string) {
	//get data
	person := int(zf.TotalPerson)
	weight := (float64(person) * Rice_weight)
	total_weight := math.Ceil(weight*100) / 100
	total_price := person * Rice_price

	zf.IdMuzakki = mID
	zf.TotalPerson = person
	zf.TotalWeight = total_weight
	zf.TotalPrice = total_price
}

func (zf *ZakatFitrah) Validate() map[string]string {
	var errMsg = make(map[string]string)
	var err error

	if zf.TotalPerson < 1 {
		err = errors.New("required total person")
		errMsg["Required_totalPerson"] = err.Error()
	}

	return errMsg
}

func (zf *ZakatFitrah) SaveZakatFitrah(db *gorm.DB) (*ZakatFitrah, error) {
	err := db.Debug().Model(&ZakatFitrah{}).Create(&zf).Error
	if err != nil {
		return &ZakatFitrah{}, err
	}

	return zf, nil
}

func (zf *ZakatFitrah) GetZakatFitrahs(db *gorm.DB) (*[]ZakatFitrah, error) {
	zfs := []ZakatFitrah{}
	err := db.Debug().Model(&ZakatFitrah{}).Find(&zfs).Error
	if err != nil {
		return &[]ZakatFitrah{}, err
	}

	return &zfs, nil
}

func (zf *ZakatFitrah) GetZakatFitrah(mID string, db *gorm.DB) (*ZakatFitrah, error) {
	err := db.Debug().Model(&ZakatFitrah{}).Where("id_muzakki = ?", mID).First(&zf).Error
	if err != nil {
		return &ZakatFitrah{}, err
	}

	return zf, nil
}

func (zf *ZakatFitrah) UpdateZakatFitrah(db *gorm.DB) (*ZakatFitrah, error) {
	err := db.Debug().Model(&ZakatFitrah{}).Where("id = ?", zf.ID).Updates(ZakatFitrah{
		TotalPerson: zf.TotalPerson,
		TotalWeight: zf.TotalWeight,
		TotalPrice:  zf.TotalPrice,
	}).Error
	if err != nil {
		return &ZakatFitrah{}, err
	}

	err = db.Debug().Model(&ZakatFitrah{}).Where("id = ?", zf.ID).Take(&zf).Error
	if err != nil {
		return &ZakatFitrah{}, err
	}

	return zf, nil
}

func (zf *ZakatFitrah) DeleteZakatFitrah(db *gorm.DB, uid string) (int, error) {
	db = db.Debug().Model(&ZakatFitrah{}).Where("id_muzakki = ?", uid).Take(&ZakatFitrah{}).Delete(&ZakatFitrah{})
	if db.Error != nil {
		return 0, db.Error
	}
	return int(db.RowsAffected), nil
}
