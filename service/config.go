package service

import (
	"errors"
	"fmt"

	"github.com/hoquangnam45/pharmacy-auth/model"
	"gorm.io/gorm"
)

type Config struct {
	db *gorm.DB
}

func NewConfigService(db *gorm.DB) *Config {
	return &Config{db}
}

var ErrMissingValue = errors.New("missing value")

func (s *Config) GetInNs(namespace *string, key string) (*string, error) {
	config := model.Config{}
	result := s.db.
		Where(&model.Config{
			Namespace: namespace,
			Key:       key,
		}).
		First(&config)

	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	} else {
		return config.Value, nil
	}
}

func (s *Config) GetInNsOrDefault(namespace *string, key string, defaultValue string) string {
	if val, err := s.GetInNs(namespace, key); val == nil || err != nil {
		return defaultValue
	} else {
		return *val
	}
}

func (s *Config) MustHaveInNs(namespace string, key string) (string, error) {
	val, err := s.GetInNs(&namespace, key)
	if err != nil {
		return "", err
	}
	if val == nil {
		return "", fmt.Errorf("%w %s in namespace %s", ErrMissingValue, key, namespace)
	}
	return *val, err
}

func (s *Config) Get(key string) (*string, error) {
	config := model.Config{}
	result := s.db.
		Where(&model.Config{
			Namespace: nil,
			Key:       key,
		}).
		First(&config)

	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	} else {
		return config.Value, nil
	}
}

func (s *Config) GetOrDefault(key string, defaultValue string) string {
	if val, err := s.Get(key); val == nil || err != nil {
		return defaultValue
	} else {
		return *val
	}
}

func (s *Config) MustHave(key string) (string, error) {
	val, err := s.Get(key)
	if err != nil {
		return "", err
	}
	if val == nil {
		return "", fmt.Errorf("%w %s", ErrMissingValue, key)
	}
	return *val, err
}
