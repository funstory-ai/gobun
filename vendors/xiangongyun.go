package vendors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/funstory-ai/gobun/internal/resource"
)

type Instance struct {
	ID                     string  `json:"id"`
	CreateTimestamp        int64   `json:"create_timestamp"`
	DataCenterName         string  `json:"data_center_name"`
	Name                   string  `json:"name"`
	PublicImage            string  `json:"public_image"`
	GPUModel               string  `json:"gpu_model"`
	GPUUsed                int     `json:"gpu_used"`
	CPUModel               string  `json:"cpu_model"`
	CPUCoreCount           int     `json:"cpu_core_count"`
	MemorySize             int64   `json:"memory_size"`
	SystemDiskSize         int64   `json:"system_disk_size"`
	DataDiskSize           int64   `json:"data_disk_size"`
	ExpandableDataDiskSize int64   `json:"expandable_data_disk_size"`
	DataDiskMountPath      string  `json:"data_disk_mount_path"`
	StorageMountPath       string  `json:"storage_mount_path"`
	PricePerHour           float64 `json:"price_per_hour"`
	SSHKey                 string  `json:"ssh_key"`
	SSHPort                string  `json:"ssh_port"`
	SSHUser                string  `json:"ssh_user"`
	Password               string  `json:"password"`
	Status                 string  `json:"status"`
	SSHDomain              string  `json:"ssh_domain"`
	ImageID                string  `json:"image_id"`
	ImageType              string  `json:"image_type"`
	ImageSave              bool    `json:"image_save"`
}

type XianGongYunProvider struct {
	authorization string
	client        *http.Client
}

func (p *XianGongYunProvider) Init(ctx context.Context, options VendorOptions) error {
	authorization, ok := options.APISecret["authorization"]
	if !ok {
		return fmt.Errorf("missing required authorization in APISecret")
	}
	p.authorization = authorization
	p.client = &http.Client{}
	return nil
}

func (p *XianGongYunProvider) doRequest(method string, path string, body interface{}) ([]byte, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	url := "https://api.xiangongyun.com" + path
	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", p.authorization)
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (p *XianGongYunProvider) gpuModelMapping(gpuModel string) (string, error) {
	switch gpuModel {
	case "RTX4090":
		return "NVIDIA GeForce RTX 4090", nil
	case "RTX4090_D":
		return "NVIDIA GeForce RTX 4090 D", nil
	default:
		return "", fmt.Errorf("unsupported gpu model: %s", gpuModel)
	}
}

func (p *XianGongYunProvider) CreatePod(ctx context.Context, pod resource.PodOptions) (resource.Pod, error) {
	xgyGPUModel, err := p.gpuModelMapping(string(pod.GPUModel))
	if err != nil {
		return resource.Pod{}, err
	}

	payload := map[string]interface{}{
		"gpu_model":      xgyGPUModel,
		"gpu_count":      pod.GPUCount,
		"data_center_id": 1,
		"image":          "2f98442f-1e6e-4531-8b92-88a09d5d8a20",
		"image_type":     "public",
	}

	result, err := p.doRequest("POST", "/open/instance/deploy", payload)
	if err != nil {
		return resource.Pod{}, err
	}

	var response struct {
		Code int `json:"code"`
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(bytes.NewReader(result)).Decode(&response); err != nil {
		return resource.Pod{}, err
	}

	if response.Code != 200 {
		return resource.Pod{}, fmt.Errorf("failed to create pod, response code: %d", response.Code)
	}

	return p.getPod(response.Data.ID)
}

func (p *XianGongYunProvider) GetPod(ctx context.Context, pod resource.Pod) (resource.Pod, error) {
	return p.getPod(pod.ID)
}

func (p *XianGongYunProvider) getPod(id string) (resource.Pod, error) {
	result, err := p.doRequest("GET", "/open/instance/"+id, nil)
	if err != nil {
		return resource.Pod{}, err
	}

	var response struct {
		Code int      `json:"code"`
		Data Instance `json:"data"`
	}

	if err := json.NewDecoder(bytes.NewReader(result)).Decode(&response); err != nil {
		return resource.Pod{}, err
	}

	return resource.Pod{
		ID:              response.Data.ID,
		Name:            response.Data.Name,
		GPUModel:        resource.GPUModel(response.Data.GPUModel),
		GPUCount:        response.Data.GPUUsed,
		CPUModel:        response.Data.CPUModel,
		CPUCoreCount:    response.Data.CPUCoreCount,
		MemorySize:      response.Data.MemorySize,
		SystemDiskSize:  response.Data.SystemDiskSize,
		DataDiskSize:    response.Data.DataDiskSize,
		SSHDomain:       response.Data.SSHDomain,
		SSHPort:         response.Data.SSHPort,
		SSHUser:         response.Data.SSHUser,
		Password:        response.Data.Password,
		Status:          response.Data.Status,
		CreateTimestamp: response.Data.CreateTimestamp,
	}, nil
}

func (p *XianGongYunProvider) DestroyPod(ctx context.Context, pod resource.Pod) error {
	payload := map[string]interface{}{
		"id": pod.ID,
	}

	result, err := p.doRequest("POST", "/open/instance/shutdown_destroy", payload)
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

func (p *XianGongYunProvider) ListPods(ctx context.Context) ([]resource.Pod, error) {
	result, err := p.doRequest("GET", "/open/instances", nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Code int `json:"code"`
		Data struct {
			List []Instance `json:"list"`
		} `json:"data"`
	}
	if err := json.NewDecoder(bytes.NewReader(result)).Decode(&response); err != nil {
		return nil, err
	}

	pods := make([]resource.Pod, len(response.Data.List))
	for i, instance := range response.Data.List {
		pods[i] = resource.Pod{
			ID:              instance.ID,
			Name:            instance.Name,
			GPUModel:        resource.GPUModel(instance.GPUModel),
			GPUCount:        instance.GPUUsed,
			CPUModel:        instance.CPUModel,
			CPUCoreCount:    instance.CPUCoreCount,
			MemorySize:      instance.MemorySize,
			SystemDiskSize:  instance.SystemDiskSize,
			DataDiskSize:    instance.DataDiskSize,
			SSHDomain:       instance.SSHDomain,
			SSHPort:         instance.SSHPort,
			SSHUser:         instance.SSHUser,
			Password:        instance.Password,
			Status:          instance.Status,
			CreateTimestamp: instance.CreateTimestamp,
		}
	}

	return pods, nil
}
