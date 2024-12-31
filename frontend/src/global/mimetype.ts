import { CompressType } from '@/enums/files';
import i18n from '@/lang';

export const Mimetypes = new Map([
    ['application/zip', CompressType.Zip],
    ['application/x-zip', CompressType.Zip],
    ['application/x-zip-compressed', CompressType.Zip],
    ['application/x-tar', CompressType.Tar],
    ['application/x-bzip2', CompressType.Bz2],
    ['application/gzip', CompressType.TarGz],
    ['application/x-gzip', CompressType.TarGz],
    ['application/x-gunzip', CompressType.TarGz],
    ['application/gzipped', CompressType.TarGz],
    ['application/gzip-compressed', CompressType.TarGz],
    ['application/x-gzip-compressed', CompressType.TarGz],
    ['gzip/document', CompressType.TarGz],
    ['application/x-xz', CompressType.Xz],
    ['application/octet-stream', CompressType.Tar],
]);

export const Languages = [
    {
        label: 'plaintext',
        value: ['txt'],
    },
    {
        label: 'json',
        value: ['json'],
    },
    {
        label: 'vue',
        value: ['vue'],
    },
    {
        label: 'typescript',
        value: ['ts'],
    },
    {
        label: 'lua',
        value: ['lua'],
    },
    {
        label: 'markdown',
        value: ['md'],
    },
    {
        label: 'yaml',
        value: ['yml', 'yaml'],
    },
    {
        label: 'xml',
        value: ['xml'],
    },
    {
        label: 'php',
        value: ['php'],
    },
    {
        label: 'sql',
        value: ['sql'],
    },
    {
        label: 'go',
        value: ['go'],
    },
    {
        label: 'html',
        value: ['html'],
    },
    {
        label: 'javascript',
        value: ['js'],
    },
    {
        label: 'java',
        value: ['java'],
    },
    {
        label: 'kotlin',
        value: ['kt'],
    },
    {
        label: 'python',
        value: ['py'],
    },
    {
        label: 'redis',
        value: ['redis'],
    },
    {
        label: 'shell',
        value: ['sh'],
    },
    {
        label: 'css',
        value: ['css'],
    },
];

export const Rewrites = [
    'default',
    'wordpress',
    'wp2',
    'typecho',
    'typecho2',
    'thinkphp',
    'yii2',
    'laravel5',
    'discuz',
    'discuzx',
    'discuzx2',
    'discuzx3',
    'EduSoho',
    'EmpireCMS',
    'ShopWind',
    'crmeb',
    'dabr',
    'dbshop',
    'dedecms',
    'drupal',
    'ecshop',
    'emlog',
    'maccms',
    'mvc',
    'niushop',
    'phpcms',
    'sablog',
    'seacms',
    'shopex',
    'zblog',
];

export const Units = [
    { label: i18n.global.t('commons.units.second'), value: 's' },
    { label: i18n.global.t('commons.units.minute'), value: 'm' },
    { label: i18n.global.t('commons.units.hour'), value: 'h' },
    { label: i18n.global.t('commons.units.day'), value: 'd' },
    { label: i18n.global.t('commons.units.week'), value: 'w' },
    { label: i18n.global.t('commons.units.month'), value: 'M' },
    { label: i18n.global.t('commons.units.year'), value: 'y' },
];

export const AcmeAccountTypes = [
    { label: "Let's Encrypt", value: 'letsencrypt' },
    { label: 'ZeroSSL', value: 'zerossl' },
    { label: 'Buypass', value: 'buypass' },
    { label: 'Google Cloud', value: 'google' },
];

export const KeyTypes = [
    { label: 'EC 256', value: 'P256' },
    { label: 'EC 384', value: 'P384' },
    { label: 'RSA 2048', value: '2048' },
    { label: 'RSA 3072', value: '3072' },
    { label: 'RSA 4096', value: '4096' },
];

export const DNSTypes = [
    {
        label: i18n.global.t('website.aliyun'),
        value: 'AliYun',
    },
    {
        label: i18n.global.t('website.tencentCloud'),
        value: 'TencentCloud',
    },
    {
        label: i18n.global.t('website.huaweicloud'),
        value: 'HuaweiCloud',
    },
    {
        label: i18n.global.t('website.volcengine'),
        value: 'Volcengine',
    },
    {
        label: 'DNSPod (' + i18n.global.t('ssl.deprecated') + ')',
        value: 'DnsPod',
    },
    {
        label: 'Cloudflare',
        value: 'CloudFlare',
    },
    {
        label: 'CloudDNS',
        value: 'CloudDns',
    },
    {
        label: 'NameSilo',
        value: 'NameSilo',
    },
    {
        label: 'NameCheap',
        value: 'NameCheap',
    },
    {
        label: 'Name.com',
        value: 'NameCom',
    },
    {
        label: 'GoDaddy',
        value: 'Godaddy',
    },
    {
        label: i18n.global.t('website.rainyun'),
        value: 'RainYun',
    },
];

export const Fields = [
    {
        label: 'URL',
        value: 'URL',
    },
    {
        label: 'IP',
        value: 'IP',
    },
    {
        label: 'Header',
        value: 'Header',
    },
    {
        label: 'Host',
        value: 'Host',
    },
];

export const Patterns = [
    {
        label: i18n.global.t('xpack.waf.contain'),
        value: 'contain',
    },
    {
        label: i18n.global.t('xpack.waf.equal'),
        value: 'eq',
    },
    {
        label: i18n.global.t('xpack.waf.regex'),
        value: 'regex',
    },
    {
        label: i18n.global.t('xpack.waf.notEqual'),
        value: 'notEq',
    },
];

export const HttpCodes = [
    {
        label: i18n.global.t('xpack.waf.badReq'),
        value: 400,
    },
    {
        label: i18n.global.t('xpack.waf.forbidden'),
        value: 403,
    },
    {
        label: i18n.global.t('xpack.waf.noRes'),
        value: 444,
    },
    {
        label: i18n.global.t('xpack.waf.serverErr'),
        value: 500,
    },
];

export const Actions = [
    {
        label: i18n.global.t('xpack.waf.actionAllow'),
        value: 'allow',
    },
    {
        label: i18n.global.t('xpack.waf.deny'),
        value: 'deny',
    },
    {
        label: i18n.global.t('xpack.waf.captcha'),
        value: 'captcha',
    },
    {
        label: i18n.global.t('xpack.waf.fiveSeconds'),
        value: 'five_seconds',
    },
];
