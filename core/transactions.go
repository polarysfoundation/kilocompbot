package core

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"sort"
	"sync"

	"github.com/polarysfoundation/kilocompbot/getters"
	"github.com/polarysfoundation/kilocompbot/indexer"
	"golang.org/x/crypto/sha3"
)

var (
	errorEmptyEvent       = errors.New("error: evento vacio")
	errorNotExist         = errors.New("error: la operacion no existe")
	errorAlreadyExist     = errors.New("error: la operacion ya existe")
	errorSaleNotFound     = errors.New("error: venta no encontrada")
	errorPurchaseNotFound = errors.New("error: compra no encontrada")
)

type Sale struct {
	JettonAddress  string
	JettonName     string
	JettonSymbol   string
	JettonDecimals *big.Int
	Seller         string
	Ton            *big.Int
	Token          *big.Int
}

type Sales struct {
	Sale  map[string]*Sale
	mutex sync.RWMutex
}

func InitSales() *Sales {
	return &Sales{
		Sale: make(map[string]*Sale),
	}
}

func (p *Sales) StoreSale(sale *Sale) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	dataBytes, err := json.Marshal(sale)
	if err != nil {
		return err
	}

	hash := hash(dataBytes)

	p.Sale[hash] = sale

	return nil
}

func (s *Sales) GetSale(hash string) (*Sale, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	sale, exist := s.Sale[hash]
	if !exist {
		return nil, errorNotExist
	}

	return sale, nil
}

func (p *Sales) GetSaleHash(event *indexer.Event) (string, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	dataBytes, err := json.Marshal(event)
	if err != nil {
		return "", err
	}

	hash := hash(dataBytes)

	return hash, nil
}

func (s *Sales) AddSale(event *indexer.Event) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	decimals := big.NewInt(1000000000)

	if event == nil {
		return errorEmptyEvent
	}

	dataBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	hash := hash(dataBytes)

	if _, exist := s.Sale[hash]; exist {
		return errorAlreadyExist
	}

	seller, err := getters.GetAddress(event.Wallet)
	if err != nil {
		return err
	}

	jettonAddr, err := getters.GetAddress(event.JettonAddress)
	if err != nil {
		return err
	}

	jettonDecimal := convertToLargeInt(event.JettonDecimals)

	newSale := &Sale{
		JettonAddress:  jettonAddr,
		JettonName:     event.JettonName,
		JettonSymbol:   event.JettonSymbol,
		JettonDecimals: event.JettonDecimals,
		Seller:         seller,
		Ton:            new(big.Int).Div(event.TonOut, decimals),
		Token:          new(big.Int).Div(event.TokenIn, jettonDecimal),
	}

	s.Sale[hash] = newSale

	return nil
}

func (s *Sales) RemoveSale(hash string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exist := s.Sale[hash]; !exist {
		return errorSaleNotFound
	}

	delete(s.Sale, hash)

	return nil
}

type Purchase struct {
	JettonAddress  string
	JettonName     string
	JettonSymbol   string
	JettonDecimals *big.Int
	Buyer          string
	Ton            *big.Int
	Token          *big.Int
}

type Purchases struct {
	Purchase map[string]*Purchase
	mutex    sync.RWMutex
}

func InitPurchases() *Purchases {
	return &Purchases{
		Purchase: make(map[string]*Purchase),
	}
}

func (p *Purchases) StorePurchase(purchase *Purchase) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	dataBytes, err := json.Marshal(purchase)
	if err != nil {
		return err
	}

	hash := hash(dataBytes)

	p.Purchase[hash] = purchase

	return nil

}

func (p *Purchases) GetPurchase(hash string) (*Purchase, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	purchase, exist := p.Purchase[hash]
	if !exist {
		return nil, errorNotExist
	}

	return purchase, nil
}

func (p *Purchases) GetPurchaseHash(event *indexer.Event) (string, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	dataBytes, err := json.Marshal(event)
	if err != nil {
		return "", err
	}

	hash := hash(dataBytes)

	return hash, nil
}

func (p *Purchases) GetCompList() []*Purchase {
	purchases := make([]*Purchase, 0)

	for _, order := range p.Purchase {
		purchases = append(purchases, order)
	}

	sort.Slice(purchases, func(i, j int) bool {
		return purchases[i].Ton.Cmp(purchases[j].Ton) > 0
	})

	if len(purchases) > 10 {
		return purchases[:10]
	}

	return purchases
}

func (p *Purchases) AddPurchase(event *indexer.Event) (*Purchase, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	decimals := big.NewInt(1000000000)

	if event == nil {
		return nil, errorEmptyEvent
	}

	dataBytes, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	hash := hash(dataBytes)

	if _, exist := p.Purchase[hash]; exist {
		return nil, errorAlreadyExist
	}

	buyer, err := getters.GetAddress(event.Wallet)
	if err != nil {
		return nil, err
	}

	jettonAddr, err := getters.GetAddress(event.JettonAddress)
	if err != nil {
		return nil, err
	}

	jettonDecimal := convertToLargeInt(event.JettonDecimals)

	newPurchase := &Purchase{
		JettonAddress:  jettonAddr,
		JettonName:     event.JettonName,
		JettonSymbol:   event.JettonSymbol,
		JettonDecimals: event.JettonDecimals,
		Buyer:          buyer,
		Ton:            new(big.Int).Div(event.TonIn, decimals),
		Token:          new(big.Int).Div(event.TokenOut, jettonDecimal),
	}

	p.Purchase[hash] = newPurchase

	return newPurchase, nil
}

func (p *Purchases) GetPurchaseAndHashByBuyer(buyer string) (*Purchase, string, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for hash, purchase := range p.Purchase {
		if buyer == purchase.Buyer {
			return purchase, hash, nil
		}
	}

	return nil, "", errorCompNotExist
}

func (s *Purchases) RemovePurchase(hash string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exist := s.Purchase[hash]; !exist {
		return errorPurchaseNotFound
	}

	delete(s.Purchase, hash)

	return nil
}

/*************** Internal Functions ***************/

func hash(data []byte) string {
	hash := sha3.Sum256(data)
	hex := hex.EncodeToString(hash[:])
	return hex
}

func convertToLargeInt(value *big.Int) *big.Int {
	result := new(big.Int)
	result.SetInt64(1)

	numZeros := value.Int64()
	result.Mul(result, new(big.Int).Exp(big.NewInt(10), big.NewInt(numZeros), nil))

	return result
}
