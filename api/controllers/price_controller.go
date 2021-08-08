package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"zakat/api/models"

	"github.com/gin-gonic/gin"
)

type PriceTag struct {
	Date  string             `json:"date"`
	Type  string             `json:"base"`
	Price map[string]float64 `json:"rates"`
}

func GetPriceIDR(metal string) (*models.PriceIdr, error) {
	publicAPI := fmt.Sprintf("https://metals-api.com/api/latest?access_key=0e90u74y09tpypx491yed9tdvknyc8t0i33wc291ephx49sm83gt0abgr678&base=%s&symbols=USD,IDR", metal)

	client := &http.Client{}
	req, err := http.NewRequest("GET", publicAPI, nil)
	if err != nil {
		fmt.Print(err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}

	var price PriceTag
	json.Unmarshal(bodyBytes, &price)

	var idr float64
	if len(price.Price) > 0 {
		var ounce float64
		for key, value := range price.Price {
			if key == "USD" {
				break
			}
			ounce = math.Ceil(value*100) / 100
			idr = ounce / 28.35
		}
	}

	newPrice := models.PriceIdr{
		Date: price.Date,
		Type: price.Type,
		Idr:  math.Ceil(idr*100) / 100,
	}

	return &newPrice, nil
}

func (s *Server) CreatePriceIDR(metal string) {

	// metal := c.Param("metal")
	metal = strings.ToLower(metal)
	if metal == "emas" {
		metal = "XAU"
	}
	if metal == "perak" {
		metal = "XAG"
	}

	price, _ := GetPriceIDR(metal)

	err := s.DB.Debug().Model(&models.PriceIdr{}).Create(&price).Error
	if err != nil {
		// errList["Check_failed"] = "IDR price check failed"
		// c.JSON(http.StatusInternalServerError, gin.H{
		// 	"status": http.StatusInternalServerError,
		// 	"error":  errList,
		// })
		fmt.Printf(metal + " check failed")
		return
	}

	// c.JSON(http.StatusOK, gin.H{
	// 	"status": http.StatusOK,
	// 	"response": gin.H{
	// 		"price": price.Idr,
	// 		"type":  price.Type,
	// 		"date":  price.Date,
	// 	},
	// })
	fmt.Printf(metal + " check success")
}

func (s *Server) UpdatePriceIDR(c *gin.Context) {
	metal := c.Param("metal")
	metal = strings.ToLower(metal)
	if metal == "emas" {
		metal = "XAU"
	}
	if metal == "perak" {
		metal = "XAG"
	}

	price, _ := GetPriceIDR(metal)

	err := s.DB.Debug().Model(&models.PriceIdr{}).Where("type = ?", metal).Updates(models.PriceIdr{
		Date: price.Date,
		Type: price.Type,
		Idr:  price.Idr,
	}).Error
	if err != nil {
		errList["Update_failed"] = "Update IDR price failed"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"response": gin.H{
			"price": price.Idr,
			"type":  price.Type,
			"date":  price.Date,
		},
	})
}
