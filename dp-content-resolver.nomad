job "dp-content-resolver" {
  datacenters = ["eu-west-1"]
  region      = "eu"
  type        = "service"

  update {
    stagger      = "20s"
    max_parallel = 1
  }

  group "web" {
    count = "{{WEB_TASK_COUNT}}"

    constraint {
      attribute = "${node.class}"
      value     = "web"
    }

    task "dp-content-resolver-web" {
      driver = "docker"

      artifact {
        source = "s3::https://s3-eu-west-1.amazonaws.com/{{DEPLOYMENT_BUCKET}}/dp-content-resolver/{{REVISION}}.tar.gz"
      }

      config {
        command = "${NOMAD_TASK_DIR}/start-task"

        args = ["./dp-content-resolver"]

        image = "{{ECR_URL}}:concourse-{{REVISION}}"

        port_map {
          http = 8080
        }
      }

      service {
        name = "dp-content-resolver"
        port = "http"
        tags = ["web"]
      }

      resources {
        cpu    = "{{WEB_RESOURCE_CPU}}"
        memory = "{{WEB_RESOURCE_MEM}}"

        network {
          port "http" {}
        }
      }

      template {
        source      = "${NOMAD_TASK_DIR}/vars-template"
        destination = "${NOMAD_TASK_DIR}/vars"
      }

      vault {
        policies = ["dp-content-resolver-web"]
      }
    }
  }

  group "publising" {
    count = "{{PUBLISHING_TASK_COUNT}}"

    constraint {
      attribute = "${node.class}"
      value     = "publishing"
    }

    task "dp-content-resolver-publishing" {
      driver = "docker"

      artifact {
        source = "s3::https://s3-eu-west-1.amazonaws.com/{{DEPLOYMENT_BUCKET}}/dp-content-resolver/{{REVISION}}.tar.gz"
      }

      config {
        command = "${NOMAD_TASK_DIR}/start-task"

        args = ["./dp-content-resolver"]

        image = "{{ECR_URL}}:concourse-{{REVISION}}"

        port_map {
          http = 8080
        }
      }

      service {
        name = "dp-content-resolver"
        port = "http"
        tags = ["publishing"]
      }

      resources {
        cpu    = "{{PUBLISHING_RESOURCE_CPU}}"
        memory = "{{PUBLISHING_RESOURCE_MEM}}"

        network {
          port "http" {}
        }
      }

      template {
        source      = "${NOMAD_TASK_DIR}/vars-template"
        destination = "${NOMAD_TASK_DIR}/vars"
      }

      vault {
        policies = ["dp-content-resolver-publishing"]
      }
    }
  }
}
