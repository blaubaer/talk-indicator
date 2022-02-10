package audio

import (
	"fmt"
	"github.com/go-ole/go-ole"
	"github.com/moutend/go-wca/pkg/wca"
)

func findDevices() (Devices, error) {
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		panic(err) // Incorrect function.
	}
	defer ole.CoUninitialize()

	var de *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &de); err != nil {
		return nil, fmt.Errorf("cannot ceate IMMDeviceEnumerator instance: %w", err)
	}
	defer de.Release()

	return introspectDevicesOf(de)
}

func introspectDevicesOf(enumerator *wca.IMMDeviceEnumerator) (result Devices, _ error) {
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
		device, err := introspectDeviceOf(collection, i)
		if err != nil {
			return nil, err
		}
		result = append(result, device)
	}

	return
}

func introspectDeviceOf(collection *wca.IMMDeviceCollection, deviceIndex uint32) (Device, error) {
	var device *wca.IMMDevice
	if err := collection.Item(deviceIndex, &device); err != nil {
		return Device{}, fmt.Errorf("cannot get item %d of IMMDevice collection: %w", deviceIndex, err)
	}
	defer device.Release()

	return introspectDevice(device, deviceIndex)
}

func introspectDevice(captureDevice *wca.IMMDevice, deviceIndex uint32) (Device, error) {
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

	if sessions, err := device.getSessionsOfDevice(sessionManager); err != nil {
		return Device{}, err
	} else {
		device.Sessions = sessions
	}

	return device, nil
}
