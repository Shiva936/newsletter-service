package providers

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AsyncBatchManager handles batching for providers that don't support true bulk
type AsyncBatchManager struct {
	provider       EmailProviderInterface
	batchSize      int
	batch          []*EmailNotification
	batchTimeout   time.Duration
	mutex          sync.Mutex
	ticker         *time.Ticker
	stopChan       chan struct{}
	processingChan chan bool
}

// NewAsyncBatchManager creates a new async batch manager
func NewAsyncBatchManager(provider EmailProviderInterface, batchSize int, batchTimeout time.Duration) *AsyncBatchManager {
	manager := &AsyncBatchManager{
		provider:       provider,
		batchSize:      batchSize,
		batch:          make([]*EmailNotification, 0, batchSize),
		batchTimeout:   batchTimeout,
		stopChan:       make(chan struct{}),
		processingChan: make(chan bool, 1),
	}

	// Start the timeout ticker
	manager.ticker = time.NewTicker(batchTimeout)
	go manager.timeoutProcessor()

	return manager
}

// AddToBatch adds an email to the current batch
func (bm *AsyncBatchManager) AddToBatch(email *EmailNotification) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	// Add to batch
	bm.batch = append(bm.batch, email)

	// Check if batch is ready for processing
	if len(bm.batch) >= bm.batchSize {
		// Trigger immediate processing
		select {
		case bm.processingChan <- true:
		default:
			// Channel is full, processing already triggered
		}
	}

	return nil
}

// ProcessBatch processes the current batch
func (bm *AsyncBatchManager) ProcessBatch(ctx context.Context) error {
	bm.mutex.Lock()
	currentBatch := make([]*EmailNotification, len(bm.batch))
	copy(currentBatch, bm.batch)
	bm.batch = bm.batch[:0] // Reset batch
	bm.mutex.Unlock()

	if len(currentBatch) == 0 {
		return nil
	}

	// Check if provider supports true bulk
	if bm.provider.SupportsBulk() {
		// Convert to bulk notification
		recipients := make([]string, len(currentBatch))
		for i, email := range currentBatch {
			recipients[i] = email.To
		}

		bulkNotification := &BulkEmailNotification{
			To:      recipients,
			Subject: currentBatch[0].Subject, // Assume same subject for batch
			Body:    currentBatch[0].Body,    // Assume same body for batch
			From:    currentBatch[0].From,
		}

		return bm.provider.SendBulkEmail(ctx, bulkNotification)
	}

	// Send individually with concurrency control
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Limit concurrent sends
	errorChan := make(chan error, len(currentBatch))

	for _, email := range currentBatch {
		wg.Add(1)
		go func(e *EmailNotification) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := bm.provider.SendEmail(ctx, e); err != nil {
				errorChan <- err
			}
		}(email)
	}

	wg.Wait()
	close(errorChan)

	// Check for errors
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("batch processing failed: %d out of %d emails failed", len(errors), len(currentBatch))
	}

	return nil
}

// GetBatchSize returns current batch size
func (bm *AsyncBatchManager) GetBatchSize() int {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()
	return len(bm.batch)
}

// IsBatchReady returns true if batch is ready for processing
func (bm *AsyncBatchManager) IsBatchReady() bool {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()
	return len(bm.batch) >= bm.batchSize
}

// timeoutProcessor handles timeout-based batch processing
func (bm *AsyncBatchManager) timeoutProcessor() {
	for {
		select {
		case <-bm.ticker.C:
			// Timeout occurred, process batch if not empty
			if bm.GetBatchSize() > 0 {
				select {
				case bm.processingChan <- true:
				default:
					// Processing already triggered
				}
			}

		case <-bm.processingChan:
			// Process the batch
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := bm.ProcessBatch(ctx); err != nil {
				fmt.Printf("Batch processing error: %v\n", err)
			}
			cancel()

		case <-bm.stopChan:
			// Stop the processor
			bm.ticker.Stop()
			return
		}
	}
}

// Stop stops the batch manager and processes remaining emails
func (bm *AsyncBatchManager) Stop() error {
	close(bm.stopChan)

	// Process any remaining emails
	if bm.GetBatchSize() > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return bm.ProcessBatch(ctx)
	}

	return nil
}

// BatchedEmailProvider wraps a provider with batching capabilities
type BatchedEmailProvider struct {
	provider     EmailProviderInterface
	batchManager *AsyncBatchManager
	bulkEnabled  bool
}

// NewBatchedEmailProvider creates a provider with async batching
func NewBatchedEmailProvider(provider EmailProviderInterface, batchSize int, bulkEnabled bool) *BatchedEmailProvider {
	var batchManager *AsyncBatchManager

	if !bulkEnabled {
		// Create batch manager for non-bulk providers
		batchTimeout := 30 * time.Second // Configurable timeout
		batchManager = NewAsyncBatchManager(provider, batchSize, batchTimeout)
	}

	return &BatchedEmailProvider{
		provider:     provider,
		batchManager: batchManager,
		bulkEnabled:  bulkEnabled,
	}
}

// SendEmail sends an individual email or adds to batch
func (bp *BatchedEmailProvider) SendEmail(ctx context.Context, notification *EmailNotification) error {
	if bp.bulkEnabled || bp.batchManager == nil {
		// Send directly for bulk-enabled providers or when no batch manager
		return bp.provider.SendEmail(ctx, notification)
	}

	// Add to batch for non-bulk providers
	return bp.batchManager.AddToBatch(notification)
}

// SendBulkEmail sends bulk emails
func (bp *BatchedEmailProvider) SendBulkEmail(ctx context.Context, notification *BulkEmailNotification) error {
	if bp.bulkEnabled {
		return bp.provider.SendBulkEmail(ctx, notification)
	}

	// For non-bulk providers, add all emails to batch
	for _, recipient := range notification.To {
		email := &EmailNotification{
			To:      recipient,
			Subject: notification.Subject,
			Body:    notification.Body,
			From:    notification.From,
		}
		if err := bp.batchManager.AddToBatch(email); err != nil {
			return err
		}
	}

	return nil
}

// Delegate other methods to the underlying provider
func (bp *BatchedEmailProvider) SupportsBulk() bool {
	return bp.bulkEnabled || bp.provider.SupportsBulk()
}

func (bp *BatchedEmailProvider) GetLimits() ProviderLimits {
	return bp.provider.GetLimits()
}

func (bp *BatchedEmailProvider) GetStats() ProviderStats {
	return bp.provider.GetStats()
}

func (bp *BatchedEmailProvider) GetProviderType() EmailProvider {
	return bp.provider.GetProviderType()
}

func (bp *BatchedEmailProvider) GetProviderName() string {
	return bp.provider.GetProviderName()
}

func (bp *BatchedEmailProvider) GetPriority() int {
	return bp.provider.GetPriority()
}

func (bp *BatchedEmailProvider) IsEnabled() bool {
	return bp.provider.IsEnabled()
}

func (bp *BatchedEmailProvider) ValidateConfig() error {
	return bp.provider.ValidateConfig()
}

// Stop stops the batch manager
func (bp *BatchedEmailProvider) Stop() error {
	if bp.batchManager != nil {
		return bp.batchManager.Stop()
	}
	return nil
}
