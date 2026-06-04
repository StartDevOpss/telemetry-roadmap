import { Rate, Trend } from 'k6/metrics';

export const errorRate   = new Rate('errors');
export const telemetryDuration = new Trend('telemetry_duration', true);

export const BASE_URL = __ENV.BASE_URL || 'http://localhost:30081';

const BRAZIL_BOUNDS = {
  latMin: -33.0, latMax: 5.0,
  lonMin: -73.0, lonMax: -34.0,
};

function rand(min, max) {
  return Math.random() * (max - min) + min;
}

export function deviceId(vu) {
  return `device-${String(vu).padStart(5, '0')}`;
}

export function telemetryPayload(vu) {
  return JSON.stringify({
    device_id: deviceId(vu),
    payload: {
      lat:           rand(BRAZIL_BOUNDS.latMin, BRAZIL_BOUNDS.latMax),
      lon:           rand(BRAZIL_BOUNDS.lonMin, BRAZIL_BOUNDS.lonMax),
      battery:       parseFloat(rand(0.05, 1.0).toFixed(2)),
      temperature_c: parseFloat(rand(15.0, 55.0).toFixed(1)),
      speed_kmh:     parseFloat(rand(0.0, 120.0).toFixed(1)),
    },
  });
}
