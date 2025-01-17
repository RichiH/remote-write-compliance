package cases

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/prometheus/pkg/exemplar"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/storage"
)

type Test struct {
	Name     string
	Metrics  http.Handler
	Expected Validator
}

func metricHandler(c prometheus.Collector) http.Handler {
	r := prometheus.NewPedanticRegistry()
	r.Register(c)
	return promhttp.HandlerFor(r, promhttp.HandlerOpts{})
}

func staticHandler(contents []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(contents)
	})
}

func removeLabel(ls labels.Labels, name string) labels.Labels {
	for i := 0; i < len(ls); i++ {
		if ls[i].Name == name {
			return ls[:i+copy(ls[i:], ls[i+1:])]
		}
	}
	return ls
}

type Validator func(t *testing.T, bs []Batch)

type Appendable struct {
	sync.Mutex
	Batches []Batch
}

type Batch struct {
	appender *Appendable
	samples  []sample
}

type sample struct {
	l labels.Labels
	t int64
	v float64
}

func (m *Appendable) Appender(_ context.Context) storage.Appender {
	b := &Batch{
		appender: m,
	}
	return b
}

func (m *Batch) Append(_ uint64, l labels.Labels, t int64, v float64) (uint64, error) {
	m.samples = append(m.samples, sample{l, t, v})
	return 0, nil
}

func (m *Batch) Commit() error {
	m.appender.Mutex.Lock()
	defer m.appender.Mutex.Unlock()
	m.appender.Batches = append(m.appender.Batches, *m)
	return nil
}

func (*Batch) Rollback() error {
	return nil
}

func (*Batch) AppendExemplar(ref uint64, l labels.Labels, e exemplar.Exemplar) (uint64, error) {
	return 0, nil
}
