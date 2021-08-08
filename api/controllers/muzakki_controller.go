package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"zakat/api/auth"
	"zakat/api/models"
	"zakat/api/utils/formaterror"

	"github.com/gin-gonic/gin"
)

func (s *Server) CreateMuzakki(c *gin.Context) {
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

	m := models.Muzakki{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		errList["Unmarshal_error"] = "Cannot unmarshal body"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	tokenUID, err := auth.ExtractTokenUID(c.Request)
	if err != nil {
		errList["Invalid_request"] = "Invalid request"
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"error":  errList,
		})
		return
	}

	m.Prepare(tokenUID)
	errMsg := m.Validate()
	if len(errMsg) > 0 {
		errList = errMsg
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
	}

	data, err := m.SaveMuzakki(s.DB)
	if err != nil {
		errList := formaterror.FormatError(err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": http.StatusCreated,
		"reponse": gin.H{
			"muzakki_id": data.MuzakkiId,
			"name":       data.Name,
			"mobile":     data.Mobile,
			"address":    data.Address,
		},
	})
}

func (s *Server) GetMuzakkis(c *gin.Context) {
	errList := map[string]string{}
	m := models.Muzakki{}

	data, err := m.GetMuzakkis(s.DB)
	if err != nil {
		errList["No_data"] = "No data muzakki"
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

func (s *Server) GetMuzakki(c *gin.Context) {
	errList := map[string]string{}

	mID := c.Param("uid")

	tokenUID, err := auth.ExtractTokenUID(c.Request)
	if err != nil {
		errList["Unauthorized"] = "Unauthorized"
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	if mID != tokenUID {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	m := models.Muzakki{}
	data, err := m.GetMuzakki(s.DB, mID)
	if err != nil {
		errList["No_data"] = "No data muzakki"
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

func (s *Server) UpdateMuzakki(c *gin.Context) {
	errList = map[string]string{}

	mID := c.Param("uid")

	tokenUID, err := auth.ExtractTokenUID(c.Request)
	if err != nil {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	if mID != tokenUID {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	oriMuzakki := models.Muzakki{}
	err = s.DB.Debug().Model(&models.Muzakki{}).Where("muzakki_id = ?", mID).Take(&oriMuzakki).Error
	if err != nil {
		errList["No_data"] = "No data muzakki"
		c.JSON(http.StatusNotFound, gin.H{
			"status": http.StatusNotFound,
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

	m := models.Muzakki{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		errList["Unmarshal_error"] = "Cannot unmarshal body"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	m.ID = oriMuzakki.ID

	m.Prepare(mID)
	errMsg := m.Validate()
	if len(errMsg) > 0 {
		errList = errMsg
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	data, err := m.UpdateMuzakki(s.DB)
	if err != nil {
		errList := formaterror.FormatError(err.Error())
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

func (s *Server) DeleteMuzakki(c *gin.Context) {
	errList = map[string]string{}

	mID := c.Param("uid")

	tokenUID, err := auth.ExtractTokenUID(c.Request)
	if err != nil {
		errList["Unauthorize"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	if mID != tokenUID {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	m := models.Muzakki{}
	zf := models.ZakatFitrah{}
	zm := models.ZakatMal{}

	_, err = m.DeleteMuzakki(s.DB, mID)
	if err != nil {
		errList["Other_error"] = "Please try again later"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	_, err = zf.DeleteZakatFitrah(s.DB, mID)
	if err != nil {
		errList["Other_error"] = "Please try again later"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	_, err = zm.DeleteZakatMalByID(mID, s.DB)
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
		"response": "Muzakki deleted",
	})
}
