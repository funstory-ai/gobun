package xiangongyun

import (
	"bytes"
	"encoding/json"
	"fmt"

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

func (p *Pool) GetPod(id string) (internal.Pod, error) {
	result, err := p.api.DoRequest("GET", "/open/instance/"+id, nil)
	if err != nil {
		return internal.Pod{}, err
	}

	var response struct {
		Code int      `json:"code"`
		Data Instance `json:"data"`
	}

	if err := json.NewDecoder(bytes.NewReader(result)).Decode(&response); err != nil {
		return internal.Pod{}, err
	}

	return internal.Pod{
		ID:                     response.Data.ID,
		PoolID:                 p.id,
		CreateTimestamp:        response.Data.CreateTimestamp,
		DataCenterName:         response.Data.DataCenterName,
		Name:                   response.Data.Name,
		GPUModel:               internal.GPUModel(response.Data.GPUModel),
		GPUCount:               response.Data.GPUUsed,
		CPUModel:               response.Data.CPUModel,
		CPUCoreCount:           response.Data.CPUCoreCount,
		MemorySize:             response.Data.MemorySize,
		SystemDiskSize:         response.Data.SystemDiskSize,
		DataDiskSize:           response.Data.DataDiskSize,
		ExpandableDataDiskSize: response.Data.ExpandableDataDiskSize,
		DataDiskMountPath:      response.Data.DataDiskMountPath,
		PricePerHour:           response.Data.PricePerHour,
		SSHDomain:              response.Data.SSHDomain,
		SSHKey:                 response.Data.SSHKey,
		SSHPort:                response.Data.SSHPort,
		SSHUser:                response.Data.SSHUser,
		Password:               response.Data.Password,
		Status:                 response.Data.Status,
		ImageID:                response.Data.ImageID,
		ImageType:              response.Data.ImageType,
		ImageSave:              response.Data.ImageSave,
		Pool:                   p,
	}, nil
}

func GPUModelMapping(gpuModel internal.GPUModel) (string, error) {
	switch gpuModel {
	case internal.GPUModelRTX4090:
		return "NVIDIA GeForce RTX 4090", nil
	case internal.GPUModelRTX4090_D:
		return "NVIDIA GeForce RTX 4090 D", nil
	default:
		return "", fmt.Errorf("unsupported gpu model: %s", gpuModel)
	}
}

func (p *Pool) CreatePod(options internal.PodOptions) (internal.Pod, error) {
	xgyGPUModel, err := GPUModelMapping(options.GPUModel)
	if err != nil {
		return internal.Pod{}, err
	}

	// Prepare request payload
	payload := map[string]interface{}{
		"gpu_model":      xgyGPUModel,
		"gpu_count":      options.GPUCount,
		"data_center_id": 1, // todo: You might want to make this configurable
		"image":          "2f98442f-1e6e-4531-8b92-88a09d5d8a20",
		"image_type":     "public",
	}

	result, err := p.api.DoRequest("POST", "/open/instance/deploy", payload)
	if err != nil {
		return internal.Pod{}, err
	}

	// Parse response
	var response struct {
		Code int `json:"code"`
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(bytes.NewReader(result)).Decode(&response); err != nil {
		return internal.Pod{}, err
	}

	if response.Code != 200 {
		return internal.Pod{}, fmt.Errorf("failed to create pod, response code: %d", response.Code)
	}
	pod, err := p.GetPod(response.Data.ID)
	if err != nil {
		return internal.Pod{}, err
	}
	return pod, nil
}

func (p *Pool) DestroyPod(podID string) error {
	payload := map[string]interface{}{
		"id": podID,
	}

	result, err := p.api.DoRequest("POST", "/open/instance/shutdown_destroy", payload)
	if err != nil {
		return err
	}

	var response struct {
		Code    int    `json:"code"`
		Msg     string `json:"msg"`
		Success bool   `json:"success"`
	}
	if err := json.NewDecoder(bytes.NewReader(result)).Decode(&response); err != nil {
		return err
	}

	if response.Code != 200 {
		return fmt.Errorf("failed to destroy pod, response code: %d %s", response.Code, response.Msg)
	}

	return nil
}
