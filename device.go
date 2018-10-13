package main

import (
	"errors"
	"strconv"

	"github.com/eternal-flame-AD/go-pushbullet"
)

func locateDeviceWithClient(devicenot string, client *pushbullet.Client) (string, error) {
	devices, err := client.Devices()
	if err != nil {
		return "", errors.New("An error occured while fetching devices: " + err.Error())
	}
	if res, err := strconv.Atoi(devicenot); err == nil && len(devices) > res {
		return devices[res].Iden, nil
	}
	for _, dev := range devices {
		if dev.Nickname == devicenot {
			return dev.Iden, nil
		}
	}
	for _, dev := range devices {
		if dev.Model == devicenot {
			return dev.Iden, nil
		}
	}
	for _, dev := range devices {
		if dev.Iden == devicenot {
			return dev.Iden, nil
		}
	}
	return "", errors.New("Failed to locate device identified by " + devicenot)
}
