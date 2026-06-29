package parser

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/nats-io/nats.go"
)

// Subject constants for NATS topics
const (
	SubjectRawEvents    = "raw_events"
	SubjectParsedEvents = "parsed_events"
)

// ERC-20 Transfer event signature hash
var ERC20TransferTopic = TransferEventSignature

// RawEventMessage represents the raw event message from NATS (must match scanner's definition)
type RawEventMessage struct {
	ChainID         int64    `json:"chain_id"`
	BlockNumber     uint64   `json:"block_number"`
	BlockHash       string   `json:"block_hash"`
	TxHash          string   `json:"tx_hash"`
	LogIndex        uint     `json:"log_index"`
	ContractAddress string   `json:"contract_address"`
	Topics          []string `json:"topics"`
	Data            string   `json:"data"`
	EventName       string   `json:"event_name,omitempty"`
	FromAddress     string   `json:"from_address,omitempty"`
	ToAddress       string   `json:"to_address,omitempty"`
	Value           string   `json:"value,omitempty"`
}

// ParsedEvent represents a parsed event after decoding
type ParsedEvent struct {
	ChainID      int64  `json:"chain_id"`
	TokenID      int64  `json:"token_id"`
	TokenAddress string `json:"token_address"`
	TxHash       string `json:"tx_hash"`
	LogIndex     uint   `json:"log_index"`
	BlockNumber  int64  `json:"block_number"`
	BlockHash    string `json:"block_hash"`
	From         string `json:"from"`
	To           string `json:"to"`
	Amount       string `json:"amount"`
	EventName    string `json:"event_name"`
}

// Parser is the main parser service that consumes raw events
type Parser struct {
	natsClient      *nats.Conn
	watchedAddrRepo repository.WatchedAddressRepository
	tokenRepo       repository.TokenRepository
	chainEventRepo  repository.ChainEventRepository
	depositRepo     repository.DepositRepository
	chainRepo       repository.ChainRepository
	logger          *slog.Logger
}

// NewParser creates a new Parser instance
func NewParser(
	natsClient *nats.Conn,
	watchedAddrRepo repository.WatchedAddressRepository,
	tokenRepo repository.TokenRepository,
	chainEventRepo repository.ChainEventRepository,
	depositRepo repository.DepositRepository,
	chainRepo repository.ChainRepository,
	logger *slog.Logger,
) *Parser {
	return &Parser{
		natsClient:      natsClient,
		watchedAddrRepo: watchedAddrRepo,
		tokenRepo:       tokenRepo,
		chainEventRepo:  chainEventRepo,
		depositRepo:     depositRepo,
		chainRepo:       chainRepo,
		logger:          logger,
	}
}

// Start begins consuming raw events from NATS and processing them
func (p *Parser) Start(ctx context.Context) error {
	sub, err := p.natsClient.SubscribeSync(SubjectRawEvents)
	if err != nil {
		p.logger.Error("Failed to subscribe to raw_events", "error", err)
		return err
	}

	p.logger.Info("Parser started, listening for raw events")

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Parser stopping due to context cancellation")
			return ctx.Err()
		default:
		}

		msg, err := sub.NextMsg(time.Second)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			p.logger.Warn("Failed to get next message", "error", err)
			continue
		}

		var event RawEventMessage
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			p.logger.Error("Failed to unmarshal raw event", "error", err)
			msg.Ack()
			continue
		}

		p.logger.Debug("Processing raw event",
			"chain_id", event.ChainID,
			"tx_hash", event.TxHash,
			"log_index", event.LogIndex,
		)

		// Decode the event
		parsed, err := p.decodeEvent(&event)
		if err != nil {
			p.logger.Warn("Failed to decode event",
				"chain_id", event.ChainID,
				"tx_hash", event.TxHash,
				"log_index", event.LogIndex,
				"error", err,
			)
			msg.Ack()
			continue
		}

		// Save chain event
		if err := p.saveChainEvent(parsed); err != nil {
			p.logger.Warn("Failed to save chain event",
				"chain_id", event.ChainID,
				"tx_hash", event.TxHash,
				"error", err,
			)
		}

		// Check if this is a deposit and save if so
		if p.isDeposit(parsed) {
			if err := p.saveDeposit(parsed); err != nil {
				p.logger.Warn("Failed to save deposit",
					"chain_id", event.ChainID,
					"tx_hash", event.TxHash,
					"error", err,
				)
			} else {
				p.logger.Info("Deposit detected and saved",
					"chain_id", event.ChainID,
					"tx_hash", event.TxHash,
					"from", parsed.From,
					"to", parsed.To,
					"amount", parsed.Amount,
				)
			}
		}

		// Publish parsed event
		if err := p.publishParsedEvent(parsed); err != nil {
			p.logger.Warn("Failed to publish parsed event",
				"chain_id", event.ChainID,
				"tx_hash", event.TxHash,
				"error", err,
			)
		}

		msg.Ack()
	}
}

// decodeEvent decodes a raw event based on its topic signature
func (p *Parser) decodeEvent(event *RawEventMessage) (*ParsedEvent, error) {
	parsed := &ParsedEvent{
		ChainID:      event.ChainID,
		TokenAddress: event.ContractAddress,
		TxHash:       event.TxHash,
		LogIndex:     event.LogIndex,
		BlockNumber:  int64(event.BlockNumber),
		BlockHash:    event.BlockHash,
		EventName:    "Unknown",
	}

	if event.EventName == "NativeTransfer" {
		parsed.From = event.FromAddress
		parsed.To = event.ToAddress
		parsed.Amount = event.Value
		parsed.EventName = "NativeTransfer"
		return parsed, nil
	}

	// Check if this is an ERC-20 Transfer event (topic0 matches)
	if len(event.Topics) >= 1 && event.Topics[0] == ERC20TransferTopic.Hex() {
		decoded, err := DecodeTransferFromRaw(event)
		if err != nil {
			return nil, err
		}
		parsed.From = decoded.From.Hex()
		parsed.To = decoded.To.Hex()
		parsed.Amount = decoded.Amount.String()
		parsed.EventName = "Transfer"
	}

	return parsed, nil
}

// saveChainEvent saves the parsed event to the chain_events table
func (p *Parser) saveChainEvent(parsed *ParsedEvent) error {
	event := &repository.ChainEvent{
		ChainID:         parsed.ChainID,
		TxHash:          parsed.TxHash,
		LogIndex:        int(parsed.LogIndex),
		BlockNumber:     parsed.BlockNumber,
		BlockHash:       parsed.BlockHash,
		ContractAddress: parsed.TokenAddress,
		EventName:       parsed.EventName,
		FromAddress:     parsed.From,
		ToAddress:       parsed.To,
		Amount:          parsed.Amount,
	}

	return p.chainEventRepo.Create(event)
}

// publishParsedEvent publishes the parsed event to NATS
func (p *Parser) publishParsedEvent(parsed *ParsedEvent) error {
	data, err := json.Marshal(parsed)
	if err != nil {
		return err
	}
	return p.natsClient.Publish(SubjectParsedEvents, data)
}
