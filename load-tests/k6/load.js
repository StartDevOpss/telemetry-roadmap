/**
 * Load Test — carga normal de produção.
 * Rampa até 200 dispositivos simultâneos por 5 minutos.
 * Valida p95 < 300ms e taxa de erro < 1%.
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { BASE_URL, errorRate, telemetryDuration, telemetryPayload } from './helpers.js';

export const options = {
  stages: [
    { duration: '1m',  target: 50  },  // aquecimento
    { duration: '1m',  target: 200 },  // rampa até carga alvo
    { duration: '3m',  target: 200 },  // sustentação
    { duration: '1m',  target: 0   },  // rampa down
  ],
  thresholds: {
    http_req_duration: ['p(95)<300', 'p(99)<500'],
    errors:            ['rate<0.01'],
  },
};

export default function () {
  const res = http.post(`${BASE_URL}/telemetry`, telemetryPayload(__VU), {
    headers: { 'Content-Type': 'application/json' },
  });

  const ok = check(res, {
    'status 202':      (r) => r.status === 202,
    'latência < 300ms': (r) => r.timings.duration < 300,
  });

  errorRate.add(!ok);
  telemetryDuration.add(res.timings.duration);

  // dispositivos reais enviam a cada ~2 segundos
  sleep(2);
}
