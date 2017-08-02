package upstream

import (
	"log"
	"time"

	"github.com/millken/tcpwder/config"
	"github.com/millken/tcpwder/core"
)

/**
 * Create new Upstream based on strategy
 */
func New(cfg config.Upstream) *Upstream {
	d := Upstream{
		opts: UpstreamOpts{0},
		cfg:  cfg,
	}

	return &d
}

/**
 * Options for pull discovery
 */
type UpstreamOpts struct {
	RetryWaitDuration time.Duration
}

/**
 * Upstream
 */
type Upstream struct {

	/**
	 * Cached backends
	 */
	backends *[]core.Backend

	/**
	 * Options for fetch
	 */
	opts UpstreamOpts

	/**
	 * Upstream configuration
	 */
	cfg config.Upstream

	/**
	 * Channel where to push newly discovered backends
	 */
	out chan ([]core.Backend)
}

/**
 * Pull / fetch backends loop
 */
func (this *Upstream) Start() {

	log.Printf("[INFO] Starting upstream")
	this.out = make(chan []core.Backend)

	// Prepare interval
	interval, err := time.ParseDuration("0")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			backends, err := this.fetch()

			if err != nil {
				log.Printf("[ERROR] %s %s %s", err, " retrying in ", this.opts.RetryWaitDuration.String())

				this.backends = &[]core.Backend{}
				this.out <- *this.backends

				time.Sleep(this.opts.RetryWaitDuration)
				continue
			}

			// cache
			this.backends = backends

			// out
			this.out <- *this.backends

			// exit gorouting if no cacheTtl
			// used for static discovery
			if interval == 0 {
				return
			}

			time.Sleep(interval)
		}
	}()
}

func (this *Upstream) fetch() (*[]core.Backend, error) {
	var backends []core.Backend
	for _, s := range this.cfg {
		backend, err := core.ParseBackendDefault(s)
		if err != nil {
			log.Printf("[WARN] %s", err)
			continue
		}
		backends = append(backends, *backend)
	}

	return &backends, nil
}

/**
 * Stop discovery
 */
func (this *Upstream) Stop() {
	// TODO: Add stopping function
}

/**
 * Returns backends channel
 */
func (this *Upstream) Discover() <-chan []core.Backend {
	return this.out
}
