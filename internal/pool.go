package internal

// Pool represents an abstraction for cloud Pod management, for now we only support a could provider is a pool
type Pool interface {
	// Return the unique identifier of the pool
	ID() string
	// Returns the created Pod and any error encountered
	CreatePod(PodOptions) (Pod, error)

	// DestroyPod removes a Pod from the cloud
	// Takes a Pod ID and returns any error encountered
	DestroyPod(PodID string) error

	// ListPods returns a list of all Pods in the pool
	// Returns a slice of Pods and any error encountered
	ListPods() ([]Pod, error)
}
