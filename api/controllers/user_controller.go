package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"zakat/api/auth"
	"zakat/api/models"
	"zakat/api/security"
	"zakat/api/utils/formaterror"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) CreateUser(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		errList = map[string]string{}

		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			errList["Invalid_body"] = "Unable to get request"
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": http.StatusUnprocessableEntity,
				"error":  errList,
			})
			return
		}

		user := models.User{}

		err = json.Unmarshal(body, &user)
		if err != nil {
			errList["Unmarshal_error"] = "Cannot unmarshal body"
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": http.StatusUnprocessableEntity,
				"error":  errList,
			})
			return
		}

		user.Prepare()
		errMsg := user.Validate("")
		if len(errMsg) > 0 {
			errList = errMsg
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": http.StatusUnprocessableEntity,
				"error":  errList,
			})
			return
		}

		data, err := user.SaveUser(s.DB)
		if err != nil {
			formattedError := formaterror.FormatError(err.Error())
			errList = formattedError
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": http.StatusInternalServerError,
				"error":  errList,
			})
			return
		}

		enforcer.AddGroupingPolicy(fmt.Sprintf(user.UserId), user.Role)

		c.JSON(http.StatusCreated, gin.H{
			"status": http.StatusCreated,
			"response": map[string]interface{}{
				"user_id":  data.UserId,
				"username": data.Username,
				"email":    data.Email,
				"role":     user.Role,
			},
		})
	}

}

func (s *Server) GetUsers(c *gin.Context) {
	errList = map[string]string{}

	user := models.User{}

	data, err := user.GetUsers(s.DB)
	if err != nil {
		errList["No_user"] = "No User Found"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": data,
	})
}

func (server *Server) GetUser(c *gin.Context) {
	errList = map[string]string{}

	userID := c.Param("uid")

	tokenUID, err := auth.ExtractTokenUID(c.Request)
	if err != nil {
		errList["Invalid_request"] = "Invalid request"
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"error":  errList,
		})
		return
	}

	if userID != tokenUID {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	user := models.User{}
	data, err := user.GetUser(userID, server.DB)
	if err != nil {
		errList["No_user"] = "No user found"
		c.JSON(http.StatusNotFound, gin.H{
			"status": http.StatusNotFound,
			"error":  errList,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": data,
	})
}

func (s *Server) UpdateUser(c *gin.Context) {
	errList = map[string]string{}

	userID := c.Param("uid")

	tokenUID, err := auth.ExtractTokenUID(c.Request)
	if err != nil {
		errList["Invalid_request"] = "Invalid request"
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"error":  errList,
		})
		return
	}

	if userID != tokenUID {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errList["Invalid_body"] = "Unable to get request"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	reqBody := map[string]string{}
	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		errList["Unmarshal_error"] = "Cannot unmarshal body"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	formerUser := models.User{}
	err = s.DB.Debug().Model(&models.User{}).Where("user_id = ?", userID).Take(&formerUser).Error
	if err != nil {
		errList["User_invalid"] = "The user is does not exist"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	newUser := models.User{}
	if reqBody["current_password"] != "" && reqBody["new_password"] == "" {
		errList["Empty_current"] = "Please Provide current password"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}
	if reqBody["current_password"] != "" && reqBody["new_password"] == "" {
		errList["Empty_new"] = "Please Provide new password"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}
	if reqBody["current_password"] != "" && reqBody["new_password"] == "" {
		if len(reqBody["new_password"]) < 6 {
			errList["Invalid_password"] = "Password should be atleast 6 characters"
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": http.StatusUnprocessableEntity,
				"error":  errList,
			})
			return
		}
		err := security.VerifyPassword(formerUser.Password, reqBody["current_password"])
		if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
			errList["Password_mismatch"] = "The password not correct"
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": http.StatusUnprocessableEntity,
				"error":  errList,
			})
			return
		}

		newUser.Username = reqBody["username"]
		newUser.Email = reqBody["email"]
		newUser.Password = reqBody["new_password"]
	}

	newUser.Username = reqBody["username"]
	newUser.Email = reqBody["email"]

	newUser.Prepare()
	errMsg := newUser.Validate("update")
	if len(errMsg) > 0 {
		errList = errMsg
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	data, err := newUser.UpdateUser(userID, s.DB)
	if err != nil {
		errList := formaterror.FormatError(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"response": gin.H{
			"user_id":  data.UserId,
			"username": data.Username,
			"email":    data.Email,
		},
	})
}

func (s *Server) DeleteUser(c *gin.Context) {
	errList = map[string]string{}

	userID := c.Param("uid")

	tokenUID, err := auth.ExtractTokenUID(c.Request)
	if err != nil {
		errList["Unauthorize"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	if userID != tokenUID {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	user := models.User{}
	_, err = user.DeleteUser(userID, s.DB)
	if err != nil {
		errList["Other_error"] = "Please try again later"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	m := models.Muzakki{}
	zf := models.ZakatFitrah{}
	zm := models.ZakatMal{}

	_, err = m.DeleteMuzakki(s.DB, userID)
	if err != nil {
		errList["Other_error"] = "Please try again later"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	_, err = zf.DeleteZakatFitrah(s.DB, userID)
	if err != nil {
		errList["Other_error"] = "Please try again later"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	_, err = zm.DeleteZakatMalByID(userID, s.DB)
	if err != nil {
		errList["Other_error"] = "Please try again later"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": "User deleted",
	})
}
