package operation

import "github.com/cloudfoundry-community/go-cfclient/v3/resource"

type AppHealthCheckType string

const (
	Http    AppHealthCheckType = "http"
	Port    AppHealthCheckType = "port"
	Process AppHealthCheckType = "process"
)

type AppProcessType string

const (
	Web    AppProcessType = "web"
	Worker AppProcessType = "worker"
)

type Manifest struct {
	Version      string         `yaml:"version,omitempty"`
	Applications []*AppManifest `yaml:"applications"`
}

type AppManifest struct {
	Name               string                `yaml:"name"`
	Path               *string               `yaml:"path,omitempty"`
	Buildpacks         *[]string             `yaml:"buildpacks,omitempty"`
	Docker             *AppManifestDocker    `yaml:"docker,omitempty"`
	Env                *map[string]string    `yaml:"env,omitempty"`
	RandomRoute        *bool                 `yaml:"random-route,omitempty"`
	NoRoute            *bool                 `yaml:"no-route,omitempty"`
	DefaultRoute       *bool                 `yaml:"default-route,omitempty"`
	Routes             *AppManifestRoutes    `yaml:"routes,omitempty"`
	Services           *AppManifestServices  `yaml:"services,omitempty"`
	Sidecars           *AppManifestSideCars  `yaml:"sidecars,omitempty"`
	Processes          *AppManifestProcesses `yaml:"processes,omitempty"`
	Stack              *string               `yaml:"stack,omitempty"`
	Metadata           *resource.Metadata    `yaml:"metadata,omitempty"`
	AppManifestProcess `yaml:",inline"`
}

type AppManifestDocker struct {
	Image    *string `yaml:"image,omitempty"`
	Username *string `yaml:"username,omitempty"`
}

type AppManifestRoutes []AppManifestRoute

type AppManifestRoute struct {
	Route    *string `yaml:"route,omitempty"`
	Protocol *string `yaml:"protocol,omitempty"`
}

type AppManifestServices []AppManifestService

type AppManifestService struct {
	Name       *string                 `yaml:"name,omitempty"`
	Parameters *map[string]interface{} `yaml:"parameters,omitempty"`
}

type AppManifestSideCars []AppManifestSideCar

type AppManifestSideCar struct {
	Name         *string   `yaml:"name,omitempty"`
	ProcessTypes *[]string `yaml:"process_types,omitempty"`
	Command      *string   `yaml:"command,omitempty"`
	Memory       *string   `yaml:"memory,omitempty"`
}

type AppManifestProcesses []AppManifestProcess

type AppManifestProcess struct {
	Type                         *AppProcessType     `yaml:"type,omitempty"`
	Command                      *string             `yaml:"command,omitempty"`
	DiskQuota                    *string             `yaml:"disk_quota,omitempty"`
	HealthCheckType              *AppHealthCheckType `yaml:"health-check-type,omitempty"`
	HealthCheckHTTPEndpoint      *string             `yaml:"health-check-http-endpoint,omitempty"`
	HealthCheckInvocationTimeout *uint               `yaml:"health-check-invocation-timeout,omitempty"`
	Instances                    *uint               `yaml:"instances,omitempty"`
	LogRateLimitPerSecond        *string             `yaml:"log-rate-limit-per-second,omitempty"`
	Memory                       *string             `yaml:"memory,omitempty"`
	Timeout                      *uint               `yaml:"timeout,omitempty"`
}

// setters for AppManifest
func (a *AppManifest) WithPath(path string) {
	a.Path = &path
}
func (a *AppManifest) WithBuildpacks(buildpacks []string) {
	a.Buildpacks = &buildpacks
}
func (a *AppManifest) WithDocker(docker *AppManifestDocker) {
	a.Docker = docker
}
func (a *AppManifest) WithEnv(env map[string]string) {
	a.Env = &env
}
func (a *AppManifest) WithRandomRoute(randomRoute bool) {
	a.RandomRoute = &randomRoute
}
func (a *AppManifest) WithNoRoute(noRoute bool) {
	a.NoRoute = &noRoute
}
func (a *AppManifest) WithDefaultRoute(defaultRoute bool) {
	a.DefaultRoute = &defaultRoute
}
func (a *AppManifest) WithRoutes(routes AppManifestRoutes) {
	a.Routes = &routes
}
func (a *AppManifest) WithServices(services AppManifestServices) {
	a.Services = &services
}
func (a *AppManifest) WithSidecars(sidecars AppManifestSideCars) {
	a.Sidecars = &sidecars
}
func (a *AppManifest) WithProcesses(processes AppManifestProcesses) {
	a.Processes = &processes
}
func (a *AppManifest) WithStack(stack string) {
	a.Stack = &stack
}
func (a *AppManifest) WithMetadata(metadata *resource.Metadata) {
	a.Metadata = metadata
}
func (a *AppManifest) WithAppManifestProcess(appManifestProcess AppManifestProcess) {
	a.AppManifestProcess = appManifestProcess
}

func (d *AppManifestDocker) WithImage(image string) {
	d.Image = &image
}
func (d *AppManifestDocker) WithUsername(username string) {
	d.Username = &username
}

// setters for AppManifestRoute
func (r *AppManifestRoute) WithRoute(route string) {
	r.Route = &route
}
func (r *AppManifestRoute) WithProtocol(protocol string) {
	r.Protocol = &protocol
}

// setters for AppManifestService
func (s *AppManifestService) WithName(name string) {
	s.Name = &name
}
func (s *AppManifestService) WithParameters(parameters map[string]interface{}) {
	s.Parameters = &parameters
}

// setters for AppManifestSideCar
func (s *AppManifestSideCar) WithName(name string) {
	s.Name = &name
}
func (s *AppManifestSideCar) WithProcessTypes(processTypes []string) {
	s.ProcessTypes = &processTypes
}
func (s *AppManifestSideCar) WithCommand(command string) {
	s.Command = &command
}
func (s *AppManifestSideCar) WithMemory(memory string) {
	s.Memory = &memory
}

// setters for AppManifestProcess
func (p *AppManifestProcess) WithType(appProcessType AppProcessType) {
	p.Type = &appProcessType
}
func (p *AppManifestProcess) WithCommand(command string) {
	p.Command = &command
}
func (p *AppManifestProcess) WithDiskQuota(diskQuota string) {
	p.DiskQuota = &diskQuota
}
func (p *AppManifestProcess) WithHealthCheckType(healthCheckType AppHealthCheckType) {
	p.HealthCheckType = &healthCheckType
}
func (p *AppManifestProcess) WithHealthCheckHTTPEndpoint(healthCheckHTTPEndpoint string) {
	p.HealthCheckHTTPEndpoint = &healthCheckHTTPEndpoint
}
func (p *AppManifestProcess) WithHealthCheckInvocationTimeout(healthCheckInvocationTimeout uint) {
	p.HealthCheckInvocationTimeout = &healthCheckInvocationTimeout
}
func (p *AppManifestProcess) WithInstances(instances uint) {
	p.Instances = &instances
}
func (p *AppManifestProcess) WithLogRateLimitPerSecond(logRateLimitPerSecond string) {
	p.LogRateLimitPerSecond = &logRateLimitPerSecond
}
func (p *AppManifestProcess) WithMemory(memory string) {
	p.Memory = &memory
}
func (p *AppManifestProcess) WithTimeout(timeout uint) {
	p.Timeout = &timeout
}

func NewManifest(applications ...*AppManifest) *Manifest {
	return &Manifest{
		Version:      "1",
		Applications: applications,
	}
}

func NewAppManifest(appName string) *AppManifest {
	return &AppManifest{
		Name: appName,
	}
}
