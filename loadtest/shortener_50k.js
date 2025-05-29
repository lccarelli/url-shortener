import http from 'k6/http';
import { check } from 'k6';

export const options = {
    scenarios: {
        create: {
            executor: 'constant-arrival-rate',
            rate: 10000,
            timeUnit: '1s',
            duration: '30s',
            preAllocatedVUs: 8000,
            maxVUs: 12000,
            gracefulStop: '0s',
            exec: 'createFn',
        }
    },
    thresholds: {
        http_req_failed: ['rate<0.05'],
        'http_req_duration{expected_response:true}': ['p(95)<300'],
    },
};

// Optimización de headers y timeout
const params = {
    headers: { 'Content-Type': 'application/json' },
    timeout: '5s',
};

// Round robin directo a servicios
const hosts = [
    'http://shortener-no-redis:8080'
];


// Evita alta cardinalidad en métricas
const keys = ['a1b2c3', 'd4e5f6', 'g7h8i9', 'x1y2z3', 'p0q1r2'];

export function createFn () {
    const h = hosts[__VU % hosts.length];
    const res = http.post(`${h}/shorten`, JSON.stringify({ url: 'https://meli.com' }), params);
    check(res, { 'status is >= 200 && < 300': r => r.status >= 200 && r.status < 300 });
}
