package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// EISClient wraps an AMQP connection to the EIS message bus.
type EISClient struct {
	url          string
	exchangeName string
	conn         *amqp.Connection
	channel      *amqp.Channel
	mu           sync.Mutex
	done         chan struct{}

	// reconnectSubs holds channels that are closed on each successful reconnect.
	// Consumers select on these to detect reconnect and restart their loops.
	reconnectSubsMu sync.Mutex
	reconnectSubs   []chan struct{}
}

// SubscribeReconnect returns a channel that is closed on every successful reconnect.
// The caller MUST re-subscribe after each notification.
func (c *EISClient) SubscribeReconnect() <-chan struct{} {
	c.reconnectSubsMu.Lock()
	defer c.reconnectSubsMu.Unlock()
	ch := make(chan struct{})
	c.reconnectSubs = append(c.reconnectSubs, ch)
	return ch
}

func (c *EISClient) notifyReconnect() {
	c.reconnectSubsMu.Lock()
	defer c.reconnectSubsMu.Unlock()
	for _, ch := range c.reconnectSubs {
		close(ch)
	}
	c.reconnectSubs = nil
}

// NewEISClient connects to RabbitMQ and declares the exchange.
// Returns (nil, nil) if URL is empty (graceful degradation).
func NewEISClient(url, exchangeName string) (*EISClient, error) {
	if url == "" {
		log.Println("[RMQ] URL not configured, EIS messaging disabled")
		return nil, nil
	}
	c := &EISClient{
		url:          url,
		exchangeName: exchangeName,
		done:         make(chan struct{}),
	}
	if err := c.connect(); err != nil {
		log.Printf("[RMQ] Warning: %v — EIS messaging disabled", err)
		return nil, nil
	}
	return c, nil
}

func (c *EISClient) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var conn *amqp.Connection
	var err error
	for attempt := 0; attempt < 5; attempt++ {
		conn, err = amqp.Dial(c.url)
		if err == nil {
			break
		}
		wait := time.Duration(math.Pow(2, float64(attempt))) * time.Second
		log.Printf("[RMQ] Connect attempt %d failed: %v; retry in %v", attempt+1, err, wait)
		time.Sleep(wait)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ after 5 attempts: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	if err := ch.ExchangeDeclare(c.exchangeName, "topic", true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	if c.conn != nil {
		c.conn.Close()
	}
	c.conn = conn
	c.channel = ch
	log.Println("[RMQ] Connected to RabbitMQ")
	return nil
}

// PublishJSON marshals and publishes to the exchange with the given routing key.
func (c *EISClient) PublishJSON(routingKey string, msg interface{}) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.channel.Publish(c.exchangeName, routingKey, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		})
}

// PublishRPC publishes a message to the exchange and waits for a single reply
// on a temporary queue. Returns the raw reply body or an error on timeout.
// timeout of 0 means default 15s.
func (c *EISClient) PublishRPC(routingKey string, msg interface{}, timeout time.Duration) ([]byte, error) {
	if timeout == 0 {
		timeout = 15 * time.Second
	}

	ch, err := c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open RPC channel: %w", err)
	}
	defer ch.Close()

	replyQ, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return nil, fmt.Errorf("declare reply queue: %w", err)
	}

	corrID := fmt.Sprintf("%d", time.Now().UnixNano())

	body, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	if err := ch.Publish(c.exchangeName, routingKey, false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       replyQ.Name,
			Body:          body,
		}); err != nil {
		return nil, fmt.Errorf("publish RPC: %w", err)
	}

	deliveries, err := ch.Consume(replyQ.Name, corrID, true, true, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("consume reply: %w", err)
	}

	select {
	case d, ok := <-deliveries:
		if !ok {
			return nil, fmt.Errorf("reply channel closed")
		}
		return d.Body, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("RPC timeout after %v", timeout)
	}
}

// TypedDelivery wraps a decoded message with its AMQP delivery for deferred ack.
type TypedDelivery[T any] struct {
	Msg      T
	delivery *amqp.Delivery
}

// Ack acknowledges the underlying AMQP delivery.
func (td *TypedDelivery[T]) Ack() {
	if td.delivery != nil {
		td.delivery.Ack(false)
	}
}

// Nack sends the delivery to dead-letter (or requeue if multiple is set).
func (td *TypedDelivery[T]) Nack(requeue bool) {
	if td.delivery != nil {
		td.delivery.Nack(false, requeue)
	}
}

// ConsumeJSON returns a channel of decoded messages of type T from the given queue.
// The caller MUST call td.Ack() after successful processing.
// Messages that fail to unmarshal are nacked (dead-lettered).
func ConsumeJSON[T any](ch *amqp.Channel, queue string) (<-chan TypedDelivery[T], error) {
	deliveries, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	out := make(chan TypedDelivery[T], 10)
	go func() {
		defer close(out)
		for d := range deliveries {
			var msg T
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("[RMQ] Failed to unmarshal message: %v", err)
				d.Nack(false, false)
				continue
			}
			// Caller must ack after processing — we pass the delivery along.
			out <- TypedDelivery[T]{Msg: msg, delivery: &d}
		}
	}()
	return out, nil
}

