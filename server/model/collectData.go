package model

type CollectDataStruct struct {
	Memory  MemoryStruct               `json:"memory"`
	CPUs    CPUInfoStruct              `json:"cpus"`
	Disks   []DiskInfo                 `json:"disks"`
	Network map[string]InterfaceStruct `json:"network"`
}

type CPUAttr struct {
	ID     string `json:"cpu"`
	Value  string `json:"value"`
	Sensor string `json:"sensor"`
}

type CPUInfoStruct struct {
	Usage       []CPUAttr `json:"usage"`
	Temperature []CPUAttr `json:"temperature"`
}

type MemoryStruct struct {
	Total float64 `json:"total"`
	Free  float64 `json:"free"`
}

type Smartctl struct {
	Devices []Device `json:"devices"`
}

type Device struct {
	Name     string `json:"name"`
	InfoName string `json:"info_name"`
	Type     string `json:"type"`
	Protocol string `json:"protocol"`
}

type DiskInfo struct {
	ModelName    string       `json:"model_name"`
	SmartStatus  SmartStatus  `json:"smart_status"`
	UserCapacity UserCapacity `json:"user_capacity"`
	Temperature  Temperature  `json:"temperature"`
	PowerOnTime  PowerOnTime  `json:"power_on_time"`
	SerialNumber string       `json:"serial_number"`
	RotationRate interface{}  `json:"rotation_rate,omitempty"`
	Device       Device       `json:"device"`
	SetaVersion  SetaVersion  `json:"seta_version"`
	ScsiVendor   string       `json:"scsi_vendor"`
	ModelType    string
}

type SetaVersion struct {
	String string `json:"string"`
	Value  int64  `json:"value"`
}

type SmartStatus struct {
	Passed bool `json:"passed"`
}

type UserCapacity struct {
	Blocks int64 `json:"blocks"`
	Bytes  int64 `json:"bytes"`
}

type Temperature struct {
	Current int8 `json:"current"`
}

type PowerOnTime struct {
	Hours int64 `json:"hours"`
}

type InterfaceStruct struct {
	Receive  float64 `json:"receive"`
	Transmit float64 `json:"transmit"`
}
