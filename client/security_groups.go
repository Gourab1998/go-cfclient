package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/cloudfoundry-community/go-cfclient/resource"
)

type SecurityGroupClient commonClient

// ListByQuery retrieves security groups based on query
func (c *SecurityGroupClient) ListByQuery(query url.Values) ([]resource.SecurityGroup, error) {
	var securityGroups []resource.SecurityGroup
	requestURL, err := url.Parse("/v3/security_groups")
	if err != nil {
		return nil, err
	}
	requestURL.RawQuery = query.Encode()

	for {
		r := c.client.NewRequest("GET", fmt.Sprintf("%s?%s", requestURL.Path, requestURL.RawQuery))
		resp, err := c.client.DoRequest(r)
		if err != nil {
			return nil, fmt.Errorf("error requesting security groups: %w", err)
		}
		defer func(b io.ReadCloser) {
			_ = b.Close()
		}(resp.Body)

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error listing security groups, response code: %d", resp.StatusCode)
		}

		var data resource.ListSecurityGroupResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, fmt.Errorf("error parsing JSON from list security groups: %w", err)
		}

		securityGroups = append(securityGroups, data.Resources...)

		requestURL, err = url.Parse(data.Pagination.Next.Href)
		if err != nil {
			return nil, fmt.Errorf("error parsing next page URL: %w", err)
		}
		if requestURL.String() == "" {
			break
		}
	}

	return securityGroups, nil
}

// Create creates security group from CreateSecurityGroupRequest
func (c *SecurityGroupClient) Create(r resource.CreateSecurityGroupRequest) (*resource.SecurityGroup, error) {
	req := c.client.NewRequest("POST", "/v3/security_groups")

	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(r); err != nil {
		return nil, err
	}
	req.body = buf

	resp, err := c.client.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error while creating security group: %w", err)
	}
	defer func(b io.ReadCloser) {
		_ = b.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("error creating security group %s, response code: %d", r.Name, resp.StatusCode)
	}

	var securitygroup resource.SecurityGroup
	if err := json.NewDecoder(resp.Body).Decode(&securitygroup); err != nil {
		return nil, fmt.Errorf("error reading security group JSON: %w", err)
	}

	return &securitygroup, nil
}

// Delete deletes security group by GUID
func (c *SecurityGroupClient) Delete(GUID string) error {
	req := c.client.NewRequest("DELETE", "/v3/security_groups/"+GUID)

	resp, err := c.client.DoRequest(req)
	if err != nil {
		return fmt.Errorf("error while deleting security group: %w", err)
	}
	defer func(b io.ReadCloser) {
		_ = b.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("error deleting security group with GUID [%s], response code: %d", GUID, resp.StatusCode)
	}
	return nil
}

// Update updates security group by GUID and from UpdateSecurityGroupRequest
func (c *SecurityGroupClient) Update(GUID string, r resource.UpdateSecurityGroupRequest) (*resource.SecurityGroup, error) {
	req := c.client.NewRequest("PATCH", "/v3/security_groups/"+GUID)
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(r); err != nil {
		return nil, err
	}
	req.body = buf

	resp, err := c.client.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error while updating security group: %w", err)
	}
	defer func(b io.ReadCloser) {
		_ = b.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error updating security group %s, response code: %d", GUID, resp.StatusCode)
	}

	var securityGroup resource.SecurityGroup
	if err := json.NewDecoder(resp.Body).Decode(&securityGroup); err != nil {
		return nil, fmt.Errorf("error reading security group JSON: %w", err)
	}

	return &securityGroup, nil
}

// Get retrieves security group base on provided GUID
func (c *SecurityGroupClient) Get(GUID string) (*resource.SecurityGroup, error) {
	req := c.client.NewRequest("GET", "/v3/security_groups/"+GUID)

	resp, err := c.client.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error while getting security group: %w", err)
	}
	defer func(b io.ReadCloser) {
		_ = b.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error getting security group with GUID [%s], response code: %d", GUID, resp.StatusCode)
	}

	var securityGroup resource.SecurityGroup
	if err := json.NewDecoder(resp.Body).Decode(&securityGroup); err != nil {
		return nil, fmt.Errorf("error reading security group JSON: %w", err)
	}

	return &securityGroup, nil
}
