package main // import "github.com/boivie/lovebeat"

import (
	"flag"
	"fmt"
	"github.com/boivie/lovebeat/alert"
	"github.com/boivie/lovebeat/algorithms"
	"github.com/boivie/lovebeat/api"
	"github.com/boivie/lovebeat/backend"
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/dashboard"
	"github.com/boivie/lovebeat/eventbus"
	"github.com/boivie/lovebeat/eventlog"
	"github.com/boivie/lovebeat/metrics"
	"github.com/boivie/lovebeat/notify"
	"github.com/boivie/lovebeat/service"
	slog "github.com/boivie/lovebeat/syslog"
	"github.com/boivie/lovebeat/websocket"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
)

var (
	debug       = flag.Bool("debug", false, "Enable debug logs")
	showVersion = flag.Bool("version", false, "Print version string")
	cfgFile     = flag.String("config", "/etc/lovebeat.cfg", "Configuration file")
	cfgDir      = flag.String("config-dir", "/etc/lovebeat.conf.d", "Configuration directory")
	useSyslog   = flag.Bool("syslog", false, "Log to syslog instead of stderr")
	validate    = flag.Bool("validate-auto", false, "Evaluate auto-algorithm")
)

var log = logging.MustGetLogger("lovebeat")
var version string
var signalchan = make(chan os.Signal, 1)
var sigQuitChan = make(chan os.Signal, 1)

func signalHandler(be backend.Backend) {
	for {
		select {
		case sig := <-signalchan:
			fmt.Printf("!! Caught signal %v... shutting down\n", sig)
			be.Sync()
			return
		case <-sigQuitChan:
			buf := make([]byte, 1<<20)
			stacklen := runtime.Stack(buf, true)
			fmt.Printf("=== received SIGQUIT ===\n*** goroutine dump...\n%s\n*** end\n", buf[:stacklen])
		}
	}
}

func httpServer(cfg *config.ConfigBind, bus *eventbus.EventBus) {
	rtr := mux.NewRouter()
	api.AddEndpoints(rtr)
	websocket.Register(rtr, bus)
	dashboard.Register(rtr)
	http.Handle("/", rtr)
	log.Infof("HTTP listening on %s\n", cfg.Listen)
	http.ListenAndServe(cfg.Listen, nil)
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Sprintf("unknown_%d", os.Getpid())
	}
	return strings.Split(hostname, ".")[0]
}

func main() {
	flag.Parse()

	if *debug {
		logging.SetLevel(logging.DEBUG, "lovebeat")
	} else {
		logging.SetLevel(logging.INFO, "lovebeat")
	}
	if *useSyslog {
		slog.Init()
	} else {
		format := logging.MustStringFormatter("%{level} %{message}")
		logging.SetFormatter(format)
	}
	log.Debug("Debug logs enabled")

	if *validate {
		fmt.Fprintf(os.Stderr, "Validating auto-algorithm from stdin\n")
		algorithms.Validate()
		return
	}

	versionStr := fmt.Sprintf("lovebeat v%s (built w/%s)", version, runtime.Version())
	if *showVersion {
		fmt.Println(versionStr)
		return
	}

	wd, _ := os.Getwd()
	myName := getHostname()
	log.Info(versionStr)
	log.Infof("Started on %s, PID %d, running from %s", myName, os.Getpid(), wd)

	cfg := config.ReadConfig(*cfgFile, *cfgDir)
	bus := eventbus.New()

	eventlog.Init(cfg, bus)

	notifier := notify.Init(myName, cfg.Notify)
	alerter := alert.Init(cfg, notifier)

	m := metrics.New(&cfg.Metrics)
	service.RegisterMetrics(bus, m)

	be := backend.NewFileBackend(&cfg.Database, m, notifier)
	svcs := service.NewServices(be, bus, alerter, cfg, notifier)

	signal.Notify(signalchan, syscall.SIGTERM)
	signal.Notify(signalchan, os.Interrupt)
	signal.Notify(sigQuitChan, syscall.SIGQUIT)

	api.Init(svcs)

	go httpServer(&cfg.Http, bus)
	go api.UdpListener(&cfg.Udp)
	go api.TcpListener(&cfg.Tcp)

	m.IncCounter("started.count")

	signalHandler(be)
}
