terraform {
  required_providers {
    adinusa = {
      version = "~> 1.0.0"
      source  = "adinusa.test/adinusaprovider/adinusa"
    }
  }
}

provider "adinusa" {
  main_api_url = "https://dev.adinusa.id/api"
  api_url      = "https://dev.adinusa.id/api/pro-training"
  username     = "admin"
  password     = ""
}

resource "adinusa_class" "test_class" {
  course_name    = "Kubernetes Application Developer"
  class_name     = "TEST-API2"
  start_date     = "2024-07-17"
  end_date       = "2024-07-20"
  group_type     = "eksternal"
  is_last_batch  = false
  is_enroll_pass = false
  is_certificate = true
  is_schedule    = true
  is_active      = true
}

resource "adinusa_enroll_user" "example_enroll" {
  course_name = "Kubernetes Application Developer"
  class_name  = "TEST-API-oke"
  usernames    = ["atwatanmalikm2"]
}