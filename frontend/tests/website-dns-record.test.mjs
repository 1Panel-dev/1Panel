import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import test from 'node:test';

const read = (path) => readFileSync(new URL(path, import.meta.url), 'utf8');

const DNS_KEYS = [
    'dnsRecordManage',
    'dnsRecordAccount',
    'dnsRecordZone',
    'dnsRecordHost',
    'dnsRecordValue',
    'dnsRecordTTL',
    'dnsRecordPriority',
    'dnsRecordProxied',
    'dnsRecordRemark',
    'dnsRecordCreate',
    'dnsRecordNoAccount',
    'dnsRecordHostHelper',
    'dnsRecordUseIP',
    'dnsRecordUseIPMismatch',
    'dnsRecordUseIPEmpty',
    'dnsRecordValueHelper',
    'dnsRecordCfTTLHelper',
];

test('Manage DNS is a visible sidebar route placed right after Certificates', () => {
    const src = read('../src/routers/modules/website.ts');
    const ssl = src.indexOf("path: '/websites/ssl',");
    const dns = src.indexOf("path: '/websites/dns-record',");
    const next = src.indexOf("path: '/websites/ssl/ca',");
    assert.ok(ssl >= 0 && dns > ssl && next > dns, 'dns-record route must follow the SSL route');
    const block = src.slice(dns, next);
    assert.ok(!block.includes('hidden: true'), 'dns-record route must not be hidden');
    assert.match(block, /title: 'menu\.dnsRecord'/);
    assert.match(block, /permission: 'website_cert_view'/);
});

test('Certificates page no longer carries the Manage DNS button', () => {
    const src = read('../src/views/website/ssl/index.vue');
    assert.ok(!src.includes('openDnsRecord'));
    assert.ok(!src.includes('dnsRecordManage'));
});

test('every language defines the Manage DNS strings', () => {
    const dir = new URL('../src/lang/modules/', import.meta.url);
    for (const file of readdirSync(dir).filter((f) => f.endsWith('.ts'))) {
        const src = readFileSync(new URL(file, dir), 'utf8');
        assert.match(src, /^ {8}dnsRecord: '/m, `${file}: menu.dnsRecord`);
        for (const key of DNS_KEYS) {
            assert.match(src, new RegExp(`^ {8}${key}: '`, 'm'), `${file}: website.${key}`);
        }
    }
});
