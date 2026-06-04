resource "kind_cluster" "telemetry" {
  name = var.cluster_name

  kind_config {
    kind        = "Cluster"
    api_version = "kind.x-k8s.io/v1alpha4"

    node {
      role = "control-plane"

      kubeadm_config_patches = [
        <<-EOT
          kind: InitConfiguration
          nodeRegistration:
            kubeletExtraArgs:
              node-labels: "ingress-ready=true"
        EOT
      ]

      extra_port_mappings {
        container_port = var.ingestion_node_port
        host_port      = var.ingestion_node_port
        protocol       = "TCP"
      }

      extra_port_mappings {
        container_port = var.redpanda_console_node_port
        host_port      = var.redpanda_console_node_port
        protocol       = "TCP"
      }
    }
  }
}
