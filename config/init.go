package config

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	Params      = viper.New()
	DB          *gorm.DB
	RedisClient *redis.Client
)

func InitDB() error {

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		Params.GetString("db.live.host"),
		Params.GetString("db.live.port"),
		Params.GetString("db.live.user_name"),
		Params.GetString("db.live.password"),
		Params.GetString("db.live.db_name"),
	)

	if Params.GetString("environment") == "test" {
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			Params.GetString("db.test.host"),
			Params.GetString("db.test.port"),
			Params.GetString("db.test.user_name"),
			Params.GetString("db.test.password"),
			Params.GetString("db.test.db_name"),
		)
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return err
	} else {
		fmt.Println("DB connected", dsn)

		pgDB, _ := DB.DB()

		// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
		pgDB.SetMaxIdleConns(600)

		// SetMaxOpenConns sets the maximum number of open connections to the database.
		pgDB.SetMaxOpenConns(2000)

		// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
		pgDB.SetConnMaxLifetime(120 * time.Second)
		pgDB.SetConnMaxIdleTime(120 * time.Second)
	}

	if Params.GetString("environment") == "test" {
		testRedisConf := redis.Options{
			Addr: fmt.Sprintf("%s:%s", Params.GetString("redis.test.address"), Params.GetString("redis.test.port")),
			//Password: Params.GetString("redis.test.password"),
			DB:           Params.GetInt("redis.test.db"),
			DialTimeout:  3 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 5 * time.Second,
		}
		fmt.Printf("Test Redis Config: %+v\n", testRedisConf)
		RedisClient = redis.NewClient(&testRedisConf)
		fmt.Println("connect to redis test")
	} else {
		RedisClient = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("%s:%s", Params.GetString("redis.live.address"), Params.GetString("redis.live.port")),
			//Password: Params.GetString("redis.live.password"),
			DB:           Params.GetInt("redis.live.db"),
			DialTimeout:  3 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 5 * time.Second,
		})
		fmt.Println("connect to redis live")
	}

	return nil
}

func InitConfigFile(configPath string) error {
	Params.AddConfigPath(configPath)
	viper.SetConfigName("config.yaml")
	viper.SetConfigName("yaml")
	if err := Params.ReadInConfig(); err != nil {
		return err
	} else {
		fmt.Println("config file loaded")
	}
	return nil
}
