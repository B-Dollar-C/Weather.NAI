package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type apiConfigData struct {
	OpenWeatherMapApiKey string `json: "OpenWeatherMapApiKey"`
}

type weatherData struct {
	Name string `json: "name"`
	Main struct {
		Temp    float64 `json: "temp"`
		Celsius string  `json: "celsius"`

		Pressure int `json: "pressure"`
		Humidity int `json: "humidity"`
	} `json: "main"`
	Coord struct {
		Lat float64 `json: "lat"`
		Lon float64 `json: "long"`
	} `json: "coord"`
	Weather []Weath `json: "weather"`
	Wind    struct {
		Speed float64 `json: "speed"`
		Deg   float64 `json: "deg"`
	}
	Sys struct {
		Country     string `json: "country"`
		Sunrise     int    `json: "sunrise"`
		SunriseTime string `json: "sunriseTime"`
		Sunset      int    `json: "sunset"`
		SunsetTime  string `json: "sunsetTime"`
	}
	Timezone int `json: "timezone"`
}

type Weath struct {
	Description string `json: "description"`
}

func loadApiConfig(filename string) (apiConfigData, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return apiConfigData{}, err
	}

	var c apiConfigData
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return apiConfigData{}, err
	}
	return c, nil
}

func query(city string) (weatherData, error) {
	apiConfig, err := loadApiConfig(".apiConfig")
	if err != nil {
		return weatherData{}, nil
	}
	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=" + apiConfig.OpenWeatherMapApiKey + "&q=" + city)
	fmt.Println(resp)
	if err != nil {
		return weatherData{}, err
	}

	defer resp.Body.Close()

	var d weatherData

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return weatherData{}, err
	}
	d.Main.Celsius = fmt.Sprintf("%.2f", (d.Main.Temp - 273.15))
	d.Sys.SunriseTime = (time.Unix(int64(d.Sys.Sunrise), 0)).Format("03:04 PM")
	d.Sys.SunsetTime = (time.Unix(int64(d.Sys.Sunset), 0)).Format("03:04 PM")
	fmt.Println(d)
	return d, nil

}

func main() {
	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(CORSMiddleware())
	r.GET("/", func(c *gin.Context) {
		indexHtml, err := ioutil.ReadFile("index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHtml)
	})
	r.GET("/weather/:city", func(c *gin.Context) {
		city := c.Param("city")
		data, err := query(city)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		fmt.Println(data)
		c.JSON(http.StatusOK, data)
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Panicf("error: %s", err)
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return

		}

		c.Next()
	}

}
