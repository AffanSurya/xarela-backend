package service

type HealthService struct{}

type HealthStatus struct {
	Status string `json:"status"`
}

func NewHealthService() HealthService {
	return HealthService{}
}

func (HealthService) Status() HealthStatus {
	return HealthStatus{Status: "ok"}
}
