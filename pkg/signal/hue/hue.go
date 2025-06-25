package hue

import (
	"fmt"
	"github.com/amimof/huego"
	"github.com/blaubaer/talk-indicator/pkg/credentials"
	"github.com/blaubaer/talk-indicator/pkg/signal"
	log "github.com/echocat/slf4g"
	"sync"
	"time"
)

const appName = "github.com/blaubaer/talk-indicator"

type Hue struct {
	conf         *Configuration
	saveConfFunc func() error

	lights      []huego.Light
	groups      []huego.Group
	credentials credentials.Credentials
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
	if this.conf.Kinds.Has(HueKindLight) {
		candidates, err := bridge.GetLights()
		if err != nil {
			return nil, fmt.Errorf("cannot discover lights of bridge %s: %w", bridge.Host, err)
		}
		for _, candidate := range candidates {
			if this.conf.Name.MatchString(candidate.Name) {
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
	if this.conf.Kinds.Has(HueKindGroup) {
		candidates, err := bridge.GetGroups()
		if err != nil {
			return nil, fmt.Errorf("cannot discover groups of bridge %s: %w", bridge.Host, err)
		}
		for _, candidate := range candidates {
			if this.conf.Name.MatchString(candidate.Name) {
				if candidate.State == nil {
					candidate.State = &huego.State{}
				}
				result = append(result, candidate)
			}
		}
	}
	return
}

func (this *Hue) Ensure(ctx signal.Context) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	bridge, err := this.bridge()
	if err != nil {
		return err
	}
	if err := this.ensureLights(bridge, ctx.State()); err != nil {
		return err
	}
	if err := this.ensureGroups(bridge, ctx.State()); err != nil {
		return err
	}
	return nil
}

func (this *Hue) ensureLights(bridge *huego.Bridge, state signal.State) error {
	for i, v := range this.lights {
		if err := this.ensureLight(bridge, state, &v); err != nil {
			return err
		}
		this.lights[i] = v
	}
	return nil
}

func (this *Hue) ensureState(state signal.State, title string, hueState *huego.State) (*huego.State, error) {
	switch state {
	case signal.StateOn:
		if !hueState.On || hueState.Bri != this.conf.Brightness || hueState.Hue != this.conf.Hue || hueState.Sat != this.conf.Saturation {
			return &huego.State{
				On:  true,
				Bri: this.conf.Brightness,
				Hue: this.conf.Hue,
				Sat: this.conf.Saturation,

				Ct: 0,
			}, nil
		}
	case signal.StateOff:
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

func (this *Hue) ensureLight(bridge *huego.Bridge, state signal.State, v *huego.Light) error {
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

func (this *Hue) ensureGroups(bridge *huego.Bridge, state signal.State) error {
	for i, v := range this.groups {
		if err := this.ensureGroup(bridge, state, &v); err != nil {
			return err
		}
		this.groups[i] = v
	}
	return nil
}

func (this *Hue) ensureGroup(bridge *huego.Bridge, state signal.State, v *huego.Group) error {
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

func (this *Hue) Initialize(conf *Configuration, saveConfFunc func() error) error {
	this.conf = conf
	this.saveConfFunc = saveConfFunc

	v, err := this.resolveCredentials()
	if err != nil {
		return err
	}
	this.credentials = v

	if err := this.Update(); err != nil {
		return err
	}

	return nil
}

func (this *Hue) bridge() (*huego.Bridge, error) {
	v := this.credentials
	if v.IsHueZero() {
		return nil, fmt.Errorf("not paired with hue bridge")
	}
	return huego.New(v.HueBridge, v.HueUser), nil
}

func (this *Hue) resolveCredentials() (credentials.Credentials, error) {
	if u := this.conf.User; u != "" {
		bridge, err := this.discoverBridge()
		if err != nil {
			return credentials.Credentials{}, err
		}

		return credentials.Credentials{
			HueBridge: bridge.Host,
			HueUser:   u,
		}, nil
	}

	if this.conf.Pair {
		v, err := this.pair()
		if err != nil {
			return credentials.Credentials{}, err
		}
		return v, nil
	}

	v, err := this.readCredentials()
	if err != nil {
		return credentials.Credentials{}, err
	}

	if !v.IsHueZero() {
		return v, nil
	}

	return this.pair()
}

func (this *Hue) discoverBridge() (*huego.Bridge, error) {
	if this.conf.Bridge != "" {
		return &huego.Bridge{
			Host: this.conf.Bridge,
		}, nil
	}

	return huego.Discover()
}

func (this *Hue) pair() (credentials.Credentials, error) {
	bridge, err := this.discoverBridge()
	if err != nil {
		return credentials.Credentials{}, err
	}

	for {
		log.Info("Wait for hue link button been pressed...")
		user, err := bridge.CreateUser(appName)
		if apiErr, ok := err.(*huego.APIError); ok && apiErr.Type == 101 && apiErr.Description == "link button not pressed" {
			time.Sleep(1 * time.Second)
			continue
		} else if err != nil {
			return credentials.Credentials{}, fmt.Errorf("was not able to pair with %s: %w", bridge.Host, err)
		} else {
			v := credentials.Credentials{
				HueBridge: bridge.Host,
				HueUser:   user,
			}

			if err := this.storeCredentials(v); err != nil {
				log.WithError(err).
					Warn("Cannot store credentials. The app will work now, but next time the pairing might be required again.")
			}

			log.With("bridge", bridge.Host).
				Info("Successful paired.")
			return v, nil
		}
	}
}

func (this *Hue) Dispose() error {
	this.conf = nil
	this.saveConfFunc = nil
	return nil
}

func (this *Hue) GetType() signal.Type {
	return signal.TypeHue
}

func (this *Hue) readCredentials() (credentials.Credentials, error) {
	var v credentials.Credentials
	if _, err := v.ReadFromStore(); err != nil {
		return credentials.Credentials{}, err
	}

	if v.HueBridge == "" {
		v.HueBridge = this.conf.Bridge
	}
	if v.HueUser == "" {
		v.HueUser = this.conf.User
	}

	return v, nil
}

func (this *Hue) storeCredentials(v credentials.Credentials) error {
	supported, err := v.WriteToStore()
	if err != nil {
		return err
	}
	if supported {
		return nil
	}

	this.conf.Bridge = v.HueBridge
	this.conf.User = v.HueUser
	return this.saveConfFunc()
}
