# Contratos de Eventos

Este documento define o schema canônico de cada evento que trafega pelo barramento (Redpanda/Kafka). Todos os produtores **devem** respeitar estes schemas. Todos os consumers **devem** ignorar campos desconhecidos (compatibilidade futura).

---

## Princípios

- **Entrega at-least-once:** o broker pode reentregr mensagens. Cada consumer deve ser idempotente — processar o mesmo evento duas vezes deve produzir o mesmo resultado.
- **Chave de partição = `device_id`:** garante ordenação de eventos por dispositivo dentro de uma partição.
- **Envelope comum:** todos os eventos compartilham o mesmo wrapper com campo `payload` específico por tipo.
- **Versionamento:** mudanças incompatíveis exigem novo valor de `version` e período de migração com dupla publicação.

---

## Envelope Comum

```json
{
  "event_id":   "<uuid-v4>",
  "event_type": "<string>",
  "version":    "1.0",
  "occurred_at": "<RFC3339 UTC>",
  "device_id":  "<string>",
  "payload":    {}
}
```

| Campo | Tipo | Descrição |
|---|---|---|
| `event_id` | UUID v4 | ID único desta instância de evento. Usado para verificação de idempotência. |
| `event_type` | string | Um dos tipos listados abaixo. |
| `version` | string | Versão do schema, ex: `"1.0"`. |
| `occurred_at` | RFC3339 | Timestamp UTC de quando o evento foi produzido. |
| `device_id` | string | ID do dispositivo que originou o evento. Usado como chave de partição. |
| `payload` | object | Dados específicos do domínio. Schema varia por `event_type`. |

---

## Eventos

### `telemetry.received`

**Tópico:** `telemetry.received`
**Produtor:** `ingestion-service`
**Consumers:** `device-service`, `alert-service`

Publicado quando uma leitura de telemetria é recebida e validada pelo serviço de ingestão.

```json
{
  "event_id":    "550e8400-e29b-41d4-a716-446655440000",
  "event_type":  "telemetry.received",
  "version":     "1.0",
  "occurred_at": "2026-06-02T14:30:00Z",
  "device_id":   "device-001",
  "payload": {
    "lat":           -15.62,
    "lon":           -47.66,
    "battery":        0.18,
    "temperature_c":  42.5,
    "speed_kmh":      95.0
  }
}
```

| Campo do Payload | Tipo | Descrição |
|---|---|---|
| `lat` | float | Latitude em graus decimais. |
| `lon` | float | Longitude em graus decimais. |
| `battery` | float | Nível de bateria entre 0.0 (vazio) e 1.0 (cheio). |
| `temperature_c` | float | Temperatura interna do dispositivo em °C. |
| `speed_kmh` | float | Velocidade em km/h (0 se estacionário). |

---

### `device.state.updated`

**Tópico:** `device.state.updated`
**Produtor:** `device-service`
**Consumers:** `dashboard-service`

Publicado após o `device-service` persistir o estado mais recente do dispositivo no PostgreSQL e atualizar o Redis.

```json
{
  "event_id":    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "event_type":  "device.state.updated",
  "version":     "1.0",
  "occurred_at": "2026-06-02T14:30:01Z",
  "device_id":   "device-001",
  "payload": {
    "last_lat":           -15.62,
    "last_lon":           -47.66,
    "last_battery":        0.18,
    "last_temperature_c":  42.5,
    "last_speed_kmh":      95.0,
    "updated_at":         "2026-06-02T14:30:01Z"
  }
}
```

---

### `alert.triggered`

**Tópico:** `alert.triggered`
**Produtor:** `alert-service`
**Consumers:** `dashboard-service`

Publicado quando uma regra de alerta é violada. Um único evento de telemetria pode gerar múltiplos alertas.

```json
{
  "event_id":    "6ba7b811-9dad-11d1-80b4-00c04fd430c8",
  "event_type":  "alert.triggered",
  "version":     "1.0",
  "occurred_at": "2026-06-02T14:30:01Z",
  "device_id":   "device-001",
  "payload": {
    "rule":     "battery_low",
    "severity": "critical",
    "value":    0.18,
    "limit":    0.20
  }
}
```

| Campo do Payload | Tipo | Valores possíveis |
|---|---|---|
| `rule` | string | `battery_low`, `high_temperature`, `speeding` |
| `severity` | string | `warning`, `critical` |
| `value` | float | Valor medido que violou o limite. |
| `limit` | float | Limite configurado para a regra. |

### Regras de alerta (Fase 1)

| Regra | Condição | Severidade |
|---|---|---|
| `battery_low` | `battery < 0.20` | `critical` |
| `high_temperature` | `temperature_c > 40.0` | `warning` |
| `speeding` | `speed_kmh > 80.0` | `warning` |

---

## Tópicos

| Tópico | Partições | Retenção | Produtor | Consumers |
|---|---|---|---|---|
| `telemetry.received` | 12 | 24h | ingestion-service | device-service, alert-service |
| `device.state.updated` | 6 | 7 dias | device-service | dashboard-service |
| `alert.triggered` | 6 | 7 dias | alert-service | dashboard-service |

---

## Consumer Groups

| Consumer Group | Serviço | Tópicos |
|---|---|---|
| `device-service-cg` | device-service | `telemetry.received` |
| `alert-service-cg` | alert-service | `telemetry.received` |
| `dashboard-service-cg` | dashboard-service | `device.state.updated`, `alert.triggered` |

---

## Padrão de Idempotência

Cada consumer armazena os `event_id` já processados em sua própria tabela:

```sql
CREATE TABLE processed_events (
    event_id     UUID PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Antes de processar qualquer evento, o consumer verifica se o `event_id` já existe. Se sim, confirma o offset e descarta. Isso garante semântica exactly-once na camada de aplicação sobre entrega at-least-once.

---

## Fluxo de Eventos

```
Dispositivo
    │
    ▼
ingestion-service ──► [telemetry.received] ──────────────────► device-service
                                │                                      │
                                │                                      ▼
                                │                             [device.state.updated]
                                │                                      │
                                ▼                                      ▼
                          alert-service                       dashboard-service ──► Tela Web
                                │                                      ▲
                                └──────► [alert.triggered] ───────────┘
```
