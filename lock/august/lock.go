package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/vkorn/go-august"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/plugins/helpers"
)

// AugustLock describes August Smart Lock device.
type AugustLock struct {
	Settings *Settings
	State    *device.LockState

	logger common.ILoggerProvider
	secret common.ISecretProvider
	spec   *device.Spec

	auth *august.Authenticator
	api  *august.APIProvider
	lock *august.LockDetails

	isWaitingForInput bool
	isLockPicked      bool
	canOperate        bool
}

// Init performs initial plugin init.
func (a *AugustLock) Init(data *device.InitDataDevice) error {
	a.logger = data.Logger
	a.secret = data.Secret

	a.State = &device.LockState{
		GenericDeviceState: device.GenericDeviceState{
			Input: nil,
		},

		On:           false,
		BatteryLevel: 0,
	}

	a.spec = &device.Spec{
		UpdatePeriod:           a.Settings.pollingInterval,
		SupportedCommands:      []enums.Command{enums.CmdOn, enums.CmdOff, enums.CmdToggle, enums.CmdInput},
		SupportedProperties:    []enums.Property{enums.PropOn, enums.PropBatteryLevel, enums.PropInput},
		PostCommandDeferUpdate: 0,
	}

	return nil
}

// Unload is not used, since plugin doesn't keep any open connections.
func (a *AugustLock) Unload() {
}

// GetName returns device name.
// Either device ID or install ID is used.
func (a *AugustLock) GetName() string {
	return a.Settings.finalName
}

// GetSpec returns device specs.
func (a *AugustLock) GetSpec() *device.Spec {
	return a.spec
}

// Input processes user's verification code.
func (a *AugustLock) Input(in common.Input) error {
	val, ok := in.Params[inputKeyName]
	if !ok {
		return errors.New("code was not provided")
	}

	err := a.auth.ValidateCode(val)
	if err != nil {
		a.logger.Error("Failed to verify code, trying to request another one", err,
			common.LogUserNameToken, a.Settings.Username)
		err = a.auth.SendVerificationCode()
		if err != nil {
			a.logger.Error("Failed to request verification code", err,
				common.LogUserNameToken, a.Settings.Username)
			return errors.Wrap(err, "failed to request verification code")
		}
	}

	err = a.auth.Authenticate()
	if err != nil {
		a.logger.Error("Final authentication failed", err,
			common.LogUserNameToken, a.Settings.Username)
		return errors.Wrap(err, "failed to request verification code")
	}

	if a.auth.State.State != august.AuthAuthenticated {
		err = errors.New("failed to authenticate")
		a.logger.Error("Final authentication failed despite success code, don't know what to do", err,
			common.LogUserNameToken, a.Settings.Username)
		return err
	}

	a.State.Input = nil
	a.Settings.Token = a.auth.State.AccessToken
	err = a.secret.Set(a.getSecretName(), a.Settings.Token)
	if err != nil {
		a.logger.Error(fmt.Sprintf("failed to save token, please note it down:\n%s", a.Settings.Token), err,
			common.LogUserNameToken, a.Settings.Username)
	} else {
		a.logger.Info("Saved token to the secret storage", common.LogUserNameToken, a.Settings.Username)
	}

	a.isWaitingForInput = false
	return nil
}

// Load performs initial device connection and state pulls.
func (a *AugustLock) Load() (*device.LockState, error) {
	if "" == a.Settings.Token {
		t, err := a.secret.Get(a.getSecretName())
		if err != nil {
			a.logger.Info("Token wasn't provided and doesn't exist in the secret storage. "+
				"Starting a new auth process", common.LogUserNameToken, a.Settings.Username)
		} else {
			a.Settings.Token = t
			a.logger.Info("Found token in the secret storage", common.LogUserNameToken, a.Settings.Username)
		}
	}

	a.auth = august.NewAuthenticator(a.Settings.authMethod, a.Settings.Username, a.Settings.Password,
		a.Settings.Token, a.Settings.InstallID)

	err := a.auth.Authenticate()
	if err != nil {
		return nil, errors.Wrap(err, "failed to call August auth servers")
	}

	switch a.auth.State.State {
	case august.AuthBadPassword:
		{
			return nil, errors.New("bad username or password")
		}
	case august.AuthRequiresValidation:
		{
			a.logger.Info("Requesting verification code for the user",
				common.LogUserNameToken, a.Settings.Username)
			err := a.auth.SendVerificationCode()
			if err != nil {
				return nil, errors.New("failed to request verification code")
			}

			a.prepareInput()
			return a.State, nil
		}
	}

	return a.Update()
}

