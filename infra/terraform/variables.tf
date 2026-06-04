variable "cluster_name" {
  description = "Nome do cluster kind"
  type        = string
  default     = "telemetry-platform"
}

variable "ingestion_node_port" {
  description = "NodePort exposto pelo ingestion-service"
  type        = number
  default     = 30081
}

variable "redpanda_console_node_port" {
  description = "NodePort exposto pelo Redpanda Console"
  type        = number
  default     = 30090
}
