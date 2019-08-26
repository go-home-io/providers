package main

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/vkorn/go-august"
)

const (
	// User's input key name.
	inputKeyName = "code"
	// Log key name.
	lockIdLogKeyName = "lock_id"
)

// Settings describes plugin settings.
type Settings struct {
	LoginMethod     string `yaml:"loginMethod" validate:"required,oneof=phone email" default:"email"`
	Username        string `yaml:"username" validate:"required"`
	Password        string `yaml:"password" validate:"required"`
	Token           string `yaml:"token"`
	LockID          string `yaml:"lockId"`
	InstallID       string `yaml:"installId"`
	PollingInterval int    `yaml:"pollingInterval" validate:"numeric,gte=2" default:"30"`

	pollingInterval time.Duration
	finalName       string
	authMethod      august.AuthenticationMethods
	userID          string
}

// Validate settings.
func (s *Settings) Validate() error {
	if "" == s.LockID && "" == s.InstallID {
		return errors.New("either lockId or installId should be defined")
	}

	if "" != s.LockID {
		s.finalName = s.LockID
	} else {
		s.finalName = s.InstallID
	}

	if "" == s.InstallID {
		s.InstallID = fmt.Sprintf("gohome%s", s.LockID)
	} else {
		s.InstallID = fmt.Sprintf("gohome%s", s.InstallID)
	}

	m, err := august.AuthenticationMethodsString(s.LoginMethod)
	if err != nil {
		return errors.Wrap(err, "wrong login method")
	}

	s.authMethod = m
	if m == august.AuthMethPhone && "+" != s.Username {
		s.Username = fmt.Sprintf("+%s", s.Username)
	}

	s.pollingInterval = time.Duration(s.PollingInterval) * time.Second
	s.userID = fmt.Sprintf("%s:%s", s.authMethod.String(), s.Username)
	return nil
}
