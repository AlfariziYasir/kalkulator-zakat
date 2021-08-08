package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"zakat/api/auth"
	"zakat/api/models"
	"zakat/api/utils/formaterror"

	"github.com/gin-gonic/gin"
)

func (s *Server) CheckZakatMal(c *gin.Context) {
	errList = map[string]string{}

	var idr models.PriceIdr
	metal := c.PostForm("metal")
	total_harta, _ := strconv.ParseFloat(c.PostForm("assest"), 64)
	total_wegiht, _ := strconv.ParseFloat(c.PostForm("weight"), 64)

	getIdr, err := idr.GetIDR(metal, s.DB)
	if err != nil {
		errList["Get_fail"] = "failed to get IDR price"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	var price_weight float64
	if metal == "emas" || metal == "perak" {
		price_weight = total_wegiht * getIdr.IdrPrice
	}

	var pay_zakat float64
	if total_harta > getIdr.GetNisab && metal == "dagang" {
		pay_zakat = (total_harta * 2.5) / 100

		c.JSON(http.StatusOK, map[string]interface{}{
			"message":         "check zakat dagang success",
			"total_zakat_mal": pay_zakat,
		})
		return
	} else if price_weight > getIdr.GetNisab && metal == "emas" {
		pay_zakat = (price_weight * 2.5) / 100

		c.JSON(http.StatusOK, map[string]interface{}{
			"message":         "check zakat emas success",
			"total_zakat_mal": pay_zakat,
		})
		return
	} else if price_weight > getIdr.GetNisab && metal == "perak" {
		pay_zakat = (price_weight * 2.5) / 100

		c.JSON(http.StatusOK, map[string]interface{}{
			"message":         "check zakat perak success",
			"total_zakat_mal": pay_zakat,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, map[string]interface{}{
		"message": "tidak wajib membayar zakat",
	})
}

func (s *Server) CreateZakatMal(c *gin.Context) {
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

	zm := models.ZakatMal{}
	err = json.Unmarshal(body, &zm)
	if err != nil {
		errList["Unmarshal_error"] = "Cannot unmarshal body"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	idr := models.PriceIdr{}
	getIdr, err := idr.GetIDR(zm.TypeZakat, s.DB)
	if err != nil {
		errList["Get_fail"] = "failed to get IDR price"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	var pw float64
	if strings.ToLower(zm.TypeZakat) == "emas" || strings.ToLower(zm.TypeZakat) == "perak" {
		pw = zm.TotalWeight * getIdr.IdrPrice
	}

	var pay float64
	if float64(zm.TotalAssest) > getIdr.GetNisab && strings.ToLower(zm.TypeZakat) == "dagang" {
		pay = (float64(zm.TotalAssest) * 2.5) / 100
	} else if pw > getIdr.GetNisab && strings.ToLower(zm.TypeZakat) == "emas" {
		pay = (pw * 2.5) / 100
	} else if pw > getIdr.GetNisab && strings.ToLower(zm.TypeZakat) == "perak" {
		pay = (pw * 2.5) / 100
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "tidak wajib membayar zakat",
		})
		return
	}

	tokenUID, err := auth.ExtractTokenUID(c.Request)
	if err != nil {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	zm.Prepare(tokenUID, pay)
	errMsg := zm.Validate()
	if len(errMsg) > 0 {
		errList = errMsg
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	data, err := zm.SaveZakatMal(s.DB)
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
			"id_muzakki":   data.IdMuzakki,
			"type_zakat":   data.TypeZakat,
			"total_weight": data.TotalWeight,
			"total_assest": data.TotalAssest,
			"total_zakat":  data.TotalZakat,
		},
	})
}

func (s *Server) GetZakatMals(c *gin.Context) {
	errList = map[string]string{}

	zm := models.ZakatMal{}
	data, err := zm.GetZakatMals(s.DB)
	if err != nil {
		errList["No_data"] = "No data zakat mal"
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

func (s *Server) GetZakatMalByID(c *gin.Context) {
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

	zm := models.ZakatMal{}
	data, err := zm.GetZakatMalByID(s.DB, mID)
	if err != nil {
		errList["No_data"] = "No data zakat mal"
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

func (s *Server) GetZakatMalByType(c *gin.Context) {
	errList = map[string]string{}

	typeZakat := c.Param("type")
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

	zm := models.ZakatMal{}
	data, err := zm.GetZakatMalByType(s.DB, mID, typeZakat)
	if err != nil {
		errList["No_data"] = "No data zakat mal"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"response": gin.H{
			"id_muzakki":   data.IdMuzakki,
			"type_zakat":   data.TypeZakat,
			"total_weight": data.TotalWeight,
			"total_assest": data.TotalAssest,
			"total_zakat":  data.TotalZakat,
		},
	})
}

func (s *Server) UpdateZakatMal(c *gin.Context) {
	errList = map[string]string{}

	typeZakat := c.Param("type")
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

	oriZM := models.ZakatMal{}
	err = s.DB.Debug().Model(&models.ZakatMal{}).Where("id_muzakki = ? AND type_zakat = ?", mID, typeZakat).Take(&oriZM).Error
	if err != nil {
		errList["No_data"] = "No data zakat mal"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
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

	zm := models.ZakatMal{}
	err = json.Unmarshal(body, &zm)
	if err != nil {
		errList["Unmarshal_error"] = "Cannot unmarshal body"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	idr := models.PriceIdr{}
	getIdr, err := idr.GetIDR(zm.TypeZakat, s.DB)
	if err != nil {
		errList["Get_fail"] = "failed to get IDR price"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	var pw float64
	if strings.ToLower(zm.TypeZakat) == "emas" || strings.ToLower(zm.TypeZakat) == "perak" {
		pw = zm.TotalWeight * getIdr.IdrPrice
	}

	var pay float64
	if float64(zm.TotalAssest) > getIdr.GetNisab && strings.ToLower(zm.TypeZakat) == "dagang" {
		pay = (float64(zm.TotalAssest) * 2.5) / 100
	} else if pw > getIdr.GetNisab && strings.ToLower(zm.TypeZakat) == "emas" {
		pay = (pw * 2.5) / 100
	} else if pw > getIdr.GetNisab && strings.ToLower(zm.TypeZakat) == "perak" {
		pay = (pw * 2.5) / 100
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "tidak wajib membayar zakat",
		})
		return
	}

	zm.ID = oriZM.ID
	zm.Prepare(mID, pay)
	errMsg := zm.Validate()
	if len(errMsg) > 0 {
		errList = errMsg
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	data, err := zm.UpdateZakatMal(s.DB, zm.TypeZakat)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		errList = formattedError
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"response": gin.H{
			"id_muzakki":   data.IdMuzakki,
			"type_zakat":   data.TypeZakat,
			"total_weight": data.TotalWeight,
			"total_assest": data.TotalAssest,
			"total_zakat":  data.TotalZakat,
		},
	})

}

func (s *Server) DeleteZakatMalByID(c *gin.Context) {
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

	zm := models.ZakatMal{}

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
		"response": "Zakat Mal deleted",
	})
}

func (s *Server) DeleteZakatMalByType(c *gin.Context) {
	errList = map[string]string{}

	typeZakat := c.Param("type")
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

	zm := models.ZakatMal{}
	_, err = zm.DeleteZakatMalByType(mID, typeZakat, s.DB)
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
		"response": "Zakat Mal " + zm.TypeZakat + " deleted",
	})
}
