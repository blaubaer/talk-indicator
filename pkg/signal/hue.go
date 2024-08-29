package signal

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/amimof/huego"
	log "github.com/echocat/slf4g"

	"github.com/blaubaer/talk-indicator/pkg/audio"
	"github.com/blaubaer/talk-indicator/pkg/common"
)

const appName = "github.com/blaubaer/talk-indicator"

type Hue struct {
	Pair   bool
	Bridge string
	User   string

	Kinds HueKinds
	Name  *regexp.Regexp

	Britness   uint8
	Hue        uint16
	Saturation uint8

	lights      []huego.Light
	groups      []huego.Group
	credentials HueCredentials
	mutex       sync.Mutex
}

func (this *Hue) Update() error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	bridge, err := this.bridge()
	if err != nil {
		return err
	}

	lights, err := this.discoverLights(bridge)
	if err != nil {
		return err
	}
	groups, err := this.discoverGroups(bridge)
	if err != nil {
		return err
	}

	this.lights = lights
	this.groups = groups

	return nil
}

func (this *Hue) discoverLights(bridge *huego.Bridge) (result []huego.Light, _ error) {
	if this.Kinds.Has(HueKindLight) {
		candidates, err := bridge.GetLights()
		if err != nil {
			return nil, fmt.Errorf("cannot discover lights of bridge %s: %w", bridge.Host, err)
		}
		for _, candidate := range candidates {
			if this.Name.MatchString(candidate.Name) {
				if candidate.State == nil {
					candidate.State = &huego.State{}
				}
				result = append(result, candidate)
			}
		}
	}
	return
}

func (this *Hue) discoverGroups(bridge *huego.Bridge) (result []huego.Group, _ error) {
	if this.Kinds.Has(HueKindGroup) {
		candidates, err := bridge.GetGroups()
		if err != nil {
			return nil, fmt.Errorf("cannot discover groups of bridge %s: %w", bridge.Host, err)
		}
		for _, candidate := range candidates {
			if this.Name.MatchString(candidate.Name) {
				if candidate.State == nil {
					candidate.State = &huego.State{}
				}
				result = append(result, candidate)
			}
		}
	}
	return
}

func (this *Hue) Ensure(state State, _ []audio.Device) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	bridge, err := this.bridge()
	if err != nil {
		return err
	}
	if err := this.ensureLights(bridge, state); err != nil {
		return err
	}
	if err := this.ensureGroups(bridge, state); err != nil {
		return err
	}
	return nil
}

func (this *Hue) ensureLights(bridge *huego.Bridge, state State) error {
	for i, v := range this.lights {
		if err := this.ensureLight(bridge, state, &v); err != nil {
			return err
		}
		this.lights[i] = v
	}
	return nil
}

func (this *Hue) ensureState(state State, title string, hueState *huego.State) (*huego.State, error) {
	switch state {
	case StateOn:
		if !hueState.On || hueState.Bri != this.Britness || hueState.Hue != this.Hue || hueState.Sat != this.Saturation {
			return &huego.State{
				On:  true,
				Bri: this.Britness,
				Hue: this.Hue,
				Sat: this.Saturation,

				Ct: 0,
			}, nil
		}
	case StateOff:
		if hueState.On {
			return &huego.State{
				On: false,
			}, nil
		}
	default:
		return nil, fmt.Errorf("cannot ensure hue light state for %s: %v", title, state)
	}
	return nil, nil
}

func (this *Hue) ensureLight(bridge *huego.Bridge, state State, v *huego.Light) error {
	if newState, err := this.ensureState(state, fmt.Sprintf("light %q#%d", v.Name, v.ID), v.State); err != nil {
		return err
	} else if newState != nil {
		if _, err := bridge.SetLightState(v.ID, *newState); err != nil {
			return fmt.Errorf("cannot switch to hue light state %v for light %q#%d: %w", state, v.Name, v.ID, err)
		}
		v.State = &(*newState)
	}
	return nil
}

func (this *Hue) ensureGroups(bridge *huego.Bridge, state State) error {
	for i, v := range this.groups {
		if err := this.ensureGroup(bridge, state, &v); err != nil {
			return err
		}
		this.groups[i] = v
	}
	return nil
}

func (this *Hue) ensureGroup(bridge *huego.Bridge, state State, v *huego.Group) error {
	if newState, err := this.ensureState(state, fmt.Sprintf("group %q#%d", v.Name, v.ID), v.State); err != nil {
		return err
	} else if newState != nil {
		if _, err := bridge.SetLightState(v.ID, *newState); err != nil {
			return fmt.Errorf("cannot switch to hue light state %v for group %q#%d: %w", state, v.Name, v.ID, err)
		}
		v.State = &(*newState)
	}
	return nil
}

