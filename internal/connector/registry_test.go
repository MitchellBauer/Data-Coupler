package connector_test

import (
	"testing"

	"github.com/mitchellbauer/data-coupler/internal/connector"
)

// fakeConn is a minimal Connector used only in these registry tests.
type fakeConn struct{ name string }

func (f *fakeConn) Name() string                                          { return f.name }
func (f *fakeConn) Connect(cfg connector.ConnectionConfig) error          { return nil }
func (f *fakeConn) Disconnect() error                                     { return nil }
func (f *fakeConn) Columns(_ string) ([]string, error)                    { return nil, nil }
func (f *fakeConn) Rows(_ string) (<-chan []string, error)                 { return nil, nil }

func TestRegisterGet(t *testing.T) {
	fc := &fakeConn{name: "__test_registry_fake__"}
	connector.Register(fc)

	got, ok := connector.Get("__test_registry_fake__")
	if !ok {
		t.Fatal("Get() returned false, want true")
	}
	if got.Name() != fc.name {
		t.Errorf("Get().Name() = %q, want %q", got.Name(), fc.name)
	}
}

func TestGet_NotFound(t *testing.T) {
	_, ok := connector.Get("__definitely_not_registered__")
	if ok {
		t.Error("Get() returned true for unknown connector, want false")
	}
}
