package health

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
	"unsafe"

	"water_pipe/config"
)

type Status int

const (
	StatusUnknown Status = iota
	StatusHealthy
	StatusUnhealthy
)

type Update struct {
	NodeID string
	Status Status
}

type Checker struct {
	config         config.HealthCheckConfig
	nodes          map[string]string // NodeID -> Address
	statuses       map[string]Status // NodeID -> Status
	failures       map[string]int    // NodeID -> Consecutive failures
	recoveries     map[string]int    // NodeID -> Consecutive recoveries
	subscribers    []chan<- Update
	subscribersMu  sync.Mutex
	nodesMu        sync.RWMutex
}

func NewChecker(cfg config.HealthCheckConfig) (*Checker, error) {
	return &Checker{
		config:     cfg,
		nodes:      make(map[string]string),
		statuses:   make(map[string]Status),
		failures:   make(map[string]int),
		recoveries: make(map[string]int),
	}, nil
}

func (c *Checker) RegisterNode(nodeID, address string) {
	c.nodesMu.Lock()
	defer c.nodesMu.Unlock()

	c.nodes[nodeID] = address
	c.statuses[nodeID] = StatusUnknown
	c.failures[nodeID] = 0
	c.recoveries[nodeID] = 0
}

func (c *Checker) UnregisterNode(nodeID string) {
	c.nodesMu.Lock()
	defer c.nodesMu.Unlock()

	delete(c.nodes, nodeID)
	delete(c.statuses, nodeID)
	delete(c.failures, nodeID)
	delete(c.recoveries, nodeID)
}

func (c *Checker) Start(ctx context.Context) {
	go c.run(ctx)
}

func (c *Checker) Subscribe() <-chan Update {
	c.subscribersMu.Lock()
	defer c.subscribersMu.Unlock()

	ch := make(chan Update, 10)
	c.subscribers = append(c.subscribers, ch)
	return ch
}

func (c *Checker) Unsubscribe(ch <-chan Update) {
	c.subscribersMu.Lock()
	defer c.subscribersMu.Unlock()

	chPtr := uintptr((*[2]unsafe.Pointer)(unsafe.Pointer(&ch))[1])
	
	for i, sub := range c.subscribers {
		subPtr := uintptr((*[2]unsafe.Pointer)(unsafe.Pointer(&sub))[1])
		
		if subPtr == chPtr {
			c.subscribers = append(c.subscribers[:i], c.subscribers[i+1:]...)
			close(sub)
			break
		}
	}
}

func (c *Checker) run(ctx context.Context) {
	checkTicker := time.NewTicker(c.config.Interval)
	defer checkTicker.Stop()
	
	logInterval := 30 * time.Second
	if c.config.LogInterval > 0 {
		logInterval = c.config.LogInterval
	}
	logTicker := time.NewTicker(logInterval)
	defer logTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-checkTicker.C:
			c.checkAll(ctx)
		case <-logTicker.C:
			c.logNodeHealth()
		}
	}
}

func (c *Checker) logNodeHealth() {
	c.nodesMu.RLock()
	defer c.nodesMu.RUnlock()
	
	if len(c.nodes) == 0 {
		return
	}
	
	fmt.Println("=== Next Hop Node Health Status ===")
	fmt.Println("Node ID\t\tAddress\t\t\tStatus\t\tLast Check")
	fmt.Println("---------------------------------------------------------------")
	
	for nodeID, address := range c.nodes {
		status := "Unknown"
		switch c.statuses[nodeID] {
		case StatusHealthy:
			status = "Healthy"
		case StatusUnhealthy:
			status = "Unhealthy"
		}
		
		lastCheck := time.Now().Format("2006-01-02 15:04:05")
		
		fmt.Printf("%s\t%s\t%s\t%s\n", nodeID, address, status, lastCheck)
	}
	fmt.Println("===============================================================")
}

func (c *Checker) checkAll(ctx context.Context) {
	c.nodesMu.RLock()
	nodes := make(map[string]string, len(c.nodes))
	for id, addr := range c.nodes {
		nodes[id] = addr
	}
	c.nodesMu.RUnlock()

	for id, addr := range nodes {
		go c.checkNode(ctx, id, addr)
	}
}

func (c *Checker) checkNode(ctx context.Context, nodeID, address string) {
	checkCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	var healthy bool
	switch c.config.Method {
	case "tcp_connect":
		healthy = c.checkTCPConnect(checkCtx, address)
	case "application":
		healthy = c.checkApplication(checkCtx, address)
	default:
		healthy = c.checkTCPConnect(checkCtx, address)
	}

	c.updateStatus(nodeID, healthy)
}

func (c *Checker) checkTCPConnect(ctx context.Context, address string) bool {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (c *Checker) checkApplication(ctx context.Context, address string) bool {
	return c.checkTCPConnect(ctx, address)
}

func (c *Checker) updateStatus(nodeID string, healthy bool) {
	c.nodesMu.Lock()
	defer c.nodesMu.Unlock()

	currentStatus, exists := c.statuses[nodeID]
	if !exists {
		return
	}

	if healthy {
		c.failures[nodeID] = 0
		c.recoveries[nodeID]++
	} else {
		c.failures[nodeID]++
		c.recoveries[nodeID] = 0
	}

	var newStatus Status
	if healthy && (currentStatus != StatusHealthy && c.recoveries[nodeID] >= c.config.RecoveryThreshold) {
		newStatus = StatusHealthy
	} else if !healthy && (currentStatus != StatusUnhealthy && c.failures[nodeID] >= c.config.FailureThreshold) {
		newStatus = StatusUnhealthy
	} else {
		return
	}

	c.statuses[nodeID] = newStatus

	update := Update{
		NodeID: nodeID,
		Status: newStatus,
	}
	c.notifySubscribers(update)
}

func (c *Checker) notifySubscribers(update Update) {
	c.subscribersMu.Lock()
	defer c.subscribersMu.Unlock()

	for _, sub := range c.subscribers {
		select {
		case sub <- update:
		default:
		}
	}
}
