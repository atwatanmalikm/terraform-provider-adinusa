package adinusa

import (
	"context"
	"encoding/json"
	"net/http"
	"bytes"

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