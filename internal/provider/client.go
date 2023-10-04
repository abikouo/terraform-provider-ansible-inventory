package provider

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Client -
type Client struct {
	HostURL    				string
	Username   				*string
	Password   				*string
	InsecureSkipVerify		bool
}

// ansible host
type AnsibleHost struct {
	Name        string             			`json:"name"`
	Groups      []string           			`json:"groups"`
	Variables   map[string]string           `json:"variables"`
}

// NewClient -
func NewClient(host string, username *string, password *string, insecure_skip_verify bool) (*Client, error) {
	client := Client{
		HostURL: HostURL,
		Username: username,
		Password: password,
		InsecureSkipVerify: insecure_skip_verify, 
	}

	return &client, nil
}


func (c *Client) GetHosts(stateEndPoint *string) (*AnsibleHost, error) {
	
	if stateEndPoint == nil {
		return nil, fmt.Errorf("missing mandatory state endpoint")
	}

	req, err := http.NewRequest("GET", c.HostURL + *stateEndPoint, nil)
	if c.Username != nil && c.Password != nil {
    	req.SetBasicAuth(*c.Username, *c.Password)
	}

    req.Header.Set("Accept", "application/json")
    req.Header.Set("Content-Type", "application/json")

    tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: c.InsecureSkipVerify},
    }
    client := &http.Client{Transport: tr}
    resp, err := client.Do(req)

    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

    if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

    return GetAnsibleHost(body)
}


func GetAnsibleHost(body []byte)(*AnsibleHost, error){

	var result map[string]interface{}
    err = json.Unmarshal(body, &result)
    if err != nil {
		return nil, err
	}

    var hosts []AnsibleHost
    resources, ok := result["resources"].([]interface{})
    if ok {
        for _, resource := range resources{
            resource_obj := resource.(map[string]interface{})
            resource_type, ok := resource_obj["type"]
            if ok && resource_type == "ansible_host" {
                instances, ok := resource_obj["instances"].([]interface{})
                if ok{
                    for _, instance := range instances{
                        attributes, ok := instance.(map[string]interface{})["attributes"].(map[string]interface{})
                        if ok{
                            name := attributes["name"].(string)
                            var groups []string
                            for _, group := range attributes["groups"].([]interface{}){
                                groups = append(groups, group.(string))
                            }
                            variables := make(map[string]string)
                            for key, value := range attributes["variables"].(map[string]interface{}){
                                variables[key] = value.(string)
                            }
                            hosts = append(hosts, AnsibleHost{
                                Name: name,
                                Groups: groups,
                                Variables: variables,
                            })
                        }
                    }
                }
            }
        }
    }
    return &hosts, nil
}