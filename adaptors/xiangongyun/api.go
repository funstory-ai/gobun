package xiangongyun

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
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
	JupyterToken           string  `json:"jupyter_token"`
	JupyterURL             string  `json:"jupyter_url"`
	XGCOSToken             string  `json:"xgcos_token"`
	XGCOSURL               string  `json:"xgcos_url"`
	StartTimestamp         int64   `json:"start_timestamp"`
	StopTimestamp          int64   `json:"stop_timestamp"`
	Status                 string  `json:"status"`
	SSHDomain              string  `json:"ssh_domain"`
	WebURL                 string  `json:"web_url"`
	Progress               int     `json:"progress"`
	ImageID                string  `json:"image_id"`
	ImageType              string  `json:"image_type"`
	ImagePrice             float64 `json:"image_price"`
	ImageSave              bool    `json:"image_save"`
	BasePrice              float64 `json:"base_price"`
	AutoShutdown           int     `json:"auto_shutdown"`
	AutoShutdownAction     int     `json:"auto_shutdown_action"`
}

type API struct {
	authorization string
	client        *http.Client
}

func InitAPI(authorization string) *API {
	return &API{
		authorization: authorization,
		client:        &http.Client{},
	}
}

func (api *API) DoRequest(method string, path string, body interface{}) ([]byte, error) {
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
	req.Header.Set("Authorization", api.authorization)
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body to bytes
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}
