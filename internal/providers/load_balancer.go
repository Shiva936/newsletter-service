package providers

import (
	"sort"
	"sync/atomic"
)

// RoundRobinLoadBalancer implements round-robin load balancing
type RoundRobinLoadBalancer struct {
	counter uint64
}

// NewRoundRobinLoadBalancer creates a new round-robin load balancer
func NewRoundRobinLoadBalancer() LoadBalancer {
	return &RoundRobinLoadBalancer{}
}

// SelectProvider selects a provider using round-robin strategy
func (lb *RoundRobinLoadBalancer) SelectProvider(providers []EmailProviderInterface, emailCount int) EmailProviderInterface {
	if len(providers) == 0 {
		return nil
	}

	// Filter healthy providers
	healthy := make([]EmailProviderInterface, 0)
	for _, provider := range providers {
		if provider.GetStats().IsHealthy {
			healthy = append(healthy, provider)
		}
	}

	if len(healthy) == 0 {
		return providers[0] // Fallback to first provider
	}

	index := atomic.AddUint64(&lb.counter, 1) % uint64(len(healthy))
	return healthy[index]
}

// DistributeLoad distributes emails evenly across providers
func (lb *RoundRobinLoadBalancer) DistributeLoad(providers []EmailProviderInterface, emails []EmailNotification) map[EmailProviderInterface][]EmailNotification {
	distribution := make(map[EmailProviderInterface][]EmailNotification)

	if len(providers) == 0 || len(emails) == 0 {
		return distribution
	}

	// Filter healthy providers
	healthy := make([]EmailProviderInterface, 0)
	for _, provider := range providers {
		if provider.GetStats().IsHealthy {
			healthy = append(healthy, provider)
		}
	}

	if len(healthy) == 0 {
		return distribution
	}

	// Distribute emails round-robin
	for i, email := range emails {
		provider := healthy[i%len(healthy)]
		distribution[provider] = append(distribution[provider], email)
	}

	return distribution
}

// WeightedLoadBalancer implements priority/weight-based load balancing
type WeightedLoadBalancer struct{}

// NewWeightedLoadBalancer creates a new weighted load balancer
func NewWeightedLoadBalancer() LoadBalancer {
	return &WeightedLoadBalancer{}
}

// SelectProvider selects a provider based on priority and capacity
func (lb *WeightedLoadBalancer) SelectProvider(providers []EmailProviderInterface, emailCount int) EmailProviderInterface {
	if len(providers) == 0 {
		return nil
	}

	// Sort by priority (lower number = higher priority)
	sorted := make([]EmailProviderInterface, len(providers))
	copy(sorted, providers)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GetPriority() < sorted[j].GetPriority()
	})

	// Find first healthy provider with capacity
	for _, provider := range sorted {
		stats := provider.GetStats()
		limits := provider.GetLimits()

		if stats.IsHealthy &&
			stats.EmailsSentLastHour < limits.MaxEmailsPerHour &&
			stats.CurrentLoad < 90 { // Not overloaded
			return provider
		}
	}

	// Fallback to highest priority provider
	return sorted[0]
}

// DistributeLoad distributes based on provider capacity and priority
func (lb *WeightedLoadBalancer) DistributeLoad(providers []EmailProviderInterface, emails []EmailNotification) map[EmailProviderInterface][]EmailNotification {
	distribution := make(map[EmailProviderInterface][]EmailNotification)

	if len(providers) == 0 || len(emails) == 0 {
		return distribution
	}

	// Sort by priority
	sorted := make([]EmailProviderInterface, len(providers))
	copy(sorted, providers)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GetPriority() < sorted[j].GetPriority()
	})

	// Distribute emails based on available capacity
	remainingEmails := emails
	for _, provider := range sorted {
		if len(remainingEmails) == 0 {
			break
		}

		stats := provider.GetStats()
		limits := provider.GetLimits()

		if !stats.IsHealthy {
			continue
		}

		// Calculate how many emails this provider can handle
		availableCapacity := limits.MaxEmailsPerHour - stats.EmailsSentLastHour
		if availableCapacity <= 0 {
			continue
		}

		// Take up to available capacity or remaining emails
		takeCount := min(availableCapacity, len(remainingEmails))
		distribution[provider] = remainingEmails[:takeCount]
		remainingEmails = remainingEmails[takeCount:]
	}

	// If there are still remaining emails, distribute to first available provider
	if len(remainingEmails) > 0 && len(sorted) > 0 {
		firstProvider := sorted[0]
		distribution[firstProvider] = append(distribution[firstProvider], remainingEmails...)
	}

	return distribution
}

// LeastLoadLoadBalancer selects provider with least current load
type LeastLoadLoadBalancer struct{}

// NewLeastLoadBalancer creates a new least-load balancer
func NewLeastLoadBalancer() LoadBalancer {
	return &LeastLoadLoadBalancer{}
}

// SelectProvider selects provider with lowest current load
func (lb *LeastLoadLoadBalancer) SelectProvider(providers []EmailProviderInterface, emailCount int) EmailProviderInterface {
	if len(providers) == 0 {
		return nil
	}

	var bestProvider EmailProviderInterface
	lowestLoad := 101 // Higher than max possible load (100%)

	for _, provider := range providers {
		stats := provider.GetStats()
		if stats.IsHealthy && stats.CurrentLoad < lowestLoad {
			lowestLoad = stats.CurrentLoad
			bestProvider = provider
		}
	}

	if bestProvider != nil {
		return bestProvider
	}

	// Fallback to first provider
	return providers[0]
}

// DistributeLoad distributes emails to least loaded providers first
func (lb *LeastLoadLoadBalancer) DistributeLoad(providers []EmailProviderInterface, emails []EmailNotification) map[EmailProviderInterface][]EmailNotification {
	distribution := make(map[EmailProviderInterface][]EmailNotification)

	if len(providers) == 0 || len(emails) == 0 {
		return distribution
	}

	// Sort by current load (ascending)
	sorted := make([]EmailProviderInterface, len(providers))
	copy(sorted, providers)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GetStats().CurrentLoad < sorted[j].GetStats().CurrentLoad
	})

	// Distribute emails starting with least loaded
	emailsPerProvider := len(emails) / len(sorted)
	remainder := len(emails) % len(sorted)

	start := 0
	for i, provider := range sorted {
		if !provider.GetStats().IsHealthy {
			continue
		}

		count := emailsPerProvider
		if i < remainder {
			count++ // Distribute remainder emails
		}

		if start+count <= len(emails) {
			distribution[provider] = emails[start : start+count]
			start += count
		}
	}

	return distribution
}

// Helper function for minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
