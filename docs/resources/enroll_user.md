# `adinusa_enroll_user` Resource

The adinusa_enroll_user resource allows you to manage user enrollment in Adinusa.


## Example Usage

```hcl
resource "adinusa_enroll_user" "example_enroll" {
  course_name = "Kubernetes Application Developer"
  class_name  = "K9DEV-CLASS-1"
  usernames = [
    "user1",
    "user2",
    "user3"
  ]
}
```

## Argument Reference

* `course_name` - (Required) The name of the course associated with the class. This should match an existing course in Adinusa.
* `class_name` - (Required) The name of the class. This name will be used to identify the class in Adinusa.
* `usernames` - (Required) A list of usernames to be enrolled in the specified class. 