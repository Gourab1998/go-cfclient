package operation

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestManifestMarshalling(t *testing.T) {
	m := &Manifest{
		Applications: []*AppManifest{
			{
				Name: "spring-music",
			},
		},
	}
	m.Applications[0].WithBuildpacks([]string{"java_buildpack_offline"})
	m.Applications[0].WithEnv(map[string]string{
		"SPRING_CLOUD_PROFILE": "dev",
	})
	approute := AppManifestRoute{}
	approute.WithRoute("spring-music-egregious-porcupine-oa.apps.example.org")

	m.Applications[0].WithRoutes([]AppManifestRoute{approute})
	services := AppManifestService{}
	services.WithName("my-sql")

	m.Applications[0].WithServices([]AppManifestService{services})
	m.Applications[0].WithStack("cflinuxfs3")
	appProcess := AppManifestProcess{}
	appProcess.WithHealthCheckType("http")
	appProcess.WithHealthCheckHTTPEndpoint("/health")
	appProcess.WithInstances(2)
	appProcess.WithMemory("1G")
	appProcess.WithTimeout(60)
	appProcess.WithCommand("java")
	appProcess.WithDiskQuota("1G")
	appProcess.WithLogRateLimitPerSecond("100MB")

	m.Applications[0].WithAppManifestProcess(appProcess)
	b, err := yaml.Marshal(&m)
	require.NoError(t, err)
	require.Equal(t, fullSpringMusicYaml, string(b))

	a := NewAppManifest("spring-music")
	a.WithBuildpacks([]string{"java_buildpack_offline"})
	a.WithMemory("1G")
	a.WithNoRoute(true)
	a.WithStack("cflinuxfs3")

	m = &Manifest{
		Applications: []*AppManifest{a},
	}
	b, err = yaml.Marshal(&m)
	require.NoError(t, err)
	require.Equal(t, minimalSpringMusicYaml, string(b))
}

const fullSpringMusicYaml = `applications:
- name: spring-music
  buildpacks:
  - java_buildpack_offline
  env:
    SPRING_CLOUD_PROFILE: dev
  routes:
  - route: spring-music-egregious-porcupine-oa.apps.example.org
  services:
  - name: my-sql
  stack: cflinuxfs3
  command: java
  disk_quota: 1G
  health-check-type: http
  health-check-http-endpoint: /health
  instances: 2
  log-rate-limit-per-second: 100MB
  memory: 1G
  timeout: 60
`

const minimalSpringMusicYaml = `applications:
- name: spring-music
  buildpacks:
  - java_buildpack_offline
  no-route: true
  stack: cflinuxfs3
  memory: 1G
`