// Update pulls device status.
func (a *AugustLock) Update() (*device.LockState, error) {
	if a.isWaitingForInput {
		return a.State, nil
	}

	if nil == a.api {
		a.api = august.NewAPIProvider(a.auth)
	}

	if !a.isLockPicked {
		err := a.findLock()
		if err != nil {
			return nil, err
		}
	}

	l, err := a.api.GetLockDetails(a.Settings.LockID)

	if err != nil {
		if isUnAuth(err) {
			a.logger.Error("Got unauthenticated error. Renewing token", err,
				common.LogUserNameToken, a.Settings.Username)
			_ = a.resetToken()
		}

		return nil, err
	}

	a.lock = l

	a.canOperate = false
	for _, v := range a.lock.Users {
		if helpers.SliceContainsString(v.IDs, a.Settings.userID) {
			a.canOperate = v.CanOperate()
			break
		}
	}

	if !a.canOperate {
		a.spec.SupportedCommands = []enums.Command{enums.CmdInput}
	}

	a.State.On = a.lock.Status.Status == august.LockUnlocked
	a.State.BatteryLevel = uint8(a.lock.Battery * 100)

	return a.State, nil
}

// On makes an attempt to turn device on.
func (a *AugustLock) On() error {
	err := a.checkCommandConditions()
	if err != nil {
		return err
	}

	err = a.api.OpenLock(a.lock.ID)

	if err != nil {
		if isUnAuth(err) {
			a.logger.Error("Got unauthenticated error while opening a lock. Renewing token", err,
				common.LogUserNameToken, a.Settings.Username)
			_ = a.resetToken()
		}

		return err
	}

	return nil
}

// Off makes an attempt to turn device off.
func (a *AugustLock) Off() error {
	err := a.checkCommandConditions()
	if err != nil {
		return err
	}

	err = a.api.CloseLock(a.lock.ID)

	if err != nil {
		if isUnAuth(err) {
			a.logger.Error("Got unauthenticated error while closing a lock. Renewing token", err,
				common.LogUserNameToken, a.Settings.Username)
			_ = a.resetToken()
		}

		return err
	}

	return nil
}

// Toggle makes an attempt to toggle device state.
func (a *AugustLock) Toggle() error {
	err := a.checkCommandConditions()
	if err != nil {
		return err
	}

	if a.lock.Status.Status == august.LockLocked {
		return a.On()
	}

	return a.Off()
}

// Validates whether user can invoke a command.
func (a *AugustLock) checkCommandConditions() error {
	if a.isWaitingForInput {
		return errors.New("waiting for the verification code")
	}

	if !a.isLockPicked {
		return errors.New("lock was not found")
	}

	if !a.canOperate {
		return errors.New("user is not allowed to operate this lock")
	}

	return nil
}

// Prepares user's input.
func (a *AugustLock) prepareInput() {
	a.isWaitingForInput = true
	a.State.Input = &common.Input{
		Title:  fmt.Sprintf("Please provider the verification code sent to %s", a.Settings.Username),
		Params: map[string]string{inputKeyName: "Code"},
	}
}

// Generates token secret name.
func (a *AugustLock) getSecretName() string {
	return fmt.Sprintf("august-lock-%s", a.Settings.InstallID)
}

// Finds a desired lock.
func (a *AugustLock) findLock() error {
	locks, err := a.api.GetLocks()
	if err != nil {
		if isUnAuth(err) {
			a.logger.Error("got unauthenticated error. Renewing token", err,
				common.LogUserNameToken, a.Settings.Username)
			_ = a.resetToken()
		}

		return err
	}

	if "" == a.Settings.LockID && 0 == len(locks) {
		return errors.New("no locks found")
	}

	for _, v := range locks {
		a.logger.Info(fmt.Sprintf("Found available August lock: %s", v.Name),
			common.LogUserNameToken, a.Settings.Username, lockIDLogKeyName, v.ID)
	}

	if "" == a.Settings.LockID {
		a.Settings.LockID = locks[0].ID
		a.isLockPicked = true
		return nil
	}

	for _, v := range locks {
		if v.ID == a.Settings.LockID {
			a.Settings.LockID = v.ID
			a.isLockPicked = true
			return nil
		}
	}

	a.logger.Warn("Lock wasn't found", common.LogUserNameToken, a.Settings.Username,
		lockIDLogKeyName, a.Settings.LockID)
	return errors.New("no locks found")
}

// Resets auth token.
func (a *AugustLock) resetToken() error {
	a.prepareInput()

	err := a.auth.SendVerificationCode()
	if err != nil {
		a.logger.Error("Failed to request verification code", err,
			common.LogUserNameToken, a.Settings.Username)
		return errors.Wrap(err, "failed to request verification code")
	}

	return nil
}

// Validates whether error is actually auth error.
func isUnAuth(err error) bool {
	switch err.(type) {
	case *august.ErrorUnAuth:
		return true
	default:
		return false
	}
}
