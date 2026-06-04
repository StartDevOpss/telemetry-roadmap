output "cluster_name" {
  description = "Nome do cluster kind criado"
  value       = kind_cluster.telemetry.name
}

output "kubeconfig" {
  description = "Caminho do kubeconfig gerado pelo kind"
  value       = kind_cluster.telemetry.kubeconfig_path
  sensitive   = true
}

output "endpoint" {
  description = "Endpoint do API server do cluster"
  value       = kind_cluster.telemetry.endpoint
}
