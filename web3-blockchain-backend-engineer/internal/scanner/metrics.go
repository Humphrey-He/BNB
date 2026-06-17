package scanner

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds Prometheus metrics for the scanner
type Metrics struct {
	blocksScannedTotal   prometheus.Counter
	eventsExtractedTotal prometheus.Counter
	batchSizeGauge       prometheus.Gauge
	scanLagGauge         prometheus.Gauge
	lastScannedBlock     prometheus.Gauge
	errorsTotal          prometheus.Counter
	reorgsDetectedTotal  prometheus.Counter
	natsPublishTotal     prometheus.Counter
	natsPublishErrors    prometheus.Counter
}

// NewMetrics creates a new Metrics instance with Prometheus collectors
func NewMetrics(namespace string) *Metrics {
	return &Metrics{
		blocksScannedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "scanner_blocks_scanned_total",
			Help:      "Total number of blocks scanned",
		}),
		eventsExtractedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "scanner_events_extracted_total",
			Help:      "Total number of events extracted",
		}),
		batchSizeGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "scanner_batch_size",
			Help:      "Current batch size for scanning",
		}),
		scanLagGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "scanner_lag_blocks",
			Help:      "Number of blocks behind the chain head",
		}),
		lastScannedBlock: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "scanner_last_scanned_block",
			Help:      "Last scanned block number",
		}),
		errorsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "scanner_errors_total",
			Help:      "Total number of scanner errors",
		}),
		reorgsDetectedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "scanner_reorgs_detected_total",
			Help:      "Total number of reorgs detected",
		}),
		natsPublishTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "scanner_nats_publish_total",
			Help:      "Total number of messages published to NATS",
		}),
		natsPublishErrors: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "scanner_nats_publish_errors_total",
			Help:      "Total number of NATS publish errors",
		}),
	}
}

// IncBlocksScanned increments the blocks scanned counter
func (m *Metrics) IncBlocksScanned(count uint64) {
	m.blocksScannedTotal.Add(float64(count))
}

// IncEventsExtracted increments the events extracted counter
func (m *Metrics) IncEventsExtracted(count uint64) {
	m.eventsExtractedTotal.Add(float64(count))
}

// SetBatchSize sets the current batch size
func (m *Metrics) SetBatchSize(size uint64) {
	m.batchSizeGauge.Set(float64(size))
}

// SetScanLag sets the current scan lag in blocks
func (m *Metrics) SetScanLag(lag uint64) {
	m.scanLagGauge.Set(float64(lag))
}

// SetLastScannedBlock sets the last scanned block number
func (m *Metrics) SetLastScannedBlock(block uint64) {
	m.lastScannedBlock.Set(float64(block))
}

// IncErrors increments the errors counter
func (m *Metrics) IncErrors() {
	m.errorsTotal.Inc()
}

// IncReorgsDetected increments the reorgs detected counter
func (m *Metrics) IncReorgsDetected() {
	m.reorgsDetectedTotal.Inc()
}

// IncNatsPublish increments the NATS publish counter
func (m *Metrics) IncNatsPublish(count uint64) {
	m.natsPublishTotal.Add(float64(count))
}

// IncNatsPublishError increments the NATS publish errors counter
func (m *Metrics) IncNatsPublishError() {
	m.natsPublishErrors.Inc()
}
