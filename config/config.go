package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	MongoURI          string `envconfig:"MONGO_URI" default:"mongodb://localhost:27017"`
	MongoDatabaseName string `envconfig:"MONGO_DATABASE_NAME" default:"minhas-rifas"`
	Port              string `envconfig:"PORT" default:"8080"`
}

func New() (cfg Config, err error) {
	err = envconfig.Process("", &cfg)
	return
}
