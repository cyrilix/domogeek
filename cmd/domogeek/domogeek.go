package main

import (
	"context"
	"domogeek/pkg/calendar"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/hellofresh/health-go/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	cal          *calendar.Calendar
	location     *time.Location
	calCounter   *prometheus.CounterVec
	calSummary   *prometheus.SummaryVec
	calHistogram *prometheus.HistogramVec
)

func init() {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		zap.S().Fatalf("unable to load time location: %v", err)
	}
	location = loc

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

func (c *CalendarHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	now := time.Now()
	calDavHolidays, err := cal.IsHolidaysFromCaldav(now)
	if err != nil {
		zap.S().Warnf("unable to read holliday status from caldav: %v", err)
		calDavHolidays = false
	}

	cd := CalendarDay{
		Day:        now,
		WorkingDay: cal.IsWorkingDay(now),
		Ferie:      cal.IsHoliday(now),
		Holiday:    calDavHolidays,
		Weekday:    cal.IsWeekDay(now),
	}

	content, err := json.Marshal(cd)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		zap.S().Errorf("unable to marshall response %v, %v", content, err)
	} else {
		_, err = w.Write(content)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			zap.S().Errorf("unable to marshall response %v, :%v", content, err)
		}
	}
}

func main() {
	var port int
	var host string
	var caldavUrl, caldavPath, caldavSummaryPattern string

	flag.StringVar(&host, "host", "", "host to listen, default all addresses")
	flag.IntVar(&port, "port", 8080, "port to listen")
	flag.StringVar(&caldavUrl, "caldav-url", "", "caldav url to use to read holidays events")
	flag.StringVar(&caldavPath, "caldav-path", "", "caldav path to use to read holidays events")
	flag.StringVar(&caldavSummaryPattern, "caldav-summary-pattern", "Holidays", "Summary pattern that matches holidays event")
	flag.Parse()

	logLevel := zap.LevelFlag("log", zap.InfoLevel, "log level")
	flag.Parse()

	if len(os.Args) <= 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(*logLevel)
	lgr, err := config.Build()
	if err != nil {
		log.Fatalf("unable to init logger: %v", err)
	}
	defer func() {
		_ = lgr.Sync()
	}()
	zap.ReplaceGlobals(lgr)

	cdav, err := calendar.NewCaldav(caldavUrl, caldavPath)
	if err != nil {
		zap.S().Fatal("unable to init caldav instance")
	}
	cal = calendar.New(location,
		calendar.WithCaldav(cdav),
		calendar.WithCaldavPath(caldavPath),
		calendar.WithCaldavSummaryPattern(caldavSummaryPattern),
	)

	addr := fmt.Sprintf("%s:%d", host, port)
	zap.S().Infof("start server on %s", addr)

	h := promhttp.InstrumentHandlerDuration(
		calHistogram,
		promhttp.InstrumentHandlerDuration(
			calSummary,
			promhttp.InstrumentHandlerCounter(
				calCounter,
				&CalendarHandler{})))
	http.Handle("/calendar", &h)
	http.Handle("/metrics", promhttp.Handler())
	healthz, _ := health.New(health.WithChecks(health.Config{
		Name:      "calendar",
		Timeout:   time.Second * 5,
		SkipOnErr: false,
		Check: func(ctx context.Context) error {
			return nil
		},
	}),
		health.WithChecks(health.Config{
			Name:      "caldav",
			Timeout:   5 * time.Second,
			SkipOnErr: false,
			Check: func(ctx context.Context) error {
				_, err := cal.IsHolidaysFromCaldav(time.Now())
				return err
			},
		}),
	)
	http.Handle("/status", healthz.Handler())

	signChan := make(chan os.Signal, 1)
	go func() {
		zap.S().Fatal(http.ListenAndServe(addr, nil))
	}()

	signal.Notify(signChan, syscall.SIGTERM)
	<-signChan
	zap.S().Info("exit on sigterm")
}
