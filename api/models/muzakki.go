package models

import (
	"errors"
	"html"
	"strings"

	"gorm.io/gorm"
)

type Muzakki struct {
	gorm.Model
	MuzakkiId    string      `gorm:"not null;unique"`
	Name         string      `gorm:"size:255;not null" json:"name"`
	Mobile       string      `gorm:"size:255;not null" json:"mobile"`
	Address      string      `gorm:"size:255;not null" json:"address"`
	ZakatFitrahs ZakatFitrah `gorm:"foreignKey:IdMuzakki;references:MuzakkiId"`
	ZakatMals    []ZakatMal  `gorm:"foreignKey:IdMuzakki;references:MuzakkiId"`
}

func (m *Muzakki) Prepare(uid string) {
	m.MuzakkiId = uid
	m.Name = html.EscapeString(strings.TrimSpace(m.Name))
	m.Address = html.EscapeString(strings.TrimSpace(m.Address))
	m.Mobile = html.EscapeString(strings.TrimSpace(m.Mobile))
	m.ZakatFitrahs = ZakatFitrah{}
	m.ZakatMals = []ZakatMal{}
}

func (m *Muzakki) Validate() map[string]string {
	var errMsg = make(map[string]string)
	var err error

	if m.Name == "" {
		err = errors.New("required name")
		errMsg["Required_name"] = err.Error()
	}
	if m.Address == "" {
		err = errors.New("required address")
		errMsg["Required_address"] = err.Error()
	}
	if m.Mobile == "" && len(m.Mobile) < 10 {
		err = errors.New("required mobile")
		errMsg["Required_mobile"] = err.Error()
	}

	return errMsg
}

func (m *Muzakki) SaveMuzakki(db *gorm.DB) (*Muzakki, error) {
	err := db.Debug().Create(&m).Error
	if err != nil {
		return &Muzakki{}, err
	}

	return m, nil
}

func (m *Muzakki) GetMuzakkis(db *gorm.DB) (*[]Muzakki, error) {
	muzakki := []Muzakki{}
	err := db.Debug().Preload("ZakatFitrahs").Preload("ZakatMals").Find(&muzakki).Error
	if err != nil {
		return &[]Muzakki{}, err
	}

	return &muzakki, err
}

func (m *Muzakki) GetMuzakki(db *gorm.DB, mID string) (*Muzakki, error) {
	err := db.Debug().Preload("ZakatFitrahs").Preload("ZakatMals").Where("muzakki_id = ?", mID).Find(&m).Error
	errors.Is(err, gorm.ErrRecordNotFound)

	return m, err
}

func (m *Muzakki) UpdateMuzakki(db *gorm.DB) (*Muzakki, error) {
	db = db.Debug().Model(&Muzakki{}).Where("id = ?", m.ID).Take(&Muzakki{}).UpdateColumns(
		map[string]interface{}{
			"name":    m.Name,
			"mobile":  m.Mobile,
			"address": m.Address,
		},
	)
	if db.Error != nil {
		return &Muzakki{}, db.Error
	}

	err := db.Debug().Preload("ZakatFitrahs").Preload("ZakatMals").Where("id = ?", m.ID).Find(&m).Error
	if err != nil {
		return &Muzakki{}, err
	}

	return m, nil
}

func (m *Muzakki) DeleteMuzakki(db *gorm.DB, uID string) (int, error) {
	db = db.Debug().Model(&Muzakki{}).Where("id = ?", uID).Take(&Muzakki{}).Delete(&Muzakki{})
	if db.Error != nil {
		return 0, db.Error
	}
	return int(db.RowsAffected), nil
}
