package adinusa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceClass() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClassCreate,
		ReadContext:   resourceClassRead,
		UpdateContext: resourceClassUpdate,
		DeleteContext: resourceClassDelete,

		Schema: map[string]*schema.Schema{
			"class_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"course_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"start_date": {
				Type:     schema.TypeString,
				Required: true,
			},
			"end_date": {
				Type:     schema.TypeString,
				Required: true,
			},
			"group_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{"internal", "eksternal"}, false),
			},
			"is_last_batch": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"is_enroll_pass": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"is_certificate": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"is_schedule": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"is_active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceClassCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*Client)
	courseName := d.Get("course_name").(string)
	className := d.Get("class_name").(string)
	startDate := d.Get("start_date").(string)
	endDate := d.Get("end_date").(string)
	groupTypeStr := d.Get("group_type").(string)
	isLastBatch := d.Get("is_last_batch").(bool)
	isEnrollPass := d.Get("is_enroll_pass").(bool)
	isCertificate := d.Get("is_certificate").(bool)
	isSchedule := d.Get("is_schedule").(bool)

	groupType, err := convertGroupTypeToNumber(groupTypeStr)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get Course ID
	courseID, err := getCourseIDByName(client, courseName)
	if err != nil {
		return diag.FromErr(err)
	}

	// Create Class
	createURL := client.APIURL + "/admin/batchs/"
	payload := map[string]interface{}{
		"batch":          className,
		"start_date":     startDate,
		"end_date":       endDate,
		"group_type":     groupType,
		"is_last_batch":  isLastBatch,
		"is_enroll_pass": isEnrollPass,
		"is_certificate": isCertificate,
		"is_schedule":    isSchedule,
		"course":         courseID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	req, err := http.NewRequest("POST", createURL, bytes.NewBuffer(body))
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

	if resp.StatusCode != http.StatusCreated {
		return diag.Errorf("failed to create class, status: %s", resp.Status)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return diag.FromErr(err)
	}

	// Set ID
	classID := int(result["id"].(float64))
	d.SetId(fmt.Sprintf("%d", classID))

	// Activate Class if needed
	if d.Get("is_active").(bool) {
		err := changeClassStatus(client, classID, true)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceClassRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	classID := d.Id()

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/admin/batchs/%s/", client.APIURL, classID), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Authorization", "Bearer "+client.AuthToken)

	resp, err := client.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("failed to read class, status: %s", resp.Status)
	}

	var classResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&classResponse); err != nil {
		return diag.FromErr(err)
	}

	groupTypeStr, err := convertGroupTypeToString(int(classResponse["group_type"].(float64)))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("class_name", classResponse["batch"].(string))
	d.Set("course_name", classResponse["course_data"].(map[string]interface{})["title"].(string))
	d.Set("start_date", classResponse["start_date"].(string))
	d.Set("end_date", classResponse["end_date"].(string))
	d.Set("group_type", groupTypeStr)
	d.Set("is_last_batch", classResponse["is_last_batch"].(bool))
	d.Set("is_enroll_pass", classResponse["is_enroll_pass"].(bool))
	d.Set("is_certificate", classResponse["is_certificate"].(bool))
	d.Set("is_schedule", classResponse["is_schedule"].(bool))
	d.Set("is_active", classResponse["is_active"].(bool))

	return nil
}

func resourceClassUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	classID := d.Id()

	if d.HasChanges("class_name", "start_date", "end_date", "group_type", "is_last_batch", "is_enroll_pass", "is_certificate", "is_schedule", "course_name") {
		className := d.Get("class_name").(string)
		startDate := d.Get("start_date").(string)
		endDate := d.Get("end_date").(string)
		groupTypeStr := d.Get("group_type").(string)
		isLastBatch := d.Get("is_last_batch").(bool)
		isEnrollPass := d.Get("is_enroll_pass").(bool)
		isCertificate := d.Get("is_certificate").(bool)
		isSchedule := d.Get("is_schedule").(bool)
		courseName := d.Get("course_name").(string)

		groupType, err := convertGroupTypeToNumber(groupTypeStr)
		if err != nil {
			return diag.FromErr(err)
		}

		courseID, err := getCourseIDByName(client, courseName)
		if err != nil {
			return diag.FromErr(err)
		}

		classPayload := map[string]interface{}{
			"batch":          className,
			"start_date":     startDate,
			"end_date":       endDate,
			"group_type":     groupType,
			"is_last_batch":  isLastBatch,
			"is_enroll_pass": isEnrollPass,
			"is_certificate": isCertificate,
			"is_schedule":    isSchedule,
			"course":         courseID,
		}

		body, err := json.Marshal(classPayload)
		if err != nil {
			return diag.FromErr(err)
		}

		req, err := http.NewRequest("PUT", fmt.Sprintf("%s/admin/batchs/%s/", client.APIURL, classID), bytes.NewBuffer(body))
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
			return diag.Errorf("failed to update class, status: %s", resp.Status)
		}
	}

	if d.HasChange("is_active") {
		isActive := d.Get("is_active").(bool)
		classIDInt, err := strconv.Atoi(classID)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := changeClassStatus(client, classIDInt, isActive); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceClassRead(ctx, d, m)
}

func resourceClassDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*Client)
	classID := d.Id()

	url := fmt.Sprintf("%s/admin/batchs/%s/", client.APIURL, classID)
	req, err := http.NewRequest("DELETE", url, nil)
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

	// Handle 204 No Content status as a successful deletion
	if resp.StatusCode != http.StatusNoContent {
		return diag.Errorf("failed to delete class, status: %s", resp.Status)
	}

	d.SetId("")
	return diags
}

func changeClassStatus(client *Client, classID int, isActive bool) error {
	url := fmt.Sprintf("%s/admin/batchs/%d/change_status/", client.APIURL, classID)
	payload := map[string]interface{}{
		"is_broadcast": false,
		"is_active":    isActive,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.AuthToken)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to change class status, status: %s", resp.Status)
	}

	return nil
}

func convertGroupTypeToNumber(groupTypeStr string) (int, error) {
	switch groupTypeStr {
	case "internal":
		return 1, nil
	case "eksternal":
		return 2, nil
	default:
		return 0, fmt.Errorf("invalid group_type: %s", groupTypeStr)
	}
}

func convertGroupTypeToString(groupType int) (string, error) {
	switch groupType {
	case 1:
		return "internal", nil
	case 2:
		return "eksternal", nil
	default:
		return "", fmt.Errorf("invalid group_type: %d", groupType)
	}
}