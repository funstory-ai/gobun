package vendors

import (
	"context"
	"fmt"

	"github.com/funstory-ai/gobun/internal/resource"
)

type Vendor struct {
	Id         string
	Name       string
	Provider   CloudProvider
	ProviderId string
}

type VendorOptions struct {
	APISecret map[string]string
}

type CloudProvider interface {
	Init(ctx context.Context, options VendorOptions) error
	CreatePod(ctx context.Context, pod resource.PodOptions) (resource.Pod, error)
	DestroyPod(ctx context.Context, pod resource.Pod) error
	ListPods(ctx context.Context) ([]resource.Pod, error)
	GetPod(ctx context.Context, pod resource.Pod) (resource.Pod, error)
}

type CloudProviderType string

const (
	CloudProviderTypeXianGongYun CloudProviderType = "xiangongyun"
)

type CloudProviderMetadata struct {
	ProviderName string
}

func NewVendor(ctx context.Context, provider CloudProviderType, options VendorOptions) (Vendor, error) {
	switch provider {
	case CloudProviderTypeXianGongYun:
		cloudProvider := &XianGongYunProvider{}
		if err := cloudProvider.Init(ctx, options); err != nil {
			return Vendor{}, err
		}
		return Vendor{
			Id:         string(CloudProviderTypeXianGongYun),
			Name:       "xiangongyun",
			Provider:   cloudProvider,
			ProviderId: "xiangongyun",
		}, nil
	}
	return Vendor{}, fmt.Errorf("invalid provider: %s", provider)
}

func (v *Vendor) CreatePod(ctx context.Context, pod resource.PodOptions) (resource.Pod, error) {
	return v.Provider.CreatePod(ctx, pod)
}

func (v *Vendor) DestroyPod(ctx context.Context, pod resource.Pod) error {
	return v.Provider.DestroyPod(ctx, pod)
}

func (v *Vendor) ListPods(ctx context.Context) ([]resource.Pod, error) {
	return v.Provider.ListPods(ctx)
}

func (v *Vendor) GetPod(ctx context.Context, pod resource.Pod) (resource.Pod, error) {
	return v.Provider.GetPod(ctx, pod)
}
