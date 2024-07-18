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

resource "adinusa_enroll_user" "example_enroll" {
  course_name = "Kubernetes Application Developer"
  class_name  = "TEST-API2"
  username    = "atwatanmalikm2"
}