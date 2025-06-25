package homeassistant

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/blaubaer/talk-indicator/pkg/common"
	"os"
	"regexp"
	"strings"
	"time"
)

func NewConfiguration() Configuration {
	return Configuration{
		"",
		"",
		fmt.Sprintf("input_boolean.computer-%x-microphone", computerId),
		time.Second * 60,
	}
}

var forbiddenComputerIdChars = regexp.MustCompile("[^a-z0-9_]")

func normalizeEntityIdPrefix(id string) string {
	id = strings.ToLower(id)
	id = strings.TrimSpace(id)
	id = strings.ReplaceAll(id, "-", "_")
	id = strings.ReplaceAll(id, ".", "_")
	id = forbiddenComputerIdChars.ReplaceAllString(id, "_")
	return id
}

var computerId = func() string {
	if result, err := os.Hostname(); err == nil {
		return normalizeEntityIdPrefix(result)
	}

	buf := make([]byte, 8)
	if _, err := rand.Reader.Read(buf); err != nil {
		panic(fmt.Errorf("cannot generate entity id: %v", err))
	}

	return hex.EncodeToString(buf)
}()

type Configuration struct {
	Server   string `yaml:"server,omitempty"`
	Token    string `yaml:"token,omitempty"`
	EntityId string `yaml:"entityId"`

	DeadZoneInterval time.Duration `yaml:"deadZoneInterval,omitempty"`
}

func (this *Configuration) SetupConfiguration(using common.FlagHolder) {
	using.Flag("signal.homeassistant.server", "URL of the Home Assistant instance.").
		Envar("TI_SIGNAL_HOMEASSISTANT_SERVER").
		StringVar(&this.Server)
	using.Flag("signal.homeassistant.token", "Long life token to access the Home Assistant instance.").
		Envar("TI_SIGNAL_HOMEASSISTANT_TOKEN").
		StringVar(&this.Token)
	using.Flag("signal.homeassistant.entityId", "Entity ID to store the information to.").
		Envar("TI_SIGNAL_HOMEASSISTANT_ENTITY_ID").
		StringVar(&this.EntityId)
	using.Flag("signal.homeassistant.deadZoneInterval", "Duration for how long a local state is used to compare to. To prevent too often check of the remote system. As this is the source of truth.").
		Envar("TI_SIGNAL_HOMEASSISTANT_DEAD_ZONE_INTERVAL").
		DurationVar(&this.DeadZoneInterval)
}
