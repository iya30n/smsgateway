package rabbitmq

import (
	"fmt"
	"net"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	Exchange    = "sms.exchange"
	DLXExchange = "dlx.exchange"

	QueueExpress = "sms.express"
	QueueNormal  = "sms.normal"
	QueueDLQ     = "sms.dlq"

	RouteExpress = "express"
	RouteNormal  = "normal"
	RouteFailed  = "failed"
)

type Adapter struct {
	config Config
	conn   *amqp.Connection
}

func New(config Config) (*Adapter, error) {
	conn, err := amqp.DialConfig(config.URL(), amqp.Config{
		Vhost:     config.Vhost,
		Heartbeat: config.Heartbeat,
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, config.DialTimeout)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: connect: %w", err)
	}

	a := &Adapter{config, conn}

	if err := a.setup(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq: setup: %w", err)
	}

	return a, nil
}

func (a *Adapter) setup() error {
	ch, err := a.conn.Channel()
	if err != nil {
		return fmt.Errorf("open setup channel: %w", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(Exchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare %s: %w", Exchange, err)
	}

	// DLX exchange
	if err := ch.ExchangeDeclare(DLXExchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare %s: %w", DLXExchange, err)
	}

	// main queues with DLX
	for _, q := range []string{QueueExpress, QueueNormal} {
		args := amqp.Table{
			"x-dead-letter-exchange":    DLXExchange,
			"x-dead-letter-routing-key": RouteFailed,
		}

		if _, err := ch.QueueDeclare(q, true, false, false, false, args); err != nil {
			return fmt.Errorf("declare %s: %w", q, err)
		}
	}

	// bind queues to the exchange
	if err := ch.QueueBind(QueueExpress, RouteExpress, Exchange, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(QueueNormal, RouteNormal, Exchange, false, nil); err != nil {
		return err
	}

	// DLQ
	if _, err := ch.QueueDeclare(QueueDLQ, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(QueueDLQ, RouteFailed, DLXExchange, false, nil); err != nil {
		return err
	}

    return nil
}

func (a *Adapter) Close() error {
	if a.conn != nil {
		return a.conn.Close()
	}

	return nil
}
