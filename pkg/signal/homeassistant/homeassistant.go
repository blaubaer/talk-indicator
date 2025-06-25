package homeassistant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/blaubaer/talk-indicator/pkg/common"
	"github.com/blaubaer/talk-indicator/pkg/credentials"
	"github.com/blaubaer/talk-indicator/pkg/signal"
	log "github.com/echocat/slf4g"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const DefaultServer = "http://homeassistant.local:8123/"

type Homeassistant struct {
	conf         *Configuration
	saveConfFunc func() error
	mutex        sync.RWMutex

	lastState atomic.Pointer[state]

	client http.Client
}

func (this *Homeassistant) Update() error {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	rsp, err := this.do("GET", "/api/")
	if err != nil {
		return err
	}
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d - %s", rsp.StatusCode, rsp.Status)
	}

	return nil
}

type state struct {
	timestamp time.Time
	state     signal.State
	sessions  stateAttrSessions
}

func (this *state) isEqualTo(o *state) bool {
	return this.state == o.state &&
		this.sessions.isEqualTo(&o.sessions)
}

func (this *Homeassistant) Ensure(ctx signal.Context) error {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	target := state{
		state:     ctx.State(),
		timestamp: time.Now(),
		sessions:  stateAttrSessions{},
	}

	for sess, err := range ctx.Sessions() {
		if err != nil {
			return err
		}
		target.sessions = append(target.sessions, stateAttrSession{
			sess.Title,
			sess.Device.Name,
			sess.Identifier,
		})
	}

	logger := log.With("entityId", this.conf.EntityId)

	if v := this.lastState.Load(); v != nil {
		if v.timestamp.Add(this.conf.DeadZoneInterval).After(time.Now()) {
			if v.isEqualTo(&target) {
				logger.Debug("Entity is already in requested state (while dead zone timeout). No updated needed.")
				return nil
			}
		}
	}

	rsp, err := this.do("GET", "/api/states/"+this.conf.EntityId)
	if err != nil {
		return err
	}
	defer func() {
		_ = rsp.Body.Close()
	}()

	current := state{
		timestamp: time.Now(),
	}
	attributes := make(map[string]any)
	forceUpdate := false

	switch rsp.StatusCode {
	case http.StatusOK:
		var gRsp stateGetResponse
		if err := json.NewDecoder(rsp.Body).Decode(&gRsp); err != nil {
			return fmt.Errorf("failed to decode response body: %w", err)
		}

		current.state = gRsp.State
		if current.sessions, err = gRsp.getAttrSessions(); err != nil {
			logger.WithError(err).Info("Cannot retrieve old sessions. Ignoring...")
		}
		if v := gRsp.Attributes; v != nil {
			attributes = v
		}

	case http.StatusNotFound:
		logger.Info("Entity not found. It will be created now...")
		forceUpdate = true
		attributes["icon"] = "mdi:microphone-message"
		attributes["friendly_name"] = strings.TrimPrefix(this.conf.EntityId, "input_boolean.")

	default:
		return fmt.Errorf("unexpected status code: %d - %s", rsp.StatusCode, rsp.Status)
	}

	if !forceUpdate && target.isEqualTo(&current) {
		logger.Debug("Entity is already in requested state. No updated needed.")
		this.lastState.Swap(&current)
		return nil
	}

	attributes["editable"] = false
	sReq := statePostRequest{
		State:      ctx.State(),
		Attributes: attributes,
	}
	sReq.setAttrSessions(target.sessions)

	sReqB, err := json.Marshal(sReq)
	if err != nil {
		return err
	}

	sRsp, err := this.do("POST", "/api/states/"+this.conf.EntityId, func(req *http.Request) error {
		req.Body = io.NopCloser(bytes.NewBuffer(sReqB))
		return nil
	})
	if err != nil {
		return err
	}
	if sRsp.StatusCode != http.StatusOK && sRsp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d - %s", sRsp.StatusCode, sRsp.Status)
	}

	logger.Debug("Entity updated.")
	this.lastState.Swap(&current)

	return nil
}

func (this *Homeassistant) Initialize(conf *Configuration, saveConfFunc func() error) error {
	this.conf = conf
	this.saveConfFunc = saveConfFunc

	if err := this.Update(); err != nil {
		return err
	}

	return nil
}

func (this *Homeassistant) loadCredentials() (credentials.Credentials, error) {
	var v credentials.Credentials
	if _, err := v.ReadFromStore(); err != nil {
		return credentials.Credentials{}, err
	}

	if v.HomeAssistantServer == "" {
		v.HomeAssistantServer = this.conf.Server
	}
	if v.HomeAssistantToken == "" {
		v.HomeAssistantToken = this.conf.Token
	}

	return v, nil
}

