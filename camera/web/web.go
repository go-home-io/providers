package main

import (
	"encoding/base64"
	"sync"
	"time"

	errs "github.com/bdlm/errors"
	"github.com/bdlm/log"
	"github.com/mkenney/go-chrome/codes"
	"github.com/mkenney/go-chrome/tot"
	"github.com/mkenney/go-chrome/tot/emulation"
	"github.com/mkenney/go-chrome/tot/page"
	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device"
	"go-home.io/x/server/plugins/device/enums"
)

// WebCamera describes web page screenshot camera.
type WebCamera struct {
	sync.Mutex
	Settings *Settings
	Logger   common.ILoggerProvider

	state          *device.CameraState
	browser        *chrome.Chrome
	tab            *chrome.Tab
	stopChan       chan bool
	stopErrorsChan chan bool
}

// Init starts remote Chrome communication.
func (c *WebCamera) Init(data *device.InitDataDevice) error {
	log.SetFormatter(&chromeLogger{Logger: data.Logger})
	c.Logger = data.Logger
	c.state = &device.CameraState{}
	c.stopChan = make(chan bool)
	c.stopErrorsChan = make(chan bool)

	c.browser = chrome.New(
		&chrome.Flags{
			"remote-debugging-address": c.Settings.ChromeAddress,
			"addr":                     c.Settings.ChromeAddress,
			"remote-debugging-port":    c.Settings.ChromePort,
			"port":                     c.Settings.ChromePort,
		},
		"", "", "", "")

	err := c.openTab()

	if err != nil {
		return errors.Wrap(err, "open tab failed")
	}

	if c.Settings.ReloadInterval > 0 {
		go c.reloadTab()
	}

	return nil
}

// Unload closes remote tab.
func (c *WebCamera) Unload() {
	c.stopChan <- true
	c.stopErrorsChan <- true

	c.closeTab()

	close(c.stopChan)
	close(c.stopErrorsChan)
}

// GetName returns page address.
func (c *WebCamera) GetName() string {
	return c.Settings.Address
}

// GetSpec returns device specification.
func (c *WebCamera) GetSpec() *device.Spec {
	return &device.Spec{
		SupportedCommands:   []enums.Command{enums.CmdTakePicture},
		SupportedProperties: []enums.Property{enums.PropPicture},
		UpdatePeriod:        time.Duration(c.Settings.PollingInterval) * time.Second,
	}
}

// Load performs initial load.
func (c *WebCamera) Load() (*device.CameraState, error) {
	return c.Update()
}

// Update pulls updated screenshot.
func (c *WebCamera) Update() (*device.CameraState, error) {
	err := c.getPicture()
	if err != nil {
		return nil, errors.Wrap(err, "get picture failed")
	}

	return c.state, nil
}

// TakePicture forces to update a screenshot.
func (c *WebCamera) TakePicture() error {
	return c.getPicture()
}

// Reloads tab if reloadInterval is specified.
//noinspection GoUnhandledErrorResult
func (c *WebCamera) reloadTab() {
	tick := time.Tick(time.Duration(c.Settings.ReloadInterval) * time.Minute)
	for {
		select {
		case <-tick:
			c.closeTab()
			c.openTab() // nolint: gosec
		case <-c.stopChan:
			return
		}
	}
}

// Opens desired tab.
func (c *WebCamera) openTab() error {
	c.Lock()
	defer c.Unlock()

	t, err := c.browser.NewTab(c.Settings.Address)
	if err != nil {
		c.Logger.Error("Failed to open a new tab", err, common.LogURLToken, c.Settings.Address)
		return errors.Wrap(err, "open tab failed")
	}
	enableResult := <-t.Page().Enable()
	if nil != enableResult.Err {
		c.Logger.Error("Failed to enable chrome tab", enableResult.Err, common.LogURLToken, c.Settings.Address)
		return errors.Wrap(err, "enable tab failed")
	}

	loadComplete := make(chan bool)

	t.Page().OnLoadEventFired(func(event *page.LoadEventFiredEvent) {
		overrideResult := <-t.Emulation().SetDeviceMetricsOverride(
			&emulation.SetDeviceMetricsOverrideParams{
				Width:  c.Settings.Width,
				Height: c.Settings.Height,
				ScreenOrientation: &emulation.ScreenOrientation{
					Type:  emulation.OrientationType.LandscapePrimary,
					Angle: 90,
				},
			},
		)
		if nil != overrideResult.Err {
			c.Logger.Error("Failed to setup chrome tab", enableResult.Err, common.LogURLToken, c.Settings.Address)
			return
		}

		loadComplete <- true
	})

	select {
	case <-loadComplete:
		c.tab = t
		go c.handleTabErrors()
		return nil
	case <-time.Tick(5 * time.Second):
		err = errors.New("page load timeout")
		c.Logger.Error("Failed to load page", err, common.LogURLToken, c.Settings.Address)
		return err
	}
}

// Handles errors occurred in a tab.
//noinspection GoUnhandledErrorResult
func (c *WebCamera) handleTabErrors() {
	select {
	case err := <-c.tab.Socket().Errors():
		e, ok := err.(errs.Err)
		if !ok || e.Code() == codes.SocketPanic {
			c.tab = nil
			c.Logger.Warn("Chrome communication error", common.LogURLToken, c.Settings.Address)
			c.openTab() // nolint: gosec
			return
		}
	case <-c.stopErrorsChan:
		return
	}
}

// Closes opened tab.
func (c *WebCamera) closeTab() {
	c.Lock()
	defer c.Unlock()

	if nil == c.tab {
		return
	}

	_, err := c.tab.Close()
	if err != nil {
		c.Logger.Error("Failed to close tab", err, common.LogURLToken, c.Settings.Address)
	}

	c.tab = nil
}

// Takes a screenshot.
//noinspection GoUnhandledErrorResult
func (c *WebCamera) getPicture() error {
	if nil == c.tab {
		go c.openTab()
		return errors.New("tab is currently closed")
	}

	c.Lock()
	defer c.Unlock()

	select {
	case screen := <-c.tab.Page().CaptureScreenshot(
		&page.CaptureScreenshotParams{
			Format:  page.Format.Jpeg,
			Quality: 100,
		}):
		if screen.Err != nil {
			c.Logger.Error("Failed to get page screenshot", screen.Err, common.LogURLToken, c.Settings.Address)
			return screen.Err
		}
		data, err := base64.StdEncoding.DecodeString(screen.Data)
		if err != nil {
			c.Logger.Error("Received corrupted image", err, common.LogURLToken, c.Settings.Address)
		}
		c.state.Picture = string(data)
		return nil
	case <-time.Tick(5 * time.Second):
		err := errors.New("timeout")
		c.Logger.Error("Timeout while making screenshot", err, common.LogURLToken, c.Settings.Address)
		return err
	}
}
