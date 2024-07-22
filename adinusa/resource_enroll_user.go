package adinusa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceEnrollUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEnrollUserCreate,
		ReadContext:   resourceEnrollUserRead,
		UpdateContext: resourceEnrollUserUpdate,
		DeleteContext: resourceEnrollUserDelete,

		Schema: map[string]*schema.Schema{
			"usernames": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
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
	_, err = getBatchIDByClass(client, courseID, className, courseName)
	if err != nil {
		return fmt.Errorf("batch '%s' not found for course '%s'", className, courseName)
	}

	return nil
}

func resourceEnrollUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*Client)
	usernames := getStringListFromSchema(d.Get("usernames").([]interface{}))
	courseName := d.Get("course_name").(string)
	className := d.Get("class_name").(string)

	// Get Course ID
	courseID, err := getCourseIDByName(client, courseName)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get Batch ID
	batchID, err := getBatchIDByClass(client, courseID, className, courseName)
	if err != nil {
		return diag.FromErr(err)
	}

	// Enroll Users
	enrollURL := client.APIURL + "/admin/enrollment/enroll_users/"
	payload := map[string]interface{}{
		"batch_id":  batchID,
		"usernames": usernames,
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
		return diag.Errorf("failed to enroll users, status: %s", resp.Status)
	}

	d.SetId(strings.Join(usernames, ","))

	return diags
}

func resourceEnrollUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*Client)
	usernames := getStringListFromSchema(d.Get("usernames").([]interface{}))
	courseName := d.Get("course_name").(string)
	className := d.Get("class_name").(string)

	// Get Course ID
	courseID, err := getCourseIDByName(client, courseName)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get Batch ID
	batchID, err := getBatchIDByClass(client, courseID, className, courseName)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check User Enrollment
	checkURL := client.APIURL + "/admin/enrollment/check_user/"
	payload := map[string]interface{}{
		"course_id": courseID,
		"batch_id":  batchID,
		"usernames": usernames,
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

	d.SetId(strings.Join(usernames, ","))

	return diags
}

func resourceEnrollUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("usernames") {
		old, new := d.GetChange("usernames")
		oldUsernames := getStringListFromSchema(old.([]interface{}))
		newUsernames := getStringListFromSchema(new.([]interface{}))

		toRevoke := difference(oldUsernames, newUsernames)
		if len(toRevoke) > 0 {
			diags := revokeUsers(ctx, d, m, toRevoke)
			if len(diags) > 0 {
				return diags
			}
		}

		toEnroll := difference(newUsernames, oldUsernames)
		if len(toEnroll) > 0 {
			diags := enrollUsers(ctx, d, m, toEnroll)
			if len(diags) > 0 {
				return diags
			}
		}
	}
	return resourceEnrollUserRead(ctx, d, m)
}

func resourceEnrollUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*Client)
	usernames := getStringListFromSchema(d.Get("usernames").([]interface{}))
	courseName := d.Get("course_name").(string)
	className := d.Get("class_name").(string)

	// Get Course ID
	courseID, err := getCourseIDByName(client, courseName)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get Batch ID
	batchID, err := getBatchIDByClass(client, courseID, className, courseName)
	if err != nil {
		return diag.FromErr(err)
	}

	// Revoke Users
	revokeURL := client.APIURL + "/admin/enrollment/revoke_users/"
	payload := map[string]interface{}{
		"course_id": courseID,
		"batch_id":  batchID,
		"usernames": usernames,
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
		return diag.Errorf("failed to revoke users, status: %s", resp.Status)
	}

	d.SetId("") // Clear the resource ID to signal deletion

	return diags
}

func getStringListFromSchema(input []interface{}) []string {
	var result []string
	for _, v := range input {
		result = append(result, v.(string))
	}
	return result
}

func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func enrollUsers(ctx context.Context, d *schema.ResourceData, m interface{}, usernames []string) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*Client)
	courseName := d.Get("course_name").(string)
	className := d.Get("class_name").(string)

	// Get Course ID
	courseID, err := getCourseIDByName(client, courseName)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get Batch ID
	batchID, err := getBatchIDByClass(client, courseID, className, courseName)
	if err != nil {
		return diag.FromErr(err)
	}

	// Enroll Users
	enrollURL := client.APIURL + "/admin/enrollment/enroll_users/"
	payload := map[string]interface{}{
		"batch_id":  batchID,
		"usernames": usernames,
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
		return diag.Errorf("failed to enroll users, status: %s", resp.Status)
	}

	return diags
}

func revokeUsers(ctx context.Context, d *schema.ResourceData, m interface{}, usernames []string) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*Client)
	courseName := d.Get("course_name").(string)
	className := d.Get("class_name").(string)

	// Get Course ID
	courseID, err := getCourseIDByName(client, courseName)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get Batch ID
	batchID, err := getBatchIDByClass(client, courseID, className, courseName)
	if err != nil {
		return diag.FromErr(err)
	}

	// Revoke Users
	revokeURL := client.APIURL + "/admin/enrollment/revoke_users/"
	payload := map[string]interface{}{
		"course_id": courseID,
		"batch_id":  batchID,
		"usernames": usernames,
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
		return diag.Errorf("failed to revoke users, status: %s", resp.Status)
	}

	return diags
}