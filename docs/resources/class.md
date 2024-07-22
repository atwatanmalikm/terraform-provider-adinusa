# `adinusa_class` Resource

The adinusa_class resource allows you to create and manage class in Adinusa.


## Example Usage

```hcl
resource "adinusa_class" "example_class" {
  course_name    = "Kubernetes Application Developer"
  class_name     = "K9DEV-CLASS-1"
  start_date     = "2024-07-17"
  end_date       = "2024-07-20"
  group_type     = "eksternal"
  is_last_batch  = false
  is_enroll_pass = false
  is_certificate = true
  is_schedule    = true
  is_active      = true
}
```

## Argument Reference

* `course_name` - (Required) The name of the course associated with the class. This should match an existing course in Adinusa.
* `class_name` - (Required) The name of the class. This name will be used to identify the class in Adinusa.
* `start_date` - (Required) The start date of the class in YYYY-MM-DD format.
* `end_date` - (Required) The end date of the class in YYYY-MM-DD format.
* `group_type` - (Required) The type of group. Valid values are: 
  - `internal`: Internal forum.
  - `eksternal`: External forum.
* `is_last_batch` - (Optional) A boolean indicating whether this is the last batch of the class. Defaults to false.
* `is_enroll_pass` - (Optional) A boolean indicating whether enrollment pass is required. Defaults to false.
* `is_certificate` - (Optional) A boolean indicating whether a certificate is issued upon completion of the class. Defaults to true.
* `is_schedule` - (Optional) A boolean indicating whether the class is scheduled. Defaults to true.
* `is_active` - (Optional) A boolean indicating whether the class is active. Defaults to true.