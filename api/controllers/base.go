package controllers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"zakat/api/middleware"
	"zakat/api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Server struct {
	DB     *gorm.DB
	Router *gin.Engine
}

var errList = make(map[string]string)

func (s *Server) Initialize(DBDriver, DBUser, DBPort, DBHost, DBName, DBPassword string) {
	var err error

	DBURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", DBHost, DBPort, DBUser, DBName, DBPassword)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // Disable color
		},
	)

	s.DB, err = gorm.Open(postgres.Open(DBURL), &gorm.Config{
		Logger:                                   newLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		fmt.Printf("Cannot connect to %s database", DBDriver)
		log.Fatal("This is the error connecting to postgres:", err)
	} else {
		fmt.Printf("We are connected to the %s database", DBDriver)
	}

	// s.DB.Debug().Migrator().DropTable(
	// 	&models.User{},
	// 	&models.Muzakki{},
	// 	&models.ZakatFitrah{},
	// 	&models.ZakatMal{},
	// 	&models.PriceIdr{},
	// )

	s.DB.Debug().AutoMigrate(
		&models.User{},
		&models.Muzakki{},
		&models.ZakatFitrah{},
		&models.ZakatMal{},
		&models.PriceIdr{},
	)

	//get price
	var timer = time.NewTimer(15 * time.Second)
	fmt.Println("get price start")
	s.CreatePriceIDR("emas")
	<-timer.C
	s.CreatePriceIDR("perak")
	fmt.Println("\nget price finish")

	s.Router = gin.Default()
	s.Router.Use(middleware.CORSMiddleware())

	s.InitializeRoutes()
}

func (s *Server) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, s.Router))
}
