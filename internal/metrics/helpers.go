package metrics

import "time"

func ObserveOperation(service, method string, start time.Time, err error) {
	RequestsTotal.WithLabelValues(service, method).Inc()
	RequestDuration.WithLabelValues(service, method).Observe(time.Since(start).Seconds())
	if err != nil {
		// choose a code label, e.g. "error"
		ErrorsTotal.WithLabelValues(service, method, "error").Inc()
	}
}
