variable "TAG" {
  default = "latest"
}

variable "REGISTRY" {
  default = "ghcr.io/nerzhul/home-bt-broker"
}

group "default" {
  targets = ["app"]
}

target "app" {
  context = "."
  dockerfile = "Dockerfile"
  platforms = ["linux/amd64", "linux/arm64"]
  tags = [
    "${REGISTRY}:${TAG}",
    "${REGISTRY}:latest"
  ]
  output = ["type=registry"]
}

# Local build target for testing
target "local" {
  inherits = ["app"]
  platforms = ["linux/amd64"]
  output = ["type=docker"]
  tags = ["home-bt-broker:local"]
}