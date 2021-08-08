package controllers

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"zakat/api/auth"
	"zakat/api/models"
	"zakat/api/utils/formaterror"

	"github.com/gin-gonic/gin"
)

func (s *Server) CheckZakatFitrah(c *gin.Context) {
	total_person, _ := strconv.Atoi(c.PostForm("total_person"))

	total_weight := models.Rice_weight * float64(total_person)

	c.JSON(http.StatusOK, gin.H{
		"status":       http.StatusOK,
		"total_weight": math.Ceil(total_weight*100) / 100,
	})
}

func (s *Server) CreateZakatFitrah(c *gin.Context) {
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

	zf := models.ZakatFitrah{}
	err = json.Unmarshal(body, &zf)
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
		errList["Unauthorize"] = "Unauthorize"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
	}

	zf.Prepare(tokenUID)
	errMsg := zf.Validate()
	if len(errMsg) > 0 {
		errList = errMsg
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	data, err := zf.SaveZakatFitrah(s.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		errList = formattedError
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": http.StatusCreated,
		"response": gin.H{
			"muzakki_id":   data.IdMuzakki,
			"total_person": data.TotalPerson,
			"total_weight": data.TotalWeight,
			"total_price":  data.TotalPrice,
		},
	})
}

func (s *Server) GetZakatFitrahs(c *gin.Context) {
	errList = map[string]string{}

	zf := models.ZakatFitrah{}

	data, err := zf.GetZakatFitrahs(s.DB)
	if err != nil {
		errList["No_data"] = "No data zakat fitrah"
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

func (s *Server) GetZakatFitrah(c *gin.Context) {
	errList = map[string]string{}

	mID := c.Param("uid")

	tokenUID, err := auth.ExtractTokenUID(c.Request)
	if err != nil {
		errList["Unauthorize"] = "Unauthorize"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
	}

	if mID != tokenUID {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	zf := models.ZakatFitrah{}
	data, err := zf.GetZakatFitrah(mID, s.DB)
	if err != nil {
		errList["No_data"] = "No data zakat fitrah"
		c.JSON(http.StatusNotFound, gin.H{
			"status": http.StatusNotFound,
			"error":  errList,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"response": gin.H{
			"muzakki_id":   data.IdMuzakki,
			"total_person": data.TotalPerson,
			"total_weight": data.TotalWeight,
			"total_price":  data.TotalPrice,
		},
	})
}

func (s *Server) UpdateZakatFitrah(c *gin.Context) {
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

	oriZF := models.ZakatFitrah{}
	err = s.DB.Debug().Model(&models.ZakatFitrah{}).Where("id_muzakki = ?", mID).Take(&oriZF).Error
	if err != nil {
		errList["No_data"] = "No data zakat fitrah"
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

	zf := models.ZakatFitrah{}
	err = json.Unmarshal(body, &zf)
	if err != nil {
		errList["Unmarshal_error"] = "Cannot unmarshal body"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	zf.ID = oriZF.ID

	zf.Prepare(mID)
	errMsg := zf.Validate()
	if len(errMsg) > 0 {
		errList = errMsg
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	data, err := zf.UpdateZakatFitrah(s.DB)
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
			"muzakki_id":   data.IdMuzakki,
			"total_person": data.TotalPerson,
			"total_weight": data.TotalWeight,
			"total_price":  data.TotalPrice,
		},
	})
}

func (s *Server) DeleteZakatFitrah(c *gin.Context) {
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

	zf := models.ZakatFitrah{}
	_, err = zf.DeleteZakatFitrah(s.DB, mID)
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
		"response": "Zakat Fitrah deleted",
	})
}
