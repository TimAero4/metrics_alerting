package handler

import (
	"html/template"
	"metrics_alerting/internal/model"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// MetricServiceProvider определяет интерфейс, необходимый хендлеру
type MetricServiceProvider interface {
	ProcessUpdate(metricType model.MetricType, name, value string) error
	GetMetric(metricType model.MetricType, name string) (string, bool)
	GetAllMetrics() (map[string]float64, map[string]int64)
}

const htmlTemplate = `
<html><body>
<h1>Gauge Metrics</h1><ul>
{{range $name, $val := .Gauges}}<li>{{$name}}: {{$val}}</li>{{end}}
</ul>
<h1>Counter Metrics</h1><ul>
{{range $name, $val := .Counters}}<li>{{$name}}: {{$val}}</li>{{end}}
</ul>
</body></html>
`

type Handler struct {
	svc MetricServiceProvider
	tpl *template.Template
}

func NewHandler(svc MetricServiceProvider) *Handler {
	tpl := template.Must(template.New("metrics").Parse(htmlTemplate))
	return &Handler{svc: svc, tpl: tpl}
}

func (h *Handler) UpdateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricTypeStr := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")
		metricValue := chi.URLParam(r, "value")

		if metricName == "" || metricValue == "" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		metricType := model.MetricType(metricTypeStr)
		if metricType != model.Gauge && metricType != model.Counter {
			http.Error(w, "Incorrect metric type", http.StatusBadRequest)
			return
		}

		if err := h.svc.ProcessUpdate(metricType, metricName, metricValue); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func (h *Handler) ValueHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricTypeStr := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")

		if metricName == "" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		metricType := model.MetricType(metricTypeStr)
		val, ok := h.svc.GetMetric(metricType, metricName)
		if !ok {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(val))
	}
}

func (h *Handler) ListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gauges, counters := h.svc.GetAllMetrics()

		gaugeStr := make(map[string]string)
		for k, v := range gauges {
			gaugeStr[k] = strconv.FormatFloat(v, 'f', -1, 64)
		}

		counterStr := make(map[string]string)
		for k, v := range counters {
			counterStr[k] = strconv.FormatInt(v, 10)
		}

		data := struct {
			Gauges   map[string]string
			Counters map[string]string
		}{
			Gauges:   gaugeStr,
			Counters: counterStr,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := h.tpl.Execute(w, data); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}
