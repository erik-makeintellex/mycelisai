package exchange

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func scanChannels(rows *sql.Rows) ([]Channel, error) {
	out := []Channel{}
	for rows.Next() {
		channel, err := scanChannelFromRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *channel)
	}
	return out, rows.Err()
}

func scanThreads(rows *sql.Rows) ([]Thread, error) {
	out := []Thread{}
	for rows.Next() {
		thread, err := scanThreadFromRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *thread)
	}
	return out, rows.Err()
}

func scanItems(rows *sql.Rows) ([]ExchangeItem, error) {
	out := []ExchangeItem{}
	for rows.Next() {
		item, err := scanItemFromRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func scanChannelRow(row *sql.Row) (*Channel, error) {
	var channel Channel
	var participantsJSON, reviewersJSON []byte
	if err := row.Scan(&channel.ID, &channel.Name, &channel.Type, &channel.Owner, &participantsJSON, &reviewersJSON, &channel.SchemaID, &channel.RetentionPolicy, &channel.Visibility, &channel.SensitivityClass, &channel.Description, &channel.Metadata, &channel.CreatedAt); err != nil {
		return nil, err
	}
	_ = json.Unmarshal(participantsJSON, &channel.Participants)
	_ = json.Unmarshal(reviewersJSON, &channel.Reviewers)
	return &channel, nil
}

func scanChannelFromRows(rows *sql.Rows) (*Channel, error) {
	var channel Channel
	var participantsJSON, reviewersJSON []byte
	if err := rows.Scan(&channel.ID, &channel.Name, &channel.Type, &channel.Owner, &participantsJSON, &reviewersJSON, &channel.SchemaID, &channel.RetentionPolicy, &channel.Visibility, &channel.SensitivityClass, &channel.Description, &channel.Metadata, &channel.CreatedAt); err != nil {
		return nil, fmt.Errorf("scan exchange channel: %w", err)
	}
	_ = json.Unmarshal(participantsJSON, &channel.Participants)
	_ = json.Unmarshal(reviewersJSON, &channel.Reviewers)
	return &channel, nil
}

func scanThreadFromRows(rows *sql.Rows) (*Thread, error) {
	var thread Thread
	var participantsJSON, reviewersJSON, escalationJSON []byte
	if err := rows.Scan(&thread.ID, &thread.ChannelID, &thread.ChannelName, &thread.ThreadType, &thread.Title, &thread.Status, &participantsJSON, &reviewersJSON, &escalationJSON, &thread.ContinuityKey, &thread.CreatedBy, &thread.Metadata, &thread.CreatedAt, &thread.UpdatedAt); err != nil {
		return nil, fmt.Errorf("scan exchange thread: %w", err)
	}
	_ = json.Unmarshal(participantsJSON, &thread.Participants)
	_ = json.Unmarshal(reviewersJSON, &thread.AllowedReviewers)
	_ = json.Unmarshal(escalationJSON, &thread.EscalationRights)
	return &thread, nil
}

func scanItemRow(row *sql.Row) (*ExchangeItem, error) {
	var item ExchangeItem
	var consumersJSON []byte
	if err := row.Scan(&item.ID, &item.ChannelID, &item.ChannelName, &item.SchemaID, &item.Payload, &item.CreatedBy, &item.AddressedTo, &item.ThreadID, &item.Visibility, &item.SensitivityClass, &item.SourceRole, &item.SourceTeam, &item.TargetRole, &item.TargetTeam, &consumersJSON, &item.CapabilityID, &item.TrustClass, &item.ReviewRequired, &item.Metadata, &item.Summary, &item.CreatedAt); err != nil {
		return nil, err
	}
	_ = json.Unmarshal(consumersJSON, &item.AllowedConsumers)
	return &item, nil
}

func scanItemFromRows(rows *sql.Rows) (*ExchangeItem, error) {
	var item ExchangeItem
	var consumersJSON []byte
	if err := rows.Scan(&item.ID, &item.ChannelID, &item.ChannelName, &item.SchemaID, &item.Payload, &item.CreatedBy, &item.AddressedTo, &item.ThreadID, &item.Visibility, &item.SensitivityClass, &item.SourceRole, &item.SourceTeam, &item.TargetRole, &item.TargetTeam, &consumersJSON, &item.CapabilityID, &item.TrustClass, &item.ReviewRequired, &item.Metadata, &item.Summary, &item.CreatedAt); err != nil {
		return nil, fmt.Errorf("scan exchange item: %w", err)
	}
	_ = json.Unmarshal(consumersJSON, &item.AllowedConsumers)
	return &item, nil
}