func (this *Homeassistant) storeCredentials(cred credentials.Credentials) error {
	supported, err := cred.WriteToStore()
	if err != nil {
		return err
	}
	if supported {
		return nil
	}

	this.conf.Server = cred.HomeAssistantServer
	this.conf.Token = cred.HomeAssistantToken
	return this.saveConfFunc()
}

type resolveCredentialsReason uint

const (
	resolveCredentialsReasonDefault resolveCredentialsReason = iota
	resolveCredentialsReasonInvalidServer
	resolveCredentialsReasonInvalidToken
)

func (this *Homeassistant) resolveCredentials(reason resolveCredentialsReason) (credentials.Credentials, error) {
	fail := func(err error) (credentials.Credentials, error) {
		return credentials.Credentials{}, err
	}
	failf := func(msg string, args ...any) (credentials.Credentials, error) {
		return fail(fmt.Errorf(msg, args...))
	}

	cred, err := this.loadCredentials()
	if err != nil {
		return fail(err)
	}

	if reason == resolveCredentialsReasonDefault && cred.HomeAssistantServer != "" && cred.HomeAssistantToken != "" {
		return cred, nil
	}

	switch reason {
	case resolveCredentialsReasonInvalidServer:
		log.With("server", this.conf.Server).
			Error("Invalid Home Assistant's Server URL.")
	case resolveCredentialsReasonInvalidToken:
		log.With("server", this.conf.Server).
			Error("Invalid Home Assistant's Server URL.")
	default:
		log.Info("Server URL and long live token required to access Home Assistant.")
	}

	check := func() (serverOk, tokenOk bool, err error) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*60)
		defer cancelFunc()

		req, err := http.NewRequestWithContext(ctx, "GET", strings.TrimRight(cred.HomeAssistantServer, "/")+"/api/", nil)
		if err != nil {
			return false, false, err
		}
		req.Header.Add("Authorization", "Bearer "+cred.HomeAssistantToken)
		rsp, err := this.client.Do(req)
		if err != nil {
			return false, false, err
		}
		switch rsp.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			return true, false, nil
		case http.StatusOK:
			return true, true, nil
		default:
			return false, false, nil
		}
	}

	for {
		cred.HomeAssistantServer = ""
		cred.HomeAssistantToken = ""
		if err := common.RequestStringContentIfRequiredFromTerminal(&cred.HomeAssistantServer, fmt.Sprintf("Server URL (empty = %s)", DefaultServer), true, true); err != nil {
			return failf("cannot request server url: %w", err)
		}
		if cred.HomeAssistantServer == "" {
			cred.HomeAssistantServer = DefaultServer
		}
		if err := common.RequestStringContentIfRequiredFromTerminal(&cred.HomeAssistantToken, "Token", false, true); err != nil {
			return failf("cannot request token: %w", err)
		}

		serverOk, tokenOk, err := check()
		if err != nil {
			return fail(err)
		}
		if serverOk && tokenOk {
			if err := this.storeCredentials(cred); err != nil {
				return failf("cannot store credentials: %w", err)
			}
			return cred, nil
		}

		if !serverOk {
			log.With("server", cred.HomeAssistantServer).
				Error("Provided Home Assistant's server URL is invalid.")
		} else {
			log.With("server", cred.HomeAssistantServer).
				Error("Provided Home Assistant's long live token is invalid.")
		}
	}
}

func (this *Homeassistant) do(method, path string, cb ...func(req *http.Request) error) (rsp *http.Response, err error) {
	cred, err := this.resolveCredentials(resolveCredentialsReasonDefault)
	if err != nil {
		return nil, err
	}

	do := func() (*http.Response, error) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*60)
		defer cancelFunc()

		req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(cred.HomeAssistantServer, "/")+path, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Authorization", "Bearer "+cred.HomeAssistantToken)
		for _, cbi := range cb {
			if err := cbi(req); err != nil {
				return nil, err
			}
		}

		rsp, err = this.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to access %v: %w", req.URL, err)
		}

		return rsp, nil
	}

	for {
		rsp, err = do()
		if err != nil {
			return nil, err
		}

		switch rsp.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			if cred, err = this.resolveCredentials(resolveCredentialsReasonInvalidToken); err != nil {
				return nil, err
			}
		default:
			return rsp, nil
		}
	}
}

func (this *Homeassistant) url(path string) string {
	return strings.TrimRight(this.conf.Server, "/api") + path
}

func (this *Homeassistant) Dispose() error {
	this.conf = nil
	this.saveConfFunc = nil
	return nil
}

func (this *Homeassistant) GetType() signal.Type {
	return signal.TypeHomeAssistant
}
