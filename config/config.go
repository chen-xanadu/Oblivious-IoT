package config

const (
	DeviceSkFile     = "config/device.rsa"
	DevicePkFile     = "config/device.rsa.pub"
	VendorSkFile     = "config/vendor.rsa"
	VendorPkFile     = "config/vendor.rsa.pub"
	IntegratorSkFile = "config/integrator.rsa"
	IntegratorPkFile = "config/integrator.rsa.pub"
)

var (
	DeviceHmacKey = []byte("device1")
)

const (
	RoundID  = 1
	VendorID = 0
)

const (
	NumDevices           = 10
	NumCommands          = 10
	MaxCommandsPerDevice = 10
)
