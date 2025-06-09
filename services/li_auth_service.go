package services

import (
	"fmt"

	"github.com/vit0-9/li-enricher-api/scraper"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) ValidateSession(sessionCookie, proxyURL string) (bool, error) {
	isValid, err := scraper.ValidateSession(sessionCookie, proxyURL)
	if err != nil {
		return false, fmt.Errorf("session validation request failed: %w", err)
	}
	return isValid, nil
}
