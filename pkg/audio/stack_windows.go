package audio

import (
	"fmt"
	"sync"

	"github.com/go-ole/go-ole"
	"github.com/moutend/go-wca/pkg/wca"
)

type Stack struct {
	initialized bool
	mutex       sync.RWMutex
}

func (this *Stack) Initialize() error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.initialized {
		return nil
	}

	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		return fmt.Errorf("failed to initialize ole: %v", err)
	}

	this.initialized = true
	return nil
}

func (this *Stack) Dispose() error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if !this.initialized {
		return nil
	}

	ole.CoUninitialize()
	this.initialized = false

	return nil
}

func (this *Stack) FindDevices() (Devices, error) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	if !this.initialized {
		return nil, fmt.Errorf("not initialized")
	}

	var de *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &de); err != nil {
		return nil, fmt.Errorf("cannot ceate IMMDeviceEnumerator instance: %w", err)
	}
	defer de.Release()

	return this.introspectDevicesOf(de)
}

func (this *Stack) introspectDevicesOf(enumerator *wca.IMMDeviceEnumerator) (result Devices, _ error) {
	var collection *wca.IMMDeviceCollection
	if err := enumerator.EnumAudioEndpoints(wca.ECapture, wca.DEVICE_STATE_ACTIVE, &collection); err != nil {
		return nil, fmt.Errorf("cannot query IMMDevices: %w", err)
	}
	defer collection.Release()

	var count uint32
	if err := collection.GetCount(&count); err != nil {
		return nil, fmt.Errorf("cannot get count of IMMDevice collection: %w", err)
	}

	for i := uint32(0); i < count; i++ {
		device, err := this.introspectDeviceOf(collection, i)
		if err != nil {
			return nil, err
		}
		result = append(result, device)
	}

	return
}

func (this *Stack) introspectDeviceOf(collection *wca.IMMDeviceCollection, deviceIndex uint32) (Device, error) {
	var device *wca.IMMDevice
	if err := collection.Item(deviceIndex, &device); err != nil {
		return Device{}, fmt.Errorf("cannot get item %d of IMMDevice collection: %w", deviceIndex, err)
	}
	defer device.Release()

	return this.introspectDevice(device, deviceIndex)
}

func (this *Stack) introspectDevice(captureDevice *wca.IMMDevice, deviceIndex uint32) (Device, error) {
	var propertyStore *wca.IPropertyStore
	if err := captureDevice.OpenPropertyStore(wca.STGM_READ, &propertyStore); err != nil {
		return Device{}, fmt.Errorf("cannot get properties of device %d of IMMDevice collection: %w", deviceIndex, err)
	}
	defer propertyStore.Release()

	var name wca.PROPVARIANT
	if err := propertyStore.GetValue(&wca.PKEY_Device_FriendlyName, &name); err != nil {
		return Device{}, fmt.Errorf("cannot get name of device %d of IMMDevice collection: %w", deviceIndex, err)
	}

	var sessionManager *wca.IAudioSessionManager2
	if err := captureDevice.Activate(wca.IID_IAudioSessionManager2, wca.CLSCTX_ALL, nil, &sessionManager); err != nil {
		return Device{}, fmt.Errorf("cannot get session for device %d of IMMDevice collection: %w", deviceIndex, err)
	}
	defer sessionManager.Release()

	device := Device{
		Name:  name.String(),
		Index: deviceIndex,
	}

	if sessions, err := device.getSessionsOfDevice(sessionManager, &device); err != nil {
		return Device{}, err
	} else {
		device.Sessions = sessions
	}

	return device, nil
}
