package observability

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type labelKey string

type Registry struct {
	mu         sync.RWMutex
	counters   map[string]float64
	gauges     map[string]float64
	histograms map[string]histogram
}

type histogram struct {
	sum   float64
	count float64
}

var defaultRegistry = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{
		counters:   map[string]float64{},
		gauges:     map[string]float64{},
		histograms: map[string]histogram{},
	}
}

func Default() *Registry { return defaultRegistry }

func IncCounter(name string, value float64, labels map[string]string) {
	defaultRegistry.IncCounter(name, value, labels)
}

func SetGauge(name string, value float64, labels map[string]string) {
	defaultRegistry.SetGauge(name, value, labels)
}

func ObserveHistogram(name string, value float64, labels map[string]string) {
	defaultRegistry.ObserveHistogram(name, value, labels)
}

func (r *Registry) IncCounter(name string, value float64, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counters[metricID(name, labels)] += value
}

func (r *Registry) SetGauge(name string, value float64, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.gauges[metricID(name, labels)] = value
}

func (r *Registry) ObserveHistogram(name string, value float64, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := metricID(name, labels)
	entry := r.histograms[id]
	entry.sum += value
	entry.count++
	r.histograms[id] = entry
}

func (r *Registry) RenderPrometheus(service string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var builder strings.Builder
	renderMap := func(metricType string, values map[string]float64) {
		keys := make([]string, 0, len(values))
		for key := range values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			name, labels := splitMetricID(key)
			builder.WriteString(fmt.Sprintf("# TYPE %s %s\n", name, metricType))
			builder.WriteString(fmt.Sprintf("%s%s %v\n", name, mergeLabels(service, labels), values[key]))
		}
	}

	renderMap("counter", r.counters)
	renderMap("gauge", r.gauges)

	histogramKeys := make([]string, 0, len(r.histograms))
	for key := range r.histograms {
		histogramKeys = append(histogramKeys, key)
	}
	sort.Strings(histogramKeys)
	for _, key := range histogramKeys {
		name, labels := splitMetricID(key)
		value := r.histograms[key]
		builder.WriteString(fmt.Sprintf("# TYPE %s_sum counter\n", name))
		builder.WriteString(fmt.Sprintf("%s_sum%s %v\n", name, mergeLabels(service, labels), value.sum))
		builder.WriteString(fmt.Sprintf("# TYPE %s_count counter\n", name))
		builder.WriteString(fmt.Sprintf("%s_count%s %v\n", name, mergeLabels(service, labels), value.count))
	}

	return builder.String()
}

func metricID(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys)+1)
	parts = append(parts, name)
	for _, key := range keys {
		parts = append(parts, key+"="+labels[key])
	}
	return strings.Join(parts, "|")
}

func splitMetricID(id string) (string, map[string]string) {
	parts := strings.Split(id, "|")
	name := parts[0]
	labels := map[string]string{}
	for _, part := range parts[1:] {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			labels[kv[0]] = kv[1]
		}
	}
	return name, labels
}

func mergeLabels(service string, labels map[string]string) string {
	all := map[string]string{"service": service}
	for key, value := range labels {
		all[key] = value
	}
	keys := make([]string, 0, len(all))
	for key := range all {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf(`%s=%q`, key, all[key]))
	}
	return "{" + strings.Join(parts, ",") + "}"
}
