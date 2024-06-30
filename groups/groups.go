package groups

import (
	"errors"
	"log"
	"sync"

	"github.com/polarysfoundation/kilocompbot/indexer"
)

var (
	errorEmptyID      = errors.New("error: el id es invalido, no puede ser vacio")
	errorAlreadyExist = errors.New("error: el temp o grupo ya existe")
	errorNoExist      = errors.New("error: el grupo o temp no existe")
	errGroupExist     = errors.New("error: el grupo ya existe")
	errGettingPools   = errors.New("error: hubo un error obteniendo los pools")
)

type GroupData struct {
	ID            string
	CompActive    bool
	JettonAddress string
	Dedust        string
	StonFi        string
	Emoji         string
}

type Groups struct {
	ActiveGroups map[string]*GroupData
	mutex        sync.RWMutex
}

func InitGroups() *Groups {
	return &Groups{
		ActiveGroups: make(map[string]*GroupData),
	}
}

func (g *Groups) GetDataGroup(id string) (*GroupData, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	group, exist := g.ActiveGroups[id]
	if !exist {
		return nil, errorNoExist
	}

	return group, nil
}

func (g *Groups) UpdateCompStatus(chatID string, update bool) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if _, exist := g.ActiveGroups[chatID]; !exist {
		return errorNoExist
	}

	group := g.ActiveGroups[chatID]

	group.CompActive = update

	return nil
}

func (g *Groups) CompStatus(id string) bool {
	return g.ActiveGroups[id].CompActive
}

func (g *Groups) GroupExist(id string) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if _, exist := g.ActiveGroups[id]; !exist {
		return false
	}
	return true
}

func (g *Groups) AddPools(id string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	pools := indexer.InitPools()

	if id == "" {
		return errorEmptyID
	}

	if _, exist := g.ActiveGroups[id]; !exist {
		return errorNoExist
	}

	group := g.ActiveGroups[id]

	err := pools.GetPools(group.JettonAddress)
	if err != nil {
		return errGettingPools
	}

	if pools.Dedust != "" {
		group.Dedust = pools.Dedust
	}

	if pools.StonFi != "" {
		group.StonFi = pools.StonFi
	}

	log.Printf("dedust: %s", pools.Dedust)
	log.Printf("stonfi: %s", pools.StonFi)

	return nil

}

func (g *Groups) AddGroup(id string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if id == "" {
		return errorEmptyID
	}

	if _, exist := g.ActiveGroups[id]; exist {
		return errGroupExist
	}

	newGroup := &GroupData{
		ID:            id,
		CompActive:    false,
		JettonAddress: "",
		Dedust:        "",
		StonFi:        "",
	}

	g.ActiveGroups[id] = newGroup

	return nil
}

type TempSetter struct {
	AwaitingToken     bool
	AwaitingEmoji     bool
	AwaitingTimestamp bool
}

type ActiveTemps struct {
	TempSetter map[string]*TempSetter
	mutex      sync.RWMutex
}

func InitTemp() *ActiveTemps {
	return &ActiveTemps{
		TempSetter: make(map[string]*TempSetter),
	}
}

func (t *ActiveTemps) GetActiveTemps(id string) (*TempSetter, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	temp, exist := t.TempSetter[id]
	if !exist {
		return nil, errorNoExist
	}

	return temp, nil
}

func (t *ActiveTemps) AddTemp(id string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if id == "" {
		return errorEmptyID
	}

	if _, exist := t.TempSetter[id]; exist {
		return errorAlreadyExist
	}

	newTemp := &TempSetter{
		AwaitingToken:     false,
		AwaitingEmoji:     false,
		AwaitingTimestamp: false,
	}

	t.TempSetter[id] = newTemp

	return nil
}

func (t *ActiveTemps) ChangeTemp(typeTemp int, id string, update bool) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if id == "" {
		return errorEmptyID
	}

	temp, exist := t.TempSetter[id]
	if !exist {
		return errorNoExist
	}

	if typeTemp == 0 {
		temp.AwaitingToken = update
		return nil
	}

	if typeTemp == 1 {
		temp.AwaitingEmoji = update
		return nil
	}

	if typeTemp == 2 {
		temp.AwaitingTimestamp = update
		return nil
	}

	return nil
}
