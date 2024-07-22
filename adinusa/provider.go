package adinusa

import (
	"context"
	"encoding/json"
	"net/http"
	"bytes"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Client struct {
	APIURL    string
	AuthToken string
	*http.Client
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"main_api_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("MAIN_API_URL", nil),
				Description: "Main API URL for Adinusa",
			},
			"api_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("API_URL", nil),
				Description: "API URL for Adinusa",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ADINUSA_USERNAME", nil),
				Description: "Username for Adinusa API",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ADINUSA_PASSWORD", nil),
				Description: "Password for Adinusa API",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"adinusa_enroll_user": resourceEnrollUser(),
			"adinusa_class": resourceClass(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	mainAPIURL := d.Get("main_api_url").(string)
	apiURL := d.Get("api_url").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	loginURL := mainAPIURL + "/auth/login"
	payload := map[string]string{
		"username": username,
		"password": password,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, diag.FromErr(err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, diag.Errorf("failed to authenticate, status: %s", resp.Status)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, diag.FromErr(err)
	}

	token, ok := result["access"].(string)
	if !ok {
		return nil, diag.Errorf("failed to get token")
	}

	return &Client{
		APIURL:    apiURL,
		AuthToken: token,
		Client:    client,
	}, diags
}

func getCourseIDByName(client *Client, courseName string) (int, error) {
	coursesURL := client.APIURL + "/courses/"
	req, err := http.NewRequest("GET", coursesURL, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+client.AuthToken)
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get courses, status: %s", resp.Status)
	}

	var courses []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&courses); err != nil {
		return 0, err
	}

	for _, course := range courses {
		if course["title"].(string) == courseName {
			return int(course["id"].(float64)), nil
		}
	}

	return 0, fmt.Errorf("course '%s' not found", courseName)
}

func getBatchIDByClass(client *Client, courseID int, className string, courseName string) (int, error) {
	batchesURL := fmt.Sprintf("%s/admin/batchs/?course_id=%d", client.APIURL, courseID)
	req, err := http.NewRequest("GET", batchesURL, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+client.AuthToken)
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get batches, status: %s", resp.Status)
	}

	var batches []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&batches); err != nil {
		return 0, err
	}

	for _, batch := range batches {
		if batch["batch"].(string) == className {
			return int(batch["id"].(float64)), nil
		}
	}

	return 0, fmt.Errorf("batch '%s' not found for course '%s'", className, courseName)
}