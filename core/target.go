package core

/**
 * Target host and port
 */
type Target struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

/**
 * Compare to other target
 */
func (t *Target) EqualTo(other Target) bool {
	return t.Host == other.Host &&
		t.Port == other.Port
}

/**
 * Get target full address
 * host:port
 */
func (this *Target) Address() string {
	return this.Host + ":" + this.Port
}

/**
 * To String conversion
 */
func (this *Target) String() string {
	return this.Address()
}

/**
 * Next r/w operation data counters
 */
type ReadWriteCount struct {

	/* Read bytes count */
	CountRead uint

	/* Write bytes count */
	CountWrite uint

	Target Target
}

func (this ReadWriteCount) IsZero() bool {
	return this.CountRead == 0 && this.CountWrite == 0
}
