package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type ProxmoxApi struct {
	url    string
	ticket string
}

type key int

var proxmoxApiKey key

func NewContext(ctx context.Context, api *ProxmoxApi) context.Context {
	return context.WithValue(ctx, proxmoxApiKey, api)
}

func FromContext(ctx context.Context) (*ProxmoxApi, bool) {
	u, ok := ctx.Value(proxmoxApiKey).(*ProxmoxApi)
	return u, ok
}

type GetTicketPostResponse struct {
	Data GetTicketPostResponseData `json:"data"`
}

type GetTicketPostResponseData struct {
	Ticket              string `json:"ticket"`
	CSRFPreventionToken string `json:"CSRFPreventionToken"`
	ClusterName         string `json:"clustername"`
}

type ListNodesGetResponse struct {
	Data []Node `json:"data"`
}

type Node struct {
	Type           string  `json:"type"`
	MaxCPU         int     `json:"maxcpu"`
	MaxDisk        int64   `json:"maxdisk"`
	Node           string  `json:"node"`
	Level          string  `json:"level"`
	Uptime         int64   `json:"uptime"`
	Disk           int64   `json:"disk"`
	Status         string  `json:"status"`
	MaxMem         int64   `json:"maxmem"`
	Id             string  `json:"id"`
	CPU            float64 `json:"cpu"`
	Mem            int64   `json:"mem"`
	SSLFingerprint string  `json:"ssl_fingerprint"`
}

type ListQemuVmGetResponse struct {
	Data []QemuVm `json:"data"`
}

type QemuVm struct {
	Mem       int64   `json:"mem"`
	MaxMem    int64   `json:"maxmem"`
	Disk      int64   `json:"disk"`
	MaxDisk   int64   `json:"maxdisk"`
	DiskWrite int64   `json:"diskwrite"`
	DiskRead  int64   `json:"diskread"`
	Uptime    int64   `json:"uptime"`
	Status    string  `json:"status"`
	Name      string  `json:"name"`
	PID       int     `json:"pid"`
	VMID      int     `json:"vmid"`
	CPU       float64 `json:"cpu"`
	CPUs      int     `json:"cpus"`
	NetOut    int64   `json:"netout"`
	NetIn     int64   `json:"netin"`
}

type IpAddress struct {
	Prefix  int    `json:"prefix"`
	Address string `json:"ip-address"`
	Type    string `json:"ip-address-type"`
}

type ListQemuNetworkInterfacesGetResponse struct {
	Data ListQemuNetworkInterfacesResultsGetResponse `json:"data"`
}

type ListQemuNetworkInterfacesResultsGetResponse struct {
	Results []QemuNetworkInterface `json:"result"`
}

type QemuNetworkInterface struct {
	Name            string           `json:"name"`
	IpAddresses     []IpAddress      `json:"ip-addresses"`
	HardwareAddress string           `json:"hardware-address"`
	Statistics      map[string]int64 `json:"statistics"`
}

func CreateProxmoxApi(host string, port int, username string, password string) (*ProxmoxApi, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	url := fmt.Sprintf("https://%s:%d/api2/json", host, port)

	reqJson, err := json.Marshal(&map[string]string{
		"username": username,
		"password": password,
	})
	body := bytes.NewBuffer(reqJson)
	if err != nil {
		return nil, err
	}

	ticketUrl := fmt.Sprintf("%s/access/ticket", url)
	res, err := http.Post(ticketUrl, "application/json", body)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	data := GetTicketPostResponse{}
	json.Unmarshal(rawBody, &data)

	api := &ProxmoxApi{
		url:    url,
		ticket: data.Data.Ticket,
	}
	return api, nil
}

func (p *ProxmoxApi) rawRequest(method string, path string) (*http.Request, error) {
	url := fmt.Sprintf("%s/%s", p.url, path)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(&http.Cookie{
		Name:  "PVEAuthCookie",
		Value: p.ticket,
	})

	return req, nil
}

func (p *ProxmoxApi) rawGet(path string) (*http.Response, error) {
	req, err := p.rawRequest("GET", path)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	return client.Do(req)
}

func (p *ProxmoxApi) rawPost(path string) (*http.Response, error) {
	req, err := p.rawRequest("POST", path)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	return client.Do(req)
}

func (p *ProxmoxApi) ListNodes() ([]Node, error) {
	res, err := p.rawGet("nodes")
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	data := ListNodesGetResponse{}
	json.Unmarshal(rawBody, &data)

	return data.Data, nil
}

func (p *ProxmoxApi) ListQemuVMs(node Node) ([]QemuVm, error) {
	path := fmt.Sprintf("nodes/%s/qemu", node.Node)
	res, err := p.rawGet(path)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	data := ListQemuVmGetResponse{}
	json.Unmarshal(rawBody, &data)

	return data.Data, nil
}

func (p *ProxmoxApi) ListNetworkInterfaces(node Node, vm QemuVm) ([]QemuNetworkInterface, error) {
	path := fmt.Sprintf("nodes/%s/qemu/%d/agent/network-get-interfaces", node.Node, vm.VMID)
	res, err := p.rawGet(path)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	data := ListQemuNetworkInterfacesGetResponse{}
	json.Unmarshal(rawBody, &data)

	results := data.Data.Results
	return results, nil
}

func (p *ProxmoxApi) FindIpAddress(name string, iptype string) (string, error) {
	nodes, err := p.ListNodes()
	if err != nil {
		return "", err
	}

	matches := make([]QemuNetworkInterface, 0)
	for _, node := range nodes {
		vms, err := p.ListQemuVMs(node)
		if err != nil {
			return "", err
		}

		for _, vm := range vms {
			if vm.Name == name {
				interfaces, err := p.ListNetworkInterfaces(node, vm)

				if err != nil {
					return "", err
				}
				matches = append(matches, interfaces...)
			}
		}
	}

	for _, ni := range matches {
		// Ignore loopback MAC address
		if ni.HardwareAddress != "00:00:00:00:00:00" {
			for _, ip := range ni.IpAddresses {
				if ip.Type == iptype {
					return ip.Address, nil
				}
			}
		}
	}

	return "", errors.New(fmt.Sprintf("No VM or network interfaces found for name \"%s\"", name))
}