func (this *Hue) SetupConfiguration(using common.FlagHolder) {
	using.Flag("signal.hue.pair", "If true this application will pair again with an existing hue. This will be implicit enabled if this application is not already paired.").
		Envar("TI_SIGNAL_HUE_PAIR").
		BoolVar(&this.Pair)
	using.Flag("signal.hue.bridge", "Usually the bridge is automatically detected. You can specify an explicit one if they are more than one. This is only required while pairing and will afterwards be ignored.").
		Envar("TI_SIGNAL_HUE_BRIDGE").
		StringVar(&this.Bridge)
	using.Flag("signal.hue.user", "Usually this is set while pairing and will then be persisted. If this set this will be used and not be persisted.").
		Envar("TI_SIGNAL_HUE_USER").
		StringVar(&this.User)
	using.Flag("signal.hue.name", "Name as regex of the lights/groups which should be handled by this app.").
		Envar("TI_SIGNAL_HUE_NAME").
		Default("^OnAir").
		RegexpVar(&this.Name)
	using.Flag("signal.hue.kind", "Kind(s) of what should be handled. Possible values: "+AllHueKinds.String()).
		Envar("TI_SIGNAL_HUE_KIND").
		SetValue(&this.Kinds)

	using.Flag("signal.hue.brightness", "The brightness value to set the light to.Brightness is a scale from 1 (the minimum the light is capable of) to 254 (the maximum).").
		Envar("TI_SIGNAL_HUE_BRIGHTNESS").
		Default("254").
		Uint8Var(&this.Britness)
	using.Flag("signal.hue.hue", "The hue value to set light to.The hue value is a wrapping value between 0 and 65535. Both 0 and 65535 are red, 25500 is green and 46920 is blue.").
		Envar("TI_SIGNAL_HUE_HUE").
		Default("65535").
		Uint16Var(&this.Hue)
	using.Flag("signal.hue.saturation", "Saturation of the light. 254 is the most saturated (colored) and 0 is the least saturated (white).").
		Envar("TI_SIGNAL_HUE_SATURATION").
		Default("254").
		Uint8Var(&this.Saturation)
}

func (this *Hue) Initialize() error {
	credentials, err := this.resolveCredentials()
	if err != nil {
		return err
	}
	this.credentials = credentials

	if err := this.Update(); err != nil {
		return err
	}

	return nil
}

func (this *Hue) bridge() (*huego.Bridge, error) {
	credentials := this.credentials
	if credentials.IsZero() {
		return nil, fmt.Errorf("not paired with hue bridge")
	}
	return credentials.Bridge(), nil
}

func (this *Hue) resolveCredentials() (HueCredentials, error) {
	if u := this.User; u != "" {
		bridge, err := this.discoverBridge()
		if err != nil {
			return HueCredentials{}, err
		}

		return HueCredentials{
			Host: bridge.Host,
			User: u,
		}, nil
	}

	if this.Pair {
		credentials, err := this.pair()
		if err != nil {
			return HueCredentials{}, err
		}
		return credentials, nil
	}

	credentials, err := this.readCredentials()
	if err != nil {
		return HueCredentials{}, err
	}

	if credentials.HasContent() {
		return credentials, nil
	}

	return this.pair()
}

func (this *Hue) discoverBridge() (*huego.Bridge, error) {
	if this.Bridge != "" {
		return &huego.Bridge{
			Host: this.Bridge,
		}, nil
	}

	return huego.Discover()
}

func (this *Hue) pair() (HueCredentials, error) {
	bridge, err := this.discoverBridge()
	if err != nil {
		return HueCredentials{}, err
	}

	for {
		log.Info("Wait for hue link button been pressed...")
		user, err := bridge.CreateUser(appName)
		if apiErr, ok := err.(*huego.APIError); ok && apiErr.Type == 101 && apiErr.Description == "link button not pressed" {
			time.Sleep(1 * time.Second)
			continue
		} else if err != nil {
			return HueCredentials{}, fmt.Errorf("was not able to pair with %s: %w", bridge.Host, err)
		} else {
			credentials := HueCredentials{
				Host: bridge.Host,
				User: user,
			}

			if err := this.storeCredentials(credentials); err != nil {
				log.WithError(err).
					Warn("Cannot store credentials. The app will work now, but next time the pairing might be required again.")
			}

			log.With("bridge", bridge.Host).
				Info("Successful paired.")
			return credentials, nil
		}
	}
}

func (this *Hue) Dispose() error {
	return nil
}

func (this *Hue) GetType() Type {
	return TypeHue
}
