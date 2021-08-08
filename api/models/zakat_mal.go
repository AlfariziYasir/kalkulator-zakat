package models

import (
	"errors"
	"html"
	"strings"

	"gorm.io/gorm"
)

type ZakatMal struct {
	gorm.Model
	IdMuzakki   string  `gorm:"column:id_muzakki;not null"`
	TypeZakat   string  `gorm:"size:255;not null" json:"type_zakat"`
	TotalWeight float64 `gorm:"not null;default:0" json:"total_weight"`
	TotalAssest int     `gorm:"not null;default:0" json:"total_price"`
	TotalZakat  int     `gorm:"not null"`
}

func (zm *ZakatMal) Prepare(mID string, totalZakat float64) {
	zm.IdMuzakki = mID
	zm.TypeZakat = html.EscapeString(strings.TrimSpace(strings.ToLower(zm.TypeZakat)))
	zm.TotalZakat = int(totalZakat)
}

func (zm *ZakatMal) Validate() map[string]string {
	var errMsg = make(map[string]string)
	var err error

	if zm.TypeZakat == "" && zm.TypeZakat != "emas" {
		err = errors.New("required type zakat and fill in this columns with emas, perak, or dagang")
		errMsg["Required_type"] = err.Error()
	} else if zm.TypeZakat == "" && zm.TypeZakat != "perak" {
		err = errors.New("required type zakat and fill in this columns with emas, perak, or dagang")
		errMsg["Required_type"] = err.Error()
	} else if zm.TypeZakat == "" && zm.TypeZakat != "dagang" {
		err = errors.New("required type zakat and fill in this columns with emas, perak, or dagang")
		errMsg["Required_type"] = err.Error()
	}
	if zm.TotalWeight == 0 && zm.TotalAssest == 0 {
		err = errors.New("required total weight or total assest")
		errMsg["Required_value"] = err.Error()
	}

	return errMsg
}

func (zm *ZakatMal) SaveZakatMal(db *gorm.DB) (*ZakatMal, error) {
	err := db.Debug().Create(&zm).Error
	if err != nil {
		return &ZakatMal{}, err
	}

	return zm, nil
}

func (zm *ZakatMal) GetZakatMals(db *gorm.DB) (*[]ZakatMal, error) {
	zakatMal := []ZakatMal{}

	err := db.Debug().Model(&ZakatMal{}).Find(&zakatMal).Error
	if err != nil {
		return &[]ZakatMal{}, err
	}

	return &zakatMal, nil
}

func (zm *ZakatMal) GetZakatMalByID(db *gorm.DB, mID string) (*[]ZakatMal, error) {
	zakatMal := []ZakatMal{}
	err := db.Debug().Model(&ZakatMal{}).Where("id_muzakki = ?", mID).Find(&zakatMal).Error
	if err != nil {
		return &[]ZakatMal{}, err
	}

	return &zakatMal, nil
}

func (zm *ZakatMal) GetZakatMalByType(db *gorm.DB, mID, tz string) (*ZakatMal, error) {
	err := db.Debug().Model(&ZakatMal{}).Where("id_muzakki = ? AND type_zakat = ?", mID, tz).First(&zm).Error
	errors.Is(err, gorm.ErrRecordNotFound)

	return zm, nil
}

func (zm *ZakatMal) UpdateZakatMal(db *gorm.DB, tz string) (*ZakatMal, error) {
	err := db.Debug().Model(&ZakatMal{}).Where("id = ? AND type_zakat = ?", zm.ID, tz).Updates(ZakatMal{
		TotalWeight: zm.TotalWeight,
		TotalAssest: zm.TotalAssest,
		TotalZakat:  zm.TotalZakat,
	}).Error
	if err != nil {
		return &ZakatMal{}, err
	}

	err = db.Debug().Model(&ZakatMal{}).Where("id = ?", zm.ID).Take(&zm).Error
	if err != nil {
		return &ZakatMal{}, err
	}

	return zm, nil
}

func (zm *ZakatMal) DeleteZakatMalByID(mID string, db *gorm.DB) (int, error) {
	db = db.Debug().Model(&ZakatMal{}).Where("id_muzakki = ?", mID).Take(&ZakatMal{}).Delete(&ZakatMal{})
	if db.Error != nil {
		return 0, db.Error
	}
	return int(db.RowsAffected), nil
}

func (zm *ZakatMal) DeleteZakatMalByType(mID, tz string, db *gorm.DB) (int, error) {
	db = db.Debug().Model(&ZakatMal{}).Where("id_muzakki = ? AND type_zakat = ?", mID, tz).Take(&ZakatMal{}).Delete(&ZakatMal{})
	if db.Error != nil {
		return 0, db.Error
	}
	return int(db.RowsAffected), nil
}
