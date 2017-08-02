package core

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

/**
 * Backend means upstream server
 * with all needed associate information
 */
type Backend struct {
	Target
	Priority int          `json:"priority"`
	Weight   int          `json:"weight"`
	Sni      string       `json:"sni,omitempty"`
	Stats    BackendStats `json:"stats"`
}

/**
 * Backend status
 */
type BackendStats struct {
	Live               bool   `json:"live"`
	TotalConnections   int64  `json:"total_connections"`
	ActiveConnections  uint   `json:"active_connections"`
	RefusedConnections uint64 `json:"refused_connections"`
	RxBytes            uint64 `json:"rx"`
	TxBytes            uint64 `json:"tx"`
	RxSecond           uint   `json:"rx_second"`
	TxSecond           uint   `json:"tx_second"`
}

const (
	DEFAULT_BACKEND_PATTERN = `^(?P<host>\S+):(?P<port>\d+)(\sweight=(?P<weight>\d+))?(\spriority=(?P<priority>\d+))?(\ssni=(?P<sni>[^\s]+))?$`
)

/**
 * Do parding of backend line with default pattern
 */
func ParseBackendDefault(line string) (*Backend, error) {
	return ParseBackend(line, DEFAULT_BACKEND_PATTERN)
}

/**
 * Do parsing of backend line
 */
func ParseBackend(line string, pattern string) (*Backend, error) {

	//trim string
	line = strings.TrimSpace(line)

	// parse string by regexp
	var reg = regexp.MustCompile(pattern)
	match := reg.FindStringSubmatch(line)

	if len(match) == 0 {
		return nil, errors.New("Cant parse " + line)
	}

	result := make(map[string]string)

	// get named capturing groups
	for i, name := range reg.SubexpNames() {
		if name != "" {
			result[name] = match[i]
		}
	}

	weight, err := strconv.Atoi(result["weight"])
	if err != nil {
		weight = 1
	}

	priority, err := strconv.Atoi(result["priority"])
	if err != nil {
		priority = 1
	}

	backend := Backend{
		Target: Target{
			Host: result["host"],
			Port: result["port"],
		},
		Weight:   weight,
		Sni:      result["sni"],
		Priority: priority,
		Stats: BackendStats{
			Live: true,
		},
	}

	return &backend, nil
}

/**
 * Check if backend equal to another
 */
func (this *Backend) EqualTo(other Backend) bool {
	return this.Target.EqualTo(other.Target)
}

/**
 * Merge another backend to this one
 */
func (this *Backend) MergeFrom(other Backend) *Backend {

	this.Priority = other.Priority
	this.Weight = other.Weight
	this.Sni = other.Sni

	return this
}

/**
 * Get backends target address
 */
func (this *Backend) Address() string {
	return this.Target.Address()
}

/**
 * String conversion
 */
func (this Backend) String() string {
	return fmt.Sprintf("{%s p=%d,w=%d,l=%t,a=%d}",
		this.Address(), this.Priority, this.Weight, this.Stats.Live, this.Stats.ActiveConnections)
}
