/**
 * Stress Test — pico extremo para acionar o HPA.
 * Sobe até 2000 dispositivos, observa auto-scaling e degradação graceful.
 * Objetivo: provar que o sistema não cai, apenas fica mais lento no pico.
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { BASE_URL, errorRate, telemetryDuration, telemetryPayload } from './helpers.js';

export const options = {
  stages: [
    { duration: '2m',  target: 100  },  // aquecimento
    { duration: '3m',  target: 500  },  // carga crescente
    { duration: '3m',  target: 2000 },  // pico — HPA deve escalar aqui
    { duration: '3m',  target: 500  },  // redução — HPA escala down
    { duration: '2m',  target: 0    },  // resfriamento
  ],
  thresholds: {
    // no pico aceita degradação, mas o sistema não pode cair (erro > 5%)
    http_req_duration: ['p(95)<2000'],
    errors:            ['rate<0.05'],
  },
};

export default function () {
  const res = http.post(`${BASE_URL}/telemetry`, telemetryPayload(__VU), {
    headers: { 'Content-Type': 'application/json' },
  });

  const ok = check(res, {
    'status 202 ou 503': (r) => r.status === 202 || r.status === 503,
  });

  errorRate.add(!ok);
  telemetryDuration.add(res.timings.duration);

  sleep(0.5);
}
