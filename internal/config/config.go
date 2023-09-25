package config

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"
)

type Config struct {
	AppPort       uint `env:"APP_PORT"`
	AppHost       string
	DbType        string `env:"DB_TYPE"`
	DbHost        string `env:"DB_HOST"`
	DbPort        string `env:"DB_PORT"`
	DbName        string `env:"DB_NAME"`
	DbUser        string `env:"DB_USER"`
	DbPassword    string `env:"DB_PASSWORD"`
	IsDev         bool   `env:"DEV_MODE" default:"false"`
	Proxy         *ProxyConfig
	Bank          *BankConfig
	PaymentMethod *PaymentMethodConfig
}

var config = &Config{}
var configOnce = sync.Once{}

var configPaths = []string{"", "./configs/", "../configs/"}

const CreateLinkTimeout = 30 * time.Second
const CheckLinkInterval = 20 * time.Second
const CheckLinkTimeout = 10 * time.Minute
const CheckLinkMaxAttempts = 100

const ModelSubscriptionLifetime = 30 * time.Second
const SubscriptionGCInterval = 1 * time.Minute

const TaskQueueInitialTPI = 300
const TaskQueueIterationDelay = 100 * time.Millisecond

const TaskReloadCardsInterval = 1 * time.Minute // 10 * time.Minute

// для сортировки карт необходимо сравнивать 2 поля models.Card
const CardSortPaymentSumKoef = 5
const CardSortPaymentCountKoef = 2

func GetConfig() *Config {
	configOnce.Do(func() {
		envPath := loadEnvFiles()
		if envPath != "" {
			fmt.Println("using env file: " + envPath)
		}

		conf, err := buildConfigFromEnv()
		if err != nil {
			log.Fatal(err)
		}

		if conf.IsDev {
			conf.AppHost = "http://127.0.0.1:" + strconv.Itoa(int(conf.AppPort))
		} else {
			conf.AppHost = "https://payment-ae.ru/"
		}

		proxyConf, err := buildProxyConfig()
		if err != nil {
			log.Fatal(err)
		}
		conf.Proxy = proxyConf

		bankConf, err := buildBankConfig()
		if err != nil {
			log.Fatal(err)
		}
		conf.Bank = bankConf

		conf.PaymentMethod = GetPaymentMethodConfig()
		conf.Withdraw = BuildWithdrawConfig()

		config = conf
	})
	return config
}

func BankLinks() Links {
	return GetConfig().Bank.Links
}

func buildConfigFromEnv() (*Config, error) {
	conf := &Config{}

	r := reflect.ValueOf(conf).Elem()
	t := r.Type()

	size := r.NumField()
	for i := 0; i < size; i++ {
		tag := t.Field(i).Tag
		envName := tag.Get("env")
		defaultValue := tag.Get("default")
		if len(envName) == 0 {
			continue
		}

		strVal := getEnvOrDefault(envName, defaultValue)

		field := r.Field(i)
		fieldType := field.Type().String()

		if fieldType == "string" {
			field.SetString(strVal)
		} else if fieldType == "int" {
			val, err := strconv.Atoi(strVal)
			if err != nil {
				return nil, err
			}

			field.SetInt(int64(val))
		} else if fieldType == "uint" {
			val, err := strconv.ParseUint(strVal, 10, 32)
			if err != nil {
				return nil, err
			}

			field.SetUint(val)
		} else if fieldType == "bool" {
			field.SetBool(strVal == "1" || strVal == "true")
		}
	}

	return conf, nil
}

// loadEnvFiles Возвращает путь к файлу. Если файл не найден, то вернёт пустую строку.
func loadEnvFiles() string {
	envFiles := []string{"local.env", ".env", "dev.env"}
	for _, root := range configPaths {
		for _, name := range envFiles {
			path := root + name
			if godotenv.Load(path) == nil {
				return path
			}
		}
	}
	return ""
}

func getEnv(name string) (string, error) {
	value, exists := os.LookupEnv(name)
	if !exists {
		return "", fmt.Errorf("env variable with name \"%s\" is not exist", name)
	}
	return value, nil
}

func getEnvOrDefault(name string, defaultValue string) string {
	val, err := getEnv(name)
	if err != nil {
		return defaultValue
	}
	return val
}

func readJSONConfig(fileName string, conf any) error {

	// ищем нужный файл и пытаемся открыть
	for _, path := range configPaths {

		data, err := os.ReadFile(path + fileName)
		if err != nil {
			if os.IsNotExist(err) { // файла нету - продолжаем поиск
				continue
			} else { // если файл есть, но не открылся - ошибка
				return err
			}
		}

		err = json.Unmarshal(data, conf)
		if err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("config file is not found - " + fileName)
}

func validateURL(str string) bool {
	if str == "" {
		return false
	}
	_, err := url.ParseRequestURI(str)
	return err == nil
}

func validatePhone(prefix string, number string) error {
	if len(prefix) != 5 {
		return fmt.Errorf("phone prefix must be 5 digits, got: " + prefix)
	} else if len(number) != 7 {
		return fmt.Errorf("phone number must be 7 digits, got: " + number)
	}
	return nil
}
