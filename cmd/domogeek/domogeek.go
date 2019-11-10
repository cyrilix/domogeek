package main

import (
	"domogeek/calendar"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	cal          calendar.Calendar
	calCounter   *prometheus.CounterVec
	calSummary   *prometheus.SummaryVec
	calHistogram *prometheus.HistogramVec
)

func init() {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		log.Fatalf("unable to load time location: %v", err)
	}
	cal = calendar.Calendar{Location: loc}

	calCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "domogeek",
		Subsystem: "calendar",
		Name:      "request_total",
		Help:      "Total request to calendar service",
	},
		[]string{
			"code",
			"method",
		})

	calSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "domogeek",
		Subsystem: "calendar",
		Name:      "summary",
		Help:      "Calendar request summary",
	},
		nil)
	calHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "domogeek",
		Subsystem: "calendar",
		Name:      "histogram",
		Help:      "Request duration histogram",
	},
		nil)
}

type CalendarDay struct {
	Day        time.Time `json:"day"`
	WorkingDay bool      `json:"working_day"`
	Ferie      bool      `json:"ferie"`
	Holiday    bool      `json:"holiday"`
	Weekday    bool      `json:"weekday"`
}

type CalendarHandler struct{}

func (c *CalendarHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	cd := CalendarDay{
		Day:        now,
		WorkingDay: cal.IsWorkingDay(now),
		Ferie:      cal.IsHoliday(now),
		Holiday:    cal.IsHoliday(now),
		Weekday:    cal.IsWeekDay(now),
	}

	content, err := json.Marshal(cd)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("unable to marshall response %v, %v", content, err)
	} else {
		_, err = w.Write(content)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("unable to marshall response %v, :%v", content, err)
		}
	}
}

func main() {
	var port int
	var host string

	flag.StringVar(&host, "host", "", "host to listen, default all addresses")
	flag.IntVar(&port, "port", 8080, "port to listen")
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("start server on %s", addr)

	h := promhttp.InstrumentHandlerDuration(
		calHistogram,
		promhttp.InstrumentHandlerDuration(
			calSummary,
			promhttp.InstrumentHandlerCounter(
				calCounter,
				&CalendarHandler{})))
	http.Handle("/calendar", &h)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(addr, nil))
}
