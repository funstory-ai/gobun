package xiangongyun

import (
	"bytes"
	"encoding/json"

	"github.com/funstory-ai/gobun/internal"
)

const (
	PoolID = "xiangongyun"
)

type Pool struct {
	api *API
	id  string
}

func NewPool(authorization string) *Pool {
	return &Pool{
		api: InitAPI(authorization),
		id:  PoolID,
	}
}

func (p *Pool) ID() string {
	return p.id
}

func (p *Pool) ListPods() ([]internal.Pod, error) {
	result, err := p.api.DoRequest("GET", "/open/instances", nil)
	if err != nil {
		return nil, err
	}
	var response struct {
		Code int `json:"code"`
		Data struct {
			List []Instance `json:"list"`
		} `json:"data"`
	}
	reader := bytes.NewReader(result)
	if err := json.NewDecoder(reader).Decode(&response); err != nil {
		return nil, err
	}
	pods := make([]internal.Pod, len(response.Data.List))
	for i, instance := range response.Data.List {
		pods[i] = internal.Pod{
			ID:                     instance.ID,
			PoolID:                 p.id,
			CreateTimestamp:        instance.CreateTimestamp,
			DataCenterName:         instance.DataCenterName,
			Name:                   instance.Name,
			GPUModel:               internal.GPUModel(instance.GPUModel),
			GPUCount:               instance.GPUUsed,
			CPUModel:               instance.CPUModel,
			CPUCoreCount:           instance.CPUCoreCount,
			MemorySize:             instance.MemorySize,
			SystemDiskSize:         instance.SystemDiskSize,
			DataDiskSize:           instance.DataDiskSize,
			ExpandableDataDiskSize: instance.ExpandableDataDiskSize,
			DataDiskMountPath:      instance.DataDiskMountPath,
			PricePerHour:           instance.PricePerHour,
			SSHDomain:              instance.SSHDomain,
			SSHKey:                 instance.SSHKey,
			SSHPort:                instance.SSHPort,
			SSHUser:                instance.SSHUser,
			Password:               instance.Password,
			Status:                 instance.Status,
			ImageID:                instance.ImageID,
			ImageType:              instance.ImageType,
			ImageSave:              instance.ImageSave,
			Pool:                   p,
		}
	}
	return pods, nil
}

func (p *Pool) CreatePod() (internal.PodOptions, error) {
	return internal.PodOptions{}, nil
}

func (p *Pool) DestroyPod(podID string) error {
	return nil
}
