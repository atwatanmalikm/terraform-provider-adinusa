package adinusa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"bytes"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceEnrollUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEnrollUserCreate,
		ReadContext:   resourceEnrollUserRead,
		UpdateContext: resourceEnrollUserCreate, // Use Create function for update
		DeleteContext: resourceEnrollUserDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"course_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"class_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: resourceEnrollUserCustomizeDiff,
	}
}

func resourceEnrollUserCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	client := m.(*Client)
	courseName := d.Get("course_name").(string)
	className := d.Get("class_name").(string)

	// Get Course ID
	courseID, err := getCourseIDByName(client, courseName)
	if err != nil {
		return fmt.Errorf("failed to get course ID: %v", err)
	}

	// Check if Batch ID exists
	_, err = getBatchIDByClass(client, courseID, className)
	if err != nil {
		return fmt.Errorf("batch '%s' not found for course '%s'", className, courseName)
	}

	return nil
}

func resourceEnrollUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*Client)
	username := d.Get("username").(string)
	courseName := d.Get("course_name").(string)
	className := d.Get("class_name").(string)

	// Get Course ID
	courseID, err := getCourseIDByName(client, courseName)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get Batch ID
	batchID, err := getBatchIDByClass(client, courseID, className)
	if err != nil {
		return diag.FromErr(err)
	}

	// Enroll User
	enrollURL := client.APIURL + "/admin/enrollment/enroll_users/"
	payload := map[string]interface{}{
		"batch_id":  batchID,
		"usernames": []string{username},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	req, err := http.NewRequest("POST", enrollURL, bytes.NewBuffer(body))
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.AuthToken)

	resp, err := client.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("failed to enroll user, status: %s", resp.Status)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return diag.FromErr(err)
	}

	if len(result) < 1 {
		return diag.Errorf("no response from enrollment endpoint")
	}

	d.SetId(username)

	return diags
}

func resourceEnrollUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*Client)
	username := d.Get("username").(string)
	courseName := d.Get("course_name").(string)
	className := d.Get("class_name").(string)

	// Get Course ID
	courseID, err := getCourseIDByName(client, courseName)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get Batch ID
	batchID, err := getBatchIDByClass(client, courseID, className)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check User Enrollment
	checkURL := client.APIURL + "/admin/enrollment/check_user/"
	payload := map[string]interface{}{
		"course_id": courseID,
		"batch_id":  batchID,
		"usernames": []string{username},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	req, err := http.NewRequest("POST", checkURL, bytes.NewBuffer(body))
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.AuthToken)

	resp, err := client.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("failed to check user enrollment, status: %s", resp.Status)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return diag.FromErr(err)
	}

	if len(result) < 1 {
		return diag.Errorf("no response from check enrollment endpoint")
	}

	d.SetId(username)

	return diags
}

func resourceEnrollUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*Client)
	username := d.Get("username").(string)
	courseName := d.Get("course_name").(string)
	className := d.Get("class_name").(string)

	// Get Course ID
	courseID, err := getCourseIDByName(client, courseName)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get Batch ID
	batchID, err := getBatchIDByClass(client, courseID, className)
	if err != nil {
		return diag.FromErr(err)
	}

	// Revoke User
	revokeURL := client.APIURL + "/admin/enrollment/revoke_users/"
	payload := map[string]interface{}{
		"course_id": courseID,
		"batch_id":  batchID,
		"usernames": []string{username},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	req, err := http.NewRequest("POST", revokeURL, bytes.NewBuffer(body))
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.AuthToken)

	resp, err := client.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("failed to revoke user, status: %s", resp.Status)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return diag.FromErr(err)
	}

	// Check for expected response message
	message, ok := result["message"].(string)
	if !ok || message != "Berhasil mencabut akses user dari course" {
		return diag.Errorf("unexpected response from revoke endpoint: %v", result)
	}

	d.SetId("") // Clear the resource ID to signal deletion

	return diags
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

func getBatchIDByClass(client *Client, courseID int, className string) (int, error) {
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

	return 0, fmt.Errorf("batch '%s' not found for course ID '%d'", className, courseID)
}