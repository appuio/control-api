package fake

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apitypes "k8s.io/apimachinery/pkg/types"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo"
)

const (
	IDStart = 2345
	IDStep  = 2
)

type fakeOdooStorage struct {
	metadataSupport bool

	storeLock sync.RWMutex
	store     map[string]*billingv1.BillingEntity

	// idCounter is used to generate unique IDs for the fake storage.
	// Access it using the nextID() method.
	idCounter uint64
}

var _ odoo.OdooStorage = &fakeOdooStorage{}

func NewFakeOdooStorage(metadataSupport bool) odoo.OdooStorage {
	return &fakeOdooStorage{
		store:     make(map[string]*billingv1.BillingEntity),
		idCounter: IDStart - IDStep,
	}
}

func (s *fakeOdooStorage) Create(ctx context.Context, be *billingv1.BillingEntity) error {
	s.storeLock.Lock()
	defer s.storeLock.Unlock()

	id := formatID(s.nextID())

	be.Name = id
	be.UID = apitypes.UID(uuid.NewString())

	s.cleanMetadata(be)

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

	s.cleanMetadata(be)

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

// cleanMetadata simulate first naive implementation of the Odoo storage.
func (s *fakeOdooStorage) cleanMetadata(be *billingv1.BillingEntity) {
	meta := metav1.ObjectMeta{
		Name: be.Name,
		// Without UID patch requests fail with a 404 error.
		UID: be.UID,
	}
	if s.metadataSupport {
		meta = metav1.ObjectMeta{
			Name:            be.Name,
			UID:             be.UID,
			ResourceVersion: be.ResourceVersion,
			Annotations:     be.Annotations,
			Labels:          be.Labels,
		}
	}
	be.ObjectMeta = meta
}

func formatID(id uint64) string {
	return "be-" + strconv.FormatUint(id, 10)
}
