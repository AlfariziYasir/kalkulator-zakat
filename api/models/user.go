package models

import (
	"crypto/rand"
	"errors"
	"fmt"
	"html"
	"log"
	"strings"
	"zakat/api/security"

	"github.com/badoux/checkmail"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserId   string `gorm:"not null;unique"`
	Username string `gorm:"size:255;not null;unique" json:"username"`
	Email    string `gorm:"size:255;not null;unique" json:"email"`
	Password string `gorm:"size:255;not null;" json:"password"`
	Role     string `json:"role"`
}

func RandomUID() (s string) {
	b := make([]byte, 2)
	_, err := rand.Read(b)
	if err != nil {
		return
	}
	s = fmt.Sprintf("%x", b)
	return
}

func (u *User) BeforeSave() error {
	hashpassword, err := security.Hash(u.Password)
	if err != nil {
		return err
	}
	u.Password = string(hashpassword)
	return nil
}

func (u *User) Prepare() {
	u.UserId = RandomUID()
	u.Username = html.EscapeString(strings.TrimSpace(u.Username))
	u.Email = html.EscapeString(strings.TrimSpace(u.Email))
	u.Role = html.EscapeString(strings.TrimSpace(u.Role))
}

func (u *User) Validate(action string) map[string]string {
	var errMsg = make(map[string]string)
	var err error

	switch strings.ToLower(action) {
	case "update":
		if u.Username == "" {
			err = errors.New("required username")
			errMsg["Required_username"] = err.Error()
		}
		if u.Email == "" {
			err = errors.New("required email")
			errMsg["Required_email"] = err.Error()
		}
		if u.Email != "" {
			if err = checkmail.ValidateFormat(u.Email); err != nil {
				err = errors.New("invalid Email")
				errMsg["Invalid_email"] = err.Error()
			}
		}
	case "login":
		if u.Password == "" {
			err = errors.New("required password")
			errMsg["Required_password"] = err.Error()
		}
		if u.Email == "" {
			err = errors.New("required email")
			errMsg["Required_email"] = err.Error()
		}
		if u.Email != "" {
			if err = checkmail.ValidateFormat(u.Email); err != nil {
				err = errors.New("invalid Email")
				errMsg["Invalid_email"] = err.Error()
			}
		}
	default:
		if u.Username == "" {
			err = errors.New("required username")
			errMsg["Required_username"] = err.Error()
		}
		if u.Password == "" {
			err = errors.New("required password")
			errMsg["Required_password"] = err.Error()
		}
		if u.Password != "" && len(u.Password) < 6 {
			err = errors.New("password should be atleast 6 characters")
			errMsg["Invalid_password"] = err.Error()
		}
		if u.Email == "" {
			err = errors.New("required email")
			errMsg["Required_email"] = err.Error()
		}
		if u.Email != "" {
			if err = checkmail.ValidateFormat(u.Email); err != nil {
				err = errors.New("invalid Email")
				errMsg["Invalid_email"] = err.Error()
			}
		}
		if u.Role == "" {
			err = errors.New("required role")
			errMsg["Required_role"] = err.Error()
		}
		if u.Role != "admin" && u.Role != "muzakki" {
			err = errors.New("invalid role")
			errMsg["Invalid_role"] = err.Error()
		}
	}
	return errMsg
}

func (u *User) SaveUser(db *gorm.DB) (*User, error) {
	err := db.Debug().Create(&u).Error
	if err != nil {
		return &User{}, err
	}

	return u, err
}

func (u *User) GetUsers(db *gorm.DB) (*[]User, error) {
	users := []User{}

	err := db.Debug().Model(&User{}).Find(&users).Error
	if err != nil {
		return &[]User{}, err
	}

	return &users, nil
}

func (u *User) GetUser(uID string, db *gorm.DB) (*User, error) {
	err := db.Debug().Model(&User{}).Where("user_id = ?", uID).First(&u).Error
	errors.Is(err, gorm.ErrRecordNotFound)

	return u, err
}

func (u *User) UpdateUser(uID string, db *gorm.DB) (*User, error) {
	if u.Password != "" {
		// hash password
		err := u.BeforeSave()
		if err != nil {
			log.Fatal(err)
		}

		db.Debug().Model(&User{}).Where("user_id = ?", uID).Updates(User{
			Password: u.Password,
			Username: u.Username,
			Email:    u.Email,
		})
	}

	db.Debug().Model(&User{}).Where("user_id = ?", uID).Updates(User{
		Username: u.Username,
		Email:    u.Email,
	})

	if db.Error != nil {
		return &User{}, db.Error
	}

	//get updated data by id
	err := db.Debug().Model(&User{}).Where("user_id = ?", uID).Take(&u).Error
	if err != nil {
		return &User{}, err
	}

	return u, nil
}

func (u *User) DeleteUser(uID string, db *gorm.DB) (int64, error) {
	db = db.Debug().Model(&User{}).Where("user_id = ?", uID).Delete(&User{})
	if db.Error != nil {
		return 0, db.Error
	}

	return db.RowsAffected, nil
}
