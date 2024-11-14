package internal

import "fmt"

type PodStatus string

const (
	StatusCreating PodStatus = "creating"
	StatusRunning  PodStatus = "running"
	StatusStopped  PodStatus = "stopped"
	StatusError    PodStatus = "error"
)

func CreatePodID(pool Pool, id string) string {
	return fmt.Sprintf("%s-%s", pool.ID(), id)
}

type GPUModel string

const (
	GPUModelA100_40G  GPUModel = "A100-40G"
	GPUModelA100_80G  GPUModel = "A100-80G"
	GPUModelA800_40G  GPUModel = "A800-40G"
	GPUModelA800_80G  GPUModel = "A800-80G"
	GPUModelRTX4090   GPUModel = "RTX4090"
	GPUModelRTX4090_D GPUModel = "RTX4090D"
	GPUModelRTX3090   GPUModel = "RTX3090"
)

// PodOptions is the options for creating a pod
type PodOptions struct {
	GPUModel GPUModel
	GPUCount int
}

// Pod represents a pod in a pool
type Pod struct {
	ID                     string
	PoolID                 string
	CreateTimestamp        int64
	DataCenterName         string
	Name                   string
	GPUModel               GPUModel
	GPUCount               int
	CPUModel               string
	CPUCoreCount           int
	MemorySize             int64
	SystemDiskSize         int64
	DataDiskSize           int64
	ExpandableDataDiskSize int64
	DataDiskMountPath      string
	PricePerHour           float64
	SSHDomain              string
	SSHKey                 string
	SSHPort                string
	SSHUser                string
	Password               string
	Status                 string
	ImageID                string
	ImageType              string
	ImageSave              bool
	Pool                   Pool
}
