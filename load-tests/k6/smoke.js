/**
 * Smoke Test — sanidade mínima do sistema.
 * 5 dispositivos, 30 segundos. Falha rápido se algo estiver quebrado.
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { BASE_URL, errorRate, telemetryDuration, telemetryPayload } from './helpers.js';

export const options = {
  vus:      5,
  duration: '30s',
  thresholds: {
    http_req_duration: ['p(95)<500'],
    errors:            ['rate<0.01'],
  },
};

export default function () {
  const res = http.post(`${BASE_URL}/telemetry`, telemetryPayload(__VU), {
    headers: { 'Content-Type': 'application/json' },
  });

  const ok = check(res, {
    'status 202': (r) => r.status === 202,
    'latência < 500ms': (r) => r.timings.duration < 500,
  });

  errorRate.add(!ok);
  telemetryDuration.add(res.timings.duration);

  sleep(1);
}
