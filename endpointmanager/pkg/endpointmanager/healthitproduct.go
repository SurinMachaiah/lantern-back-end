package endpointmanager

import (
	"context"

	"time"

	"github.com/pkg/errors"

	"github.com/google/go-cmp/cmp"
)

// HealthITProduct represents a health IT vendor product such as an
// EHR. This information is gathered from the Certified Health IT Products List
// (CHPL).
type HealthITProduct struct {
	ID                    int
	Name                  string
	Version               string
	Developer             string    // the name of the vendor that creates the product.
	Location              *Location // the address listed in CHPL for the Developer.
	AuthorizationStandard string    // examples: OAuth 2.0, Basic, etc.
	APISyntax             string    // the format of the information provided by the API, for example, REST, FHIR STU3, etc.
	APIURL                string    // the URL to the API documentation for the product.
	CertificationCriteria []string  // the ONC criteria that the product was certified to, for example, ["170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)"]
	CertificationStatus   string    // the ONC certification status, for example, "Active", "Retired", "Suspended by ONC", etc.
	CertificationDate     time.Time
	CertificationEdition  string // the product's certification edition for the ONC Health IT certification program, for example, "2014", "2015".
	LastModifiedInCHPL    time.Time
	CHPLID                string // the product's unique ID within the CHPL system.
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// HealthITProductStore is the interface for interacting with the storage layer that holds
// health IT product objects.
type HealthITProductStore interface {
	GetHealthITProduct(int) (*HealthITProduct, error)
	GetHealthITProductUsingNameAndVersion(string, string) (*HealthITProduct, error)

	AddHealthITProduct(*HealthITProduct) error
	UpdateHealthITProduct(*HealthITProduct) error
	DeleteHealthITProduct(*HealthITProduct) error

	Close()
}

type HealthITProductStoreWithContext struct {
	HealthITProductStore
}

// Equal checks each field of the two HealthITProducts except for the database ID, CHPL ID, CreatedAt and UpdatedAt fields to see if they are equal.
func (hitp *HealthITProduct) Equal(hitp2 *HealthITProduct) bool {
	if hitp == nil && hitp2 == nil {
		return true
	} else if hitp == nil {
		return false
	} else if hitp2 == nil {
		return false
	}

	if hitp.Name != hitp2.Name {
		return false
	}
	if hitp.Version != hitp2.Version {
		return false
	}
	if hitp.Developer != hitp2.Developer {
		return false
	}
	if !hitp.Location.Equal(hitp2.Location) {
		return false
	}
	if hitp.AuthorizationStandard != hitp2.AuthorizationStandard {
		return false
	}
	if hitp.APISyntax != hitp2.APISyntax {
		return false
	}
	if hitp.APIURL != hitp2.APIURL {
		return false
	}
	if !cmp.Equal(hitp.CertificationCriteria, hitp2.CertificationCriteria) {
		return false
	}
	if hitp.CertificationStatus != hitp2.CertificationStatus {
		return false
	}
	if !hitp.CertificationDate.Equal(hitp2.CertificationDate) {
		return false
	}
	if hitp.CertificationEdition != hitp2.CertificationEdition {
		return false
	}
	if !hitp.LastModifiedInCHPL.Equal(hitp2.LastModifiedInCHPL) {
		return false
	}

	return true
}

func (hitp *HealthITProduct) Update(hitp2 *HealthITProduct) error {
	if hitp == nil || hitp2 == nil {
		return errors.New("HealthITPrdouct.Update: a given health IT product is nil")
	}

	hitp.Name = hitp2.Name
	hitp.Version = hitp2.Version
	hitp.Developer = hitp2.Developer
	hitp.Location = hitp2.Location
	hitp.AuthorizationStandard = hitp2.AuthorizationStandard
	hitp.APISyntax = hitp2.APISyntax
	hitp.APIURL = hitp2.APIURL
	hitp.CertificationCriteria = hitp2.CertificationCriteria
	hitp.CertificationStatus = hitp2.CertificationStatus
	hitp.CertificationDate = hitp2.CertificationDate
	hitp.CertificationEdition = hitp2.CertificationEdition
	hitp.LastModifiedInCHPL = hitp2.LastModifiedInCHPL
	hitp.CHPLID = hitp2.CHPLID

	return nil
}

func (hitpStore *HealthITProductStoreWithContext) GetHealthITProductWithContext(ctx context.Context, id int) (*HealthITProduct, error) {
	type result struct {
		hitp *HealthITProduct
		err  error
	}

	if ctx.Err() != nil {
		return nil, NewErrOpCanceled(ctx.Err())
	}

	ch := make(chan result)
	go func() {
		hitp, err := hitpStore.GetHealthITProduct(id)
		ch <- result{hitp: hitp, err: err}
	}()

	select {
	case <-ctx.Done():
		<-ch
		return nil, NewErrOpCompleted(ctx.Err())
	case res := <-ch:
		return res.hitp, res.err
	}
}

func (hitpStore *HealthITProductStoreWithContext) GetHealthITProductUsingNameAndVersionWithContext(ctx context.Context, name string, version string) (*HealthITProduct, error) {
	type result struct {
		hitp *HealthITProduct
		err  error
	}

	if ctx.Err() != nil {
		return nil, NewErrOpCanceled(ctx.Err())
	}

	ch := make(chan result)
	go func() {
		hitp, err := hitpStore.GetHealthITProductUsingNameAndVersion(name, version)
		ch <- result{hitp: hitp, err: err}
	}()

	select {
	case <-ctx.Done():
		<-ch
		return nil, NewErrOpCompleted(ctx.Err())
	case res := <-ch:
		return res.hitp, res.err
	}
}

func (hitpStore *HealthITProductStoreWithContext) AddHealthITProductWithContext(ctx context.Context, hitp *HealthITProduct) error {
	if ctx.Err() != nil {
		return NewErrOpCanceled(ctx.Err())
	}

	ch := make(chan error)
	go func() { ch <- hitpStore.AddHealthITProduct(hitp) }()

	select {
	case <-ctx.Done():
		<-ch // wait for add to complete
		return NewErrOpCompleted(ctx.Err())
	case err := <-ch:
		return err
	}
}

func (hitpStore *HealthITProductStoreWithContext) UpdateHealthITProductWithContext(ctx context.Context, hitp *HealthITProduct) error {
	if ctx.Err() != nil {
		return NewErrOpCanceled(ctx.Err())
	}

	ch := make(chan error)
	go func() { ch <- hitpStore.UpdateHealthITProduct(hitp) }()

	select {
	case <-ctx.Done():
		<-ch
		return NewErrOpCompleted(ctx.Err())
	case err := <-ch:
		return err
	}
}

func (hitpStore *HealthITProductStoreWithContext) DeleteHealthITProductWithContext(ctx context.Context, hitp *HealthITProduct) error {
	if ctx.Err() != nil {
		return NewErrOpCanceled(ctx.Err())
	}

	ch := make(chan error)
	go func() { ch <- hitpStore.DeleteHealthITProduct(hitp) }()

	select {
	case <-ctx.Done():
		<-ch
		return NewErrOpCompleted(ctx.Err())
	case err := <-ch:
		return err
	}
}
