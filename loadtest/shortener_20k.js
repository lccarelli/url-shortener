import http from 'k6/http';
import { check } from 'k6';

export const options = {
    scenarios: {
        create: {
            executor: 'constant-arrival-rate',
            rate: 10000,
            timeUnit: '1s',
            duration: '20s',
            preAllocatedVUs: 2000,
            maxVUs: 4000,
            gracefulStop: '0s',
            exec: 'createFn',
        },
        resolve: {
            executor: 'constant-arrival-rate',
            rate: 10000,
            timeUnit: '1s',
            duration: '20s',
            startTime: '0s',
            preAllocatedVUs: 2000,
            maxVUs: 4000,
            gracefulStop: '0s',
            exec: 'resolveFn',
        },
    },
    thresholds: {
        http_req_failed: ['rate<0.01'],
        'http_req_duration{expected_response:true}': ['p(90)<100', 'p(95)<200'],
    },
};

const params = {
    headers: { 'Content-Type': 'application/json' },
    timeout: '5s',
};

const hosts = [
    'http://shortener1:8080',
    'http://shortener2:8080',
    'http://shortener3:8080',
];

export function createFn () {
    const h = hosts[__VU % hosts.length];
    const res = http.post(`${h}/shorten`, JSON.stringify({ url: 'https://meli.com' }), params);
    check(res, { 'status is < 300': r => r.status < 300 });
}

export function resolveFn () {
    const h = hosts[(__VU + 1) % hosts.length];
    const key = Math.random().toString(36).slice(2, 8);
    const res = http.get(`${h}/${key}`, { redirects: 0, timeout: '5s' });
    check(res, { 'status is < 400': r => r.status < 400 });
}
