package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

const apiEndpoint = "https://api.domeneshop.no/v0"

type Provider struct {
	APIToken  string `json:"api_token,omitempty"`
	APISecret string `json:"api_secret,omitempty"`
}

type domain struct {
	ID     int    `json:"id"`
	Domain string `json:"domain"`
}

type dnsRecord struct {
	ID   int    `json:"id"`
	Host string `json:"host"`
	Type string `json:"type"`
	Data string `json:"data"`
	TTL  int    `json:"ttl"`
}

func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	domainID, err := p.getDomainID(ctx, zone)
	if err != nil {
		return nil, err
	}
	records, err := p.getDNSRecords(ctx, domainID)
	if err != nil {
		return nil, err
	}
	var libRecords []libdns.Record
	for _, record := range records {
		libRecords = append(libRecords, libdns.RR{
			Type: record.Type,
			Name: record.Host,
			Data: record.Data,
			TTL:  time.Duration(record.TTL) * time.Second,
		})
	}
	return libRecords, nil
}

func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	domainID, err := p.getDomainID(ctx, zone)
	if err != nil {
		return nil, err
	}
	var created []libdns.Record
	for _, record := range records {
		rr := record.RR()
		rec := dnsRecord{
			Host: libdns.RelativeName(rr.Name, zone),
			Type: rr.Type,
			Data: rr.Data,
			TTL:  int(rr.TTL.Seconds()),
		}
		if rec.TTL == 0 {
			rec.TTL = 3600
		}
		err := p.createDNSRecord(ctx, domainID, rec)
		if err != nil {
			return created, err
		}
		created = append(created, record)
	}
	return created, nil
}

func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	domainID, err := p.getDomainID(ctx, zone)
	if err != nil {
		return nil, err
	}
	existing, err := p.getDNSRecords(ctx, domainID)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		rr := record.RR()
		name := libdns.RelativeName(rr.Name, zone)
		for _, existingRec := range existing {
			if existingRec.Host == name && existingRec.Type == rr.Type {
				err := p.deleteDNSRecord(ctx, domainID, existingRec.ID)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return p.AppendRecords(ctx, zone, records)
}

func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	domainID, err := p.getDomainID(ctx, zone)
	if err != nil {
		return nil, err
	}
	existing, err := p.getDNSRecords(ctx, domainID)
	if err != nil {
		return nil, err
	}
	var deleted []libdns.Record
	for _, record := range records {
		rr := record.RR()
		name := libdns.RelativeName(rr.Name, zone)
		for _, existingRec := range existing {
			if existingRec.Host == name && existingRec.Type == rr.Type && existingRec.Data == rr.Data {
				err := p.deleteDNSRecord(ctx, domainID, existingRec.ID)
				if err != nil {
					return deleted, err
				}
				deleted = append(deleted, record)
				break
			}
		}
	}
	return deleted, nil
}

func (p *Provider) getDomainID(ctx context.Context, zone string) (int, error) {
	zone = strings.TrimSuffix(zone, ".")
	req, err := http.NewRequestWithContext(ctx, "GET", apiEndpoint+"/domains", nil)
	if err != nil {
		return 0, err
	}
	req.SetBasicAuth(p.APIToken, p.APISecret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API request failed: %s", resp.Status)
	}
	var domains []domain
	if err := json.NewDecoder(resp.Body).Decode(&domains); err != nil {
		return 0, err
	}
	for _, d := range domains {
		if d.Domain == zone {
			return d.ID, nil
		}
	}
	return 0, fmt.Errorf("domain %s not found", zone)
}

func (p *Provider) getDNSRecords(ctx context.Context, domainID int) ([]dnsRecord, error) {
	url := fmt.Sprintf("%s/domains/%d/dns", apiEndpoint, domainID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(p.APIToken, p.APISecret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed: %s", resp.Status)
	}
	var records []dnsRecord
	if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
		return nil, err
	}
	return records, nil
}

func (p *Provider) createDNSRecord(ctx context.Context, domainID int, record dnsRecord) error {
	url := fmt.Sprintf("%s/domains/%d/dns", apiEndpoint, domainID)
	body, err := json.Marshal(record)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.SetBasicAuth(p.APIToken, p.APISecret)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create DNS record: %s - %s", resp.Status, string(bodyBytes))
	}
	return nil
}

func (p *Provider) deleteDNSRecord(ctx context.Context, domainID, recordID int) error {
	url := fmt.Sprintf("%s/domains/%d/dns/%d", apiEndpoint, domainID, recordID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(p.APIToken, p.APISecret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete DNS record: %s", resp.Status)
	}
	return nil
}

var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
