# Expose a TCP port
resource "dokploy_port" "http" {
  published_port = 8080
  target_port    = 80
  protocol       = "tcp"
  application_id = dokploy_application.web.id
}

# Expose a UDP port
resource "dokploy_port" "dns" {
  published_port = 5353
  target_port    = 53
  protocol       = "udp"
  application_id = dokploy_application.dns.id
}

# Host mode port (bypasses ingress)
resource "dokploy_port" "metrics" {
  published_port = 9090
  target_port    = 9090
  protocol       = "tcp"
  publish_mode   = "host"
  application_id = dokploy_application.metrics.id
}
