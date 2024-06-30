package core

import (
	"errors"
	"sync"
	"time"
)

var (
	errrorEmptyID              = errors.New("error: id invalido")
	errorCompAlreadyExist      = errors.New("error: comp ya existe")
	errorBlacklistAlreadyExist = errors.New("error: blacklist ya existe")
	errorCompNotExist          = errors.New("error: comp no existe")
	errorBlacklistNotExist     = errors.New("error: blacklist no existe")
	errorTimestampAlreadyExist = errors.New("error: timestamp ya existe")
	errorTimestampNotExist     = errors.New("error: timestamp no existe")
)

type Competition struct {
	Comps     map[string]*Purchases
	BlackList map[string]*Sales
	Timestamp map[string]int64
	mutex     sync.RWMutex
}

func InitComp() *Competition {
	return &Competition{
		Comps:     make(map[string]*Purchases),
		BlackList: make(map[string]*Sales),
		Timestamp: make(map[string]int64),
	}
}

func (c *Competition) CompExist(id string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if id == "" {
		return false
	}

	_, exist := c.Comps[id]

	return exist
}

func (c *Competition) BlackListExist(id string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if id == "" {
		return false
	}

	_, exist := c.BlackList[id]

	return exist
}

func (c *Competition) TimestampExist(id string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if id == "" {
		return false
	}

	_, exist := c.Timestamp[id]

	return exist
}

func (c *Competition) GetComp(id string) (*Purchases, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if id == "" {
		return nil, errrorEmptyID
	}

	comp, exist := c.Comps[id]
	if !exist {
		return nil, errorCompNotExist
	}

	return comp, nil
}

func (c *Competition) GetTimestamp(id string) (int64, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if id == "" {
		return 0, errrorEmptyID
	}

	timestamp, exist := c.Timestamp[id]
	if !exist {
		return 0, errorTimestampNotExist
	}

	return timestamp, nil
}

func (c *Competition) GetBlacklist(id string) (*Sales, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if id == "" {
		return nil, errrorEmptyID
	}

	blacklist, exist := c.BlackList[id]
	if !exist {
		return nil, errorBlacklistNotExist
	}

	return blacklist, nil
}

func (c *Competition) NewComp(id string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if id == "" {
		return errrorEmptyID
	}

	if _, exist := c.Comps[id]; exist {
		return errorCompAlreadyExist
	}

	newComp := &Purchases{
		Purchase: make(map[string]*Purchase),
	}

	c.Comps[id] = newComp

	return nil
}

func (c *Competition) NewBlacklist(id string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if id == "" {
		return errrorEmptyID
	}

	if _, exist := c.BlackList[id]; exist {
		return errorBlacklistAlreadyExist
	}

	newSale := &Sales{
		Sale: make(map[string]*Sale),
	}

	c.BlackList[id] = newSale

	return nil
}

func (c *Competition) NewTimestamp(id string, timestamp int64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if id == "" {
		return errrorEmptyID
	}

	if _, exist := c.Timestamp[id]; exist {
		return errorTimestampAlreadyExist
	}

	c.Timestamp[id] = timestamp

	return nil
}

func (c *Competition) RemoveCompActive(id string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if id == "" {
		return errrorEmptyID
	}

	if _, exist := c.Comps[id]; !exist {
		return errorCompNotExist
	}

	delete(c.Comps, id)

	return nil
}

func (c *Competition) RemoveBlacklistActive(id string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if id == "" {
		return errrorEmptyID
	}

	if _, exist := c.BlackList[id]; !exist {
		return errorBlacklistNotExist
	}

	delete(c.BlackList, id)

	return nil
}

func (c *Competition) RemoveTimestampActive(id string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if id == "" {
		return errrorEmptyID
	}

	if _, exist := c.Timestamp[id]; !exist {
		return errorTimestampNotExist
	}

	delete(c.Timestamp, id)

	return nil
}

func (c *Competition) IsEnded(id string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	crrTime := time.Now().Unix()

	timestamp := c.Timestamp[id]

	if timestamp > crrTime {
		return false
	} else {
		return true
	}
}
