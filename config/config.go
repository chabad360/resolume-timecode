package config

import (
	"fyne.io/fyne/v2"
	"sync"
)

const (
	OSCOutPort    = "OSCOutPort"
	OSCPort       = "OSCPort"
	OSCAddr       = "OSCAddr"
	ClipPath      = "clipPath"
	ClipInvert    = "clipInvert"
	ClientMessage = "message"

	EnableHttpClient = "enableHttpClient"
	HTTPPort         = "httpPort"

	EnableOSCClient = "enableOSCClient"
	OSCClientAddr   = "OSCClientAddr"
	OSCClientPort   = "OSCClientPort"

	AlertTime = "alertTime"
)

var (
	StringConfig = make(map[string]string)
	IntConfig    = make(map[string]int)
	BoolConfig   = make(map[string]bool)

	DefaultStringConfig = map[string]string{
		OSCOutPort:    "7001",
		OSCPort:       "7000",
		OSCAddr:       "127.0.0.1",
		HTTPPort:      "8080",
		ClipPath:      "",
		ClientMessage: "",
		OSCClientAddr: "",
		OSCClientPort: "",
	}
	DefaultIntConfig = map[string]int{
		AlertTime: 10,
	}
	DefaultBoolConfig = map[string]bool{
		ClipInvert:       false,
		EnableOSCClient:  false,
		EnableHttpClient: true,
	}

	m = sync.RWMutex{}

	a fyne.App
)

func SetString(key string, value string) {
	m.Lock()
	defer m.Unlock()
	StringConfig[key] = value
}

func GetString(key string) string {
	m.RLock()
	defer m.RUnlock()
	return StringConfig[key]
}

func SetInt(key string, value int) {
	m.Lock()
	defer m.Unlock()
	IntConfig[key] = value
}

func GetInt(key string) int {
	m.RLock()
	defer m.RUnlock()
	return IntConfig[key]
}

func SetBool(key string, value bool) {
	m.Lock()
	defer m.Unlock()
	BoolConfig[key] = value
}

func GetBool(key string) bool {
	m.RLock()
	defer m.RUnlock()
	return BoolConfig[key]
}

func StoreValues() {
	m.RLock()
	defer m.RUnlock()
	for key, value := range StringConfig {
		a.Preferences().SetString(key, value)
	}
	for key, value := range IntConfig {
		a.Preferences().SetInt(key, value)
	}
	for key, value := range BoolConfig {
		a.Preferences().SetBool(key, value)
	}
}

func loadValues() {
	m.Lock()
	defer m.Unlock()
	for key, value := range DefaultStringConfig {
		StringConfig[key] = a.Preferences().StringWithFallback(key, value)
	}
	for key, value := range DefaultIntConfig {
		IntConfig[key] = a.Preferences().IntWithFallback(key, value)
	}
	for key, value := range DefaultBoolConfig {
		BoolConfig[key] = a.Preferences().BoolWithFallback(key, value)
	}
}

func Init(app fyne.App) {
	a = app
	loadValues()
}
