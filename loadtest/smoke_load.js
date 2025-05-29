import http from 'k6/http';
import { check } from 'k6';

export const options = {
    vus: 50,
    duration: '10s',
    rps: 100,
};

const BASE = __ENV.BASE_URL || 'http://localhost';

export default function () {
    const res = http.post(`${BASE}/shorten`, '{"url":"https://meli.com"}', {
        headers: { 'Content-Type': 'application/json' },
        tags: { name: 'create' },
    });
    check(res, { 'status 2xx': r => r.status < 300 });
}
