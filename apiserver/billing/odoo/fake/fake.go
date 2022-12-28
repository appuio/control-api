package fake

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"

	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	"github.com/appuio/control-api/apiserver/billing/odoo"
)

const (
	IDStart = 2345
	IDStep  = 2
)

type fakeOdooStorage struct {
	storeLock sync.RWMutex
	store     map[string]*billingv1.BillingEntity

	// idCounter is used to generate unique IDs for the fake storage.
	// Access it using the nextID() method.
	idCounter uint64
}

var _ odoo.OdooStorage = &fakeOdooStorage{}

func NewFakeOdooStorage() odoo.OdooStorage {
	return &fakeOdooStorage{
		store:     make(map[string]*billingv1.BillingEntity),
		idCounter: IDStart - IDStep,
	}
}

func (s *fakeOdooStorage) Create(ctx context.Context, be *billingv1.BillingEntity) error {
	s.storeLock.Lock()
	defer s.storeLock.Unlock()

	id := formatID(s.nextID())

	be.ObjectMeta = metav1.ObjectMeta{
		Name: id,
	}

	s.store[id] = be.DeepCopy()

	return nil
}

func (s *fakeOdooStorage) Get(ctx context.Context, name string) (*billingv1.BillingEntity, error) {
	s.storeLock.RLock()
	defer s.storeLock.RUnlock()

	be, ok := s.store[name]
	if !ok {
		return nil, odoo.ErrNotFound
	}

	return be.DeepCopy(), nil
}

func (s *fakeOdooStorage) Update(ctx context.Context, be *billingv1.BillingEntity) error {
	s.storeLock.Lock()
	defer s.storeLock.Unlock()

	if _, ok := s.store[be.Name]; !ok {
		return odoo.ErrNotFound
	}

	be.ObjectMeta = metav1.ObjectMeta{
		Name: be.Name,
	}

	s.store[be.Name] = be.DeepCopy()

	return nil
}

func (s *fakeOdooStorage) List(ctx context.Context) ([]billingv1.BillingEntity, error) {
	s.storeLock.RLock()
	defer s.storeLock.RUnlock()

	var list []billingv1.BillingEntity
	for _, be := range s.store {
		list = append(list, *be.DeepCopy())
	}

	slices.SortFunc(list, func(a, b billingv1.BillingEntity) bool {
		return a.Name < b.Name
	})

	return list, nil
}

func (s *fakeOdooStorage) nextID() uint64 {
	return atomic.AddUint64(&s.idCounter, 2)
}

func formatID(id uint64) string {
	return "be-" + strconv.FormatUint(id, 10)
}
