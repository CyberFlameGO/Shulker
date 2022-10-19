package resources

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	shulkermciov1alpha1 "github.com/iamblueslime/shulker/api/v1alpha1"
)

const defaultServerIcon = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAMAAACdt4HsAAAAG1BMVEUAAABSNGCrhKqXaZdsSGtDJlCEUomPY4/////HT7OpAAAACXRSTlMA//////////83ApvUAAAAh0lEQVRYhe3Qyw6AIAxEURTQ//9jmYQmE1IeWwbvytD2LAzhb9JFnQTw0U0tIxsD3lGkUmmIbA7wYRwEJJdUgac0A7qIEDBCEqUKYMlDkpMqgENG8MZHb00ZQBgYYoAdYmZvqoA94tsOsNzOVAHUA3hZHfCW2uVc6x4fACwlABiy9MOEgaP6APk1HDGFXeaaAAAAAElFTkSuQmCC"

type ProxyResourceConfigMapBuilder struct {
	*ProxyResourceBuilder
}

func (b *ProxyResourceBuilder) ProxyConfigMap() *ProxyResourceConfigMapBuilder {
	return &ProxyResourceConfigMapBuilder{b}
}

func (b *ProxyResourceConfigMapBuilder) Build() (client.Object, error) {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.GetConfigMapName(),
			Namespace: b.Instance.Namespace,
			Labels:    b.getLabels(),
		},
	}, nil
}

func (b *ProxyResourceConfigMapBuilder) Update(object client.Object) error {
	configMap := object.(*corev1.ConfigMap)

	configMapData, err := GetConfigMapDataFromConfigSpec(&b.Instance.Spec.Configuration)
	if err != nil {
		return err
	}
	configMap.Data = configMapData

	if err := controllerutil.SetControllerReference(b.Instance, configMap, b.Scheme); err != nil {
		return fmt.Errorf("failed setting controller reference for ConfigMap: %v", err)
	}

	return nil
}

func (b *ProxyResourceConfigMapBuilder) CanBeUpdated() bool {
	return true
}

type configServerYml struct {
	Motd       string `yaml:"motd"`
	Address    string `yaml:"address"`
	Restricted bool   `yaml:"restricted"`
}

type configListenerYml struct {
	Host               string   `yaml:"host"`
	QueryPort          int16    `yaml:"query_port"`
	Motd               string   `yaml:"motd"`
	MaxPlayers         int32    `yaml:"max_players"`
	Priorities         []string `yaml:"priorities"`
	PingPassthrough    bool     `yaml:"ping_passthrough"`
	ForceDefaultServer bool     `yaml:"force_default_server"`
	ProxyProtocol      bool     `yaml:"proxy_protocol"`
}

type configYml struct {
	Servers                 map[string]configServerYml `yaml:"servers"`
	Listeners               []configListenerYml        `yaml:"listeners"`
	Groups                  map[string]interface{}     `yaml:"groups"`
	OnlineMode              bool                       `yaml:"online_mode"`
	IpForward               bool                       `yaml:"ip_forward"`
	PreventProxyConnections bool                       `yaml:"prevent_proxy_connections"`
	EnforceSecureProfile    bool                       `yaml:"enforce_secure_profile"`
}

func getConfigYmlFile(spec *shulkermciov1alpha1.ProxyConfigurationSpec) (string, error) {
	configYml := configYml{
		Servers: map[string]configServerYml{
			"limbo": {
				Motd:       spec.Motd,
				Address:    "localhost:25565",
				Restricted: false,
			},
		},
		Listeners: []configListenerYml{{
			Host:               "0.0.0.0:25577",
			QueryPort:          int16(25577),
			Motd:               spec.Motd,
			MaxPlayers:         spec.MaxPlayers,
			Priorities:         []string{"limbo"},
			PingPassthrough:    false,
			ForceDefaultServer: true,
			ProxyProtocol:      spec.ProxyProtocol,
		}},
		Groups:                  map[string]interface{}{},
		OnlineMode:              true,
		IpForward:               true,
		PreventProxyConnections: true,
		EnforceSecureProfile:    true,
	}

	out, err := yaml.Marshal(&configYml)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func GetConfigMapDataFromConfigSpec(spec *shulkermciov1alpha1.ProxyConfigurationSpec) (map[string]string, error) {
	configMapData := make(map[string]string)

	configMapData["init-fs.sh"] = trimScript(`
		cp -H -r $SHULKER_CONFIG_DIR/* $SHULKER_DATA_DIR/ && \
		if [ -e "$SHULKER_CONFIG_DIR/server-icon.png" ]; then cat $SHULKER_CONFIG_DIR/server-icon.png | base64 -d > $SHULKER_DATA_DIR/server-icon.png; fi
	`)

	if spec.ServerIcon != "" {
		configMapData["server-icon.png"] = spec.ServerIcon
	} else {
		configMapData["server-icon.png"] = defaultServerIcon
	}

	configYml, err := getConfigYmlFile(spec)
	if err != nil {
		return configMapData, err
	}
	configMapData["config.yml"] = configYml

	return configMapData, nil
}

func trimScript(script string) string {
	lines := strings.Split(strings.TrimSpace(script), "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return strings.Join(lines, "\n")
}
