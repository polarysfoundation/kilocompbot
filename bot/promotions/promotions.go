package promotions

import (
	"errors"
	"fmt"
	"sync"
)

const (
	mediaDefaultPath  = "assets/media/"
	defaultAdName     = "\n***AD SPACE***\n"
	defaultButtonName = "ADVERTISE HERE"
	defaultButtonLink = "https://t.me/KiloTonCoin"
)

var (
	errEmptyString      = errors.New("error: empty param")
	errInvalidTimestamp = errors.New("error: marca de tiempo invalida")
)

type Params struct {
	Media      string
	AdName     string
	ButtonName string
	ButtonLink string
	Timestamp  int64
	mutex      sync.RWMutex
}

func InitParams() *Params {
	return &Params{
		Media:      fmt.Sprintf(mediaDefaultPath+"%s", "default.mp4"),
		AdName:     defaultAdName,
		ButtonName: defaultButtonName,
		ButtonLink: defaultButtonLink,
		Timestamp:  0,
	}
}

func (p *Params) UpdateMedia(mediaName string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if mediaName == "" {
		return errEmptyString
	}

	p.Media = fmt.Sprintf(mediaDefaultPath+"%s", mediaName)

	return nil
}

func (p *Params) UpdateAdName(adName string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if adName == "" {
		return errEmptyString
	}

	p.AdName = adName

	return nil
}

func (p *Params) UpdateButtonName(buttonName string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if buttonName == "" {
		return errEmptyString
	}

	p.ButtonName = buttonName

	return nil
}

func (p *Params) UpdateButtonLink(buttonLink string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if buttonLink == "" {
		return errEmptyString
	}

	p.ButtonLink = buttonLink

	return nil
}

func (p *Params) SetTimestamp(timestamp int64) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if timestamp == 0 {
		return errInvalidTimestamp
	}

	p.Timestamp = timestamp

	return nil
}