// DeclareAndBind declares a durable queue and binds it to the exchange.
// If dlxExchange is set, the queue is configured with dead-letter-exchange.
func (c *EISClient) DeclareAndBind(queueName, routingKey, dlxExchange string) (*amqp.Channel, error) {
	args := amqp.Table{}
	if dlxExchange != "" {
		args["x-dead-letter-exchange"] = dlxExchange
	}
	ch, err := c.conn.Channel()
	if err != nil {
		return nil, err
	}
	q, err := ch.QueueDeclare(queueName, true, false, false, false, args)
	if err != nil {
		ch.Close()
		return nil, err
	}
	if err := ch.QueueBind(q.Name, routingKey, c.exchangeName, false, nil); err != nil {
		ch.Close()
		return nil, err
	}
	return ch, nil
}

// Channel returns the current AMQP channel. Caller must NOT close it.
func (c *EISClient) Channel() *amqp.Channel {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.channel
}

// IsConnected returns true if the AMQP connection is alive.
func (c *EISClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil && !c.conn.IsClosed()
}

// Close shuts down the EIS client.
func (c *EISClient) Close() {
	close(c.done)
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

// SetupInfrastructure creates DLX + DLQ + all EIS queues.
// Must be called after connect(), before any consumer goroutines.
func (c *EISClient) SetupInfrastructure() error {
	dlxName := c.exchangeName + ".dlx"
	if err := c.channel.ExchangeDeclare(dlxName, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare DLX: %w", err)
	}

	// DLQ with 24h TTL
	_, err := c.channel.QueueDeclare("eis.dlq", true, false, false, false,
		amqp.Table{"x-message-ttl": 86400000})
	if err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}
	c.channel.QueueBind("eis.dlq", "#", dlxName, false, nil)

	// Verify response queue (with DLX)
	_, err = c.channel.QueueDeclare("eis.verify.response.bot", true, false, false, false,
		amqp.Table{"x-dead-letter-exchange": dlxName})
	if err != nil {
		return err
	}
	c.channel.QueueBind("eis.verify.response.bot", "irgups.verify.response", c.exchangeName, false, nil)

	// Sync queues (with DLX) — specific routing keys, NOT wildcard.
	// Wildcard irgups.sync.# would deliver all sync messages to all queues,
	// causing residents data to be silently consumed by the employees handler.
	syncBindings := map[string]string{
		"eis.sync.residents.bot":   "irgups.sync.residents",
		"eis.sync.employees.bot":   "irgups.sync.employees",
		"eis.sync.dormitories.bot": "irgups.sync.dormitories",
	}
	for queue, key := range syncBindings {
		if _, err := c.channel.QueueDeclare(queue, true, false, false, false,
			amqp.Table{"x-dead-letter-exchange": dlxName}); err != nil {
			return err
		}
		// Remove stale wildcard binding from previous versions (idempotent — no-op if not present)
		c.channel.QueueUnbind(queue, "irgups.sync.#", c.exchangeName, nil)
		// Bind exact routing key
		if err := c.channel.QueueBind(queue, key, c.exchangeName, false, nil); err != nil {
			return err
		}
	}

	log.Println("[RMQ] Infrastructure set up: exchange + DLX + DLQ + 4 queues")
	return nil
}

// StartAutoReconnect runs a background goroutine that reconnects when the connection drops.
func (c *EISClient) StartAutoReconnect(ctx context.Context) {
	if c == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if !c.IsConnected() {
					log.Println("[RMQ] Connection lost, attempting reconnect...")
					if err := c.connect(); err != nil {
						log.Printf("[RMQ] Reconnect failed: %v", err)
					} else {
						log.Println("[RMQ] Reconnected successfully")
						if err := c.SetupInfrastructure(); err != nil {
							log.Printf("[RMQ] Re-setup infrastructure failed: %v", err)
						}
						c.notifyReconnect()
					}
				}
			}
		}
	}()
}

// ConsumeWithReconnect declares a queue, starts consuming, and restarts on reconnect.
// The handler receives TypedDelivery[T]; it MUST call td.Ack() on success.
// This is a standalone function (not a method) because Go does not support generic methods.
// This function blocks until ctx is cancelled.
func ConsumeWithReconnect[T any](
	c *EISClient,
	ctx context.Context,
	queueName, routingKey, dlxExchange string,
	handler func(ctx context.Context, td TypedDelivery[T]),
) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		ch, err := c.DeclareAndBind(queueName, routingKey, dlxExchange)
		if err != nil {
			log.Printf("[RMQ] Failed to set up queue %s: %v — retrying in 5s", queueName, err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}

		deliveries, err := ConsumeJSON[T](ch, queueName)
		if err != nil {
			ch.Close()
			log.Printf("[RMQ] Failed to consume queue %s: %v — retrying in 5s", queueName, err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}

		reconnectCh := c.SubscribeReconnect()

	loop:
		for {
			select {
			case <-ctx.Done():
				ch.Close()
				return
			case <-reconnectCh:
				// Connection re-established: close old channel, restart outer loop.
				ch.Close()
				break loop
			case td, ok := <-deliveries:
				if !ok {
					// Channel closed (likely from reconnect). Restart outer loop.
					break loop
				}
				handler(ctx, td)
			}
		}
	}
}
