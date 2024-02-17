import { AcmeAccountTypes, DNSTypes, KeyTypes } from '@/global/mimetype';
import i18n from '@/lang';
import useClipboard from 'vue-clipboard3';
const { toClipboard } = useClipboard();
import { MsgError, MsgSuccess } from '@/utils/message';

export function deepCopy<T>(obj: any): T {
    let newObj: any;
    try {
        newObj = obj.push ? [] : {};
    } catch (error) {
        newObj = {};
    }
    for (let attr in obj) {
        if (typeof obj[attr] === 'object') {
            newObj[attr] = deepCopy(obj[attr]);
        } else {
            newObj[attr] = obj[attr];
        }
    }
    return newObj;
}
export function randomNum(min: number, max: number): number {
    let num = Math.floor(Math.random() * (min - max) + max);
    return num;
}

export function getBrowserLang() {
    let browserLang = navigator.language ? navigator.language : navigator.browserLanguage;
    let defaultBrowserLang = '';
    if (
        browserLang.toLowerCase() === 'cn' ||
        browserLang.toLowerCase() === 'zh' ||
        browserLang.toLowerCase() === 'zh-cn'
    ) {
        defaultBrowserLang = 'zh';
    } else {
        defaultBrowserLang = 'en';
    }
    return defaultBrowserLang;
}
export function dateFormat(row: any, col: any, dataStr: any) {
    const date = new Date(dataStr);
    const y = date.getFullYear();
    let m: string | number = date.getMonth() + 1;
    m = m < 10 ? `0${String(m)}` : m;
    let d: string | number = date.getDate();
    d = d < 10 ? `0${String(d)}` : d;
    let h: string | number = date.getHours();
    h = h < 10 ? `0${String(h)}` : h;
    let minute: string | number = date.getMinutes();
    minute = minute < 10 ? `0${String(minute)}` : minute;
    let second: string | number = date.getSeconds();
    second = second < 10 ? `0${String(second)}` : second;
    return `${String(y)}-${String(m)}-${String(d)}   ${String(h)}:${String(minute)}:${String(second)}`;
}

//2016-01-12
export function dateFormatSimple(dataStr: any) {
    const date = new Date(dataStr);
    const y = date.getFullYear();
    let m: string | number = date.getMonth() + 1;
    m = m < 10 ? `0${String(m)}` : m;
    let d: string | number = date.getDate();
    d = d < 10 ? `0${String(d)}` : d;
    return `${String(y)}-${String(m)}-${String(d)}`;
}

// 20221013151302
export function dateFormatForName(dataStr: any) {
    const date = new Date(dataStr);
    const y = date.getFullYear();
    let m: string | number = date.getMonth() + 1;
    m = m < 10 ? `0${String(m)}` : m;
    let d: string | number = date.getDate();
    d = d < 10 ? `0${String(d)}` : d;
    let h: string | number = date.getHours();
    h = h < 10 ? `0${String(h)}` : h;
    let minute: string | number = date.getMinutes();
    minute = minute < 10 ? `0${String(minute)}` : minute;
    let second: string | number = date.getSeconds();
    second = second < 10 ? `0${String(second)}` : second;
    return `${String(y)}${String(m)}${String(d)}${String(h)}${String(minute)}${String(second)}`;
}

// 10-13 \n 15:13
export function dateFormatWithoutYear(dataStr: any) {
    const date = new Date(dataStr);
    let m: string | number = date.getMonth() + 1;
    m = m < 10 ? `0${String(m)}` : m;
    let d: string | number = date.getDate();
    d = d < 10 ? `0${String(d)}` : d;
    let h: string | number = date.getHours();
    h = h < 10 ? `0${String(h)}` : h;
    let minute: string | number = date.getMinutes();
    minute = minute < 10 ? `0${String(minute)}` : minute;
    return `${String(m)}-${String(d)}\n${String(h)}:${String(minute)}`;
}

// 20221013151302
export function dateFormatForSecond(dataStr: any) {
    const date = new Date(dataStr);
    let h: string | number = date.getHours();
    h = h < 10 ? `0${String(h)}` : h;
    let minute: string | number = date.getMinutes();
    minute = minute < 10 ? `0${String(minute)}` : minute;
    let second: string | number = date.getSeconds();
    second = second < 10 ? `0${String(second)}` : second;
    return `${String(h)}:${String(minute)}:${String(second)}`;
}

export function getRandomStr(e: number): string {
    const t = 'ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678';
    const a = t.length;
    let n = '';

    for (let i = 0; i < e; i++) {
        n += t.charAt(Math.floor(Math.random() * a));
    }
    return n;
}

export function getDBName(e: number): string {
    const t = 'abcdefhijkmnprstwxyz2345678';
    const a = t.length;
    let n = '';

    for (let i = 0; i < e; i++) {
        n += t.charAt(Math.floor(Math.random() * a));
    }
    return n;
}

export function loadZero(i: number) {
    return i < 10 ? '0' + i : '' + i;
}

export function computeSize(size: number): string {
    const num = 1024.0;
    if (size < num) return size + ' B';
    if (size < Math.pow(num, 2)) return formattedNumber((size / num).toFixed(2)) + ' KB';
    if (size < Math.pow(num, 3)) return formattedNumber((size / Math.pow(num, 2)).toFixed(2)) + ' MB';
    if (size < Math.pow(num, 4)) return formattedNumber((size / Math.pow(num, 3)).toFixed(2)) + ' GB';
    return formattedNumber((size / Math.pow(num, 4)).toFixed(2)) + ' TB';
}

export function splitSize(size: number): any {
    const num = 1024.0;
    if (size < num) return { size: Number(size), unit: 'B' };
    if (size < Math.pow(num, 2)) return { size: formattedNumber((size / num).toFixed(2)), unit: 'KB' };
    if (size < Math.pow(num, 3))
        return { size: formattedNumber((size / Number(Math.pow(num, 2).toFixed(2))).toFixed(2)), unit: 'MB' };
    if (size < Math.pow(num, 4))
        return { size: formattedNumber((size / Number(Math.pow(num, 3))).toFixed(2)), unit: 'GB' };
    return { size: formattedNumber((size / Number(Math.pow(num, 4))).toFixed(2)), unit: 'TB' };
}

export function formattedNumber(num: string) {
    return num.endsWith('.00') ? Number(num.slice(0, -3)) : Number(num);
}

export function computeSizeFromMB(size: number): string {
    const num = 1024.0;
    if (size < num) return size + ' MB';
    if (size < Math.pow(num, 2)) return (size / num).toFixed(2) + ' GB';
    return (size / Math.pow(num, 3)).toFixed(2) + ' TB';
}

export function computeSizeFromKB(size: number): string {
    const num = 1024.0;
    if (size < num) return size + ' KB';
    if (size < Math.pow(num, 2)) return (size / num).toFixed(2) + ' MB';
    if (size < Math.pow(num, 3)) return (size / Math.pow(num, 2)).toFixed(2) + ' GB';
    return (size / Math.pow(num, 3)).toFixed(2) + ' TB';
}

export function computeSizeFromKBs(size: number): string {
    const num = 1024.0;
    if (size < num) return size + ' KB/s';
    if (size < Math.pow(num, 2)) return (size / num).toFixed(2) + ' MB/s';
    if (size < Math.pow(num, 3)) return (size / Math.pow(num, 2)).toFixed(2) + ' GB/s';
    return (size / Math.pow(num, 3)).toFixed(2) + ' TB/s';
}

let icons = new Map([
    ['.zip', 'p-file-zip'],
    ['.gz', 'p-file-zip'],
    ['.tar.bz2', 'p-file-zip'],
    ['.tar', 'p-file-zip'],
    ['.tar.gz', 'p-file-zip'],
    ['.tar.xz', 'p-file-zip'],
    ['.mp3', 'p-file-mp3'],
    ['.svg', 'p-file-svg'],
    ['.txt', 'p-file-txt'],
    ['.html', 'p-file-html'],
    ['.word', 'p-file-word'],
    ['.ppt', 'p-file-ppt'],
    ['.jpg', 'p-file-jpg'],
    ['.xlsx', 'p-file-excel'],
    ['.doc', 'p-file-word'],
    ['.pdf', 'p-file-pdf'],
]);

export function getIcon(extension: string): string {
    if (icons.get(extension) != undefined) {
        const icon = icons.get(extension);
        return String(icon);
    } else {
        return 'p-file-normal';
    }
}

export function checkIp(value: string): boolean {
    if (value === '') {
        return true;
    }
    const reg =
        /^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])$/;
    if (!reg.test(value) && value !== '') {
        return true;
    } else {
        return false;
    }
}

export function checkDomain(value: string): boolean {
    if (value === '') {
        return true;
    }
    const reg = /^(?=^.{3,255}$)[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+$/;
    if (!reg.test(value) && value !== '') {
        return true;
    } else {
        return false;
    }
}

export function isDomain(value: string): boolean {
    if (value === '') {
        return true;
    }
    const reg = /^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$/i;
    if (value !== '' && reg.test(value)) {
        return true;
    } else {
        return false;
    }
}

export function checkIpV4V6(value: string): boolean {
    if (value === '') {
        return true;
    }
    const IPv4SegmentFormat = '(?:[0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])';
    const IPv4AddressFormat = `(${IPv4SegmentFormat}[.]){3}${IPv4SegmentFormat}`;
    const IPv4AddressRegExp = new RegExp(`^${IPv4AddressFormat}$`);
    const IPv6SegmentFormat = '(?:[0-9a-fA-F]{1,4})';
    const IPv6AddressRegExp = new RegExp(
        '^(' +
            `(?:${IPv6SegmentFormat}:){7}(?:${IPv6SegmentFormat}|:)|` +
            `(?:${IPv6SegmentFormat}:){6}(?:${IPv4AddressFormat}|:${IPv6SegmentFormat}|:)|` +
            `(?:${IPv6SegmentFormat}:){5}(?::${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,2}|:)|` +
            `(?:${IPv6SegmentFormat}:){4}(?:(:${IPv6SegmentFormat}){0,1}:${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,3}|:)|` +
            `(?:${IPv6SegmentFormat}:){3}(?:(:${IPv6SegmentFormat}){0,2}:${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,4}|:)|` +
            `(?:${IPv6SegmentFormat}:){2}(?:(:${IPv6SegmentFormat}){0,3}:${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,5}|:)|` +
            `(?:${IPv6SegmentFormat}:){1}(?:(:${IPv6SegmentFormat}){0,4}:${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,6}|:)|` +
            `(?::((?::${IPv6SegmentFormat}){0,5}:${IPv4AddressFormat}|(?::${IPv6SegmentFormat}){1,7}|:))` +
            ')(%[0-9a-zA-Z-.:]{1,})?$',
    );
    if (!IPv4AddressRegExp.test(value) && !IPv6AddressRegExp.test(value) && value !== '') {
        return true;
    } else {
        return false;
    }
}

export function checkIpV6(value: string): boolean {
    if (value === '' || typeof value === 'undefined' || value == null) {
        return true;
    } else {
        const IPv4SegmentFormat = '(?:[0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])';
        const IPv4AddressFormat = `(${IPv4SegmentFormat}[.]){3}${IPv4SegmentFormat}`;
        const IPv6SegmentFormat = '(?:[0-9a-fA-F]{1,4})';
        const IPv6AddressRegExp = new RegExp(
            '^(' +
                `(?:${IPv6SegmentFormat}:){7}(?:${IPv6SegmentFormat}|:)|` +
                `(?:${IPv6SegmentFormat}:){6}(?:${IPv4AddressFormat}|:${IPv6SegmentFormat}|:)|` +
                `(?:${IPv6SegmentFormat}:){5}(?::${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,2}|:)|` +
                `(?:${IPv6SegmentFormat}:){4}(?:(:${IPv6SegmentFormat}){0,1}:${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,3}|:)|` +
                `(?:${IPv6SegmentFormat}:){3}(?:(:${IPv6SegmentFormat}){0,2}:${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,4}|:)|` +
                `(?:${IPv6SegmentFormat}:){2}(?:(:${IPv6SegmentFormat}){0,3}:${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,5}|:)|` +
                `(?:${IPv6SegmentFormat}:){1}(?:(:${IPv6SegmentFormat}){0,4}:${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,6}|:)|` +
                `(?::((?::${IPv6SegmentFormat}){0,5}:${IPv4AddressFormat}|(?::${IPv6SegmentFormat}){1,7}|:))` +
                ')(%[0-9a-zA-Z-.:]{1,})?$',
        );
        if (!IPv6AddressRegExp.test(value) && value !== '') {
            return true;
        } else {
            return false;
        }
    }
}

export function checkCidr(value: string): boolean {
    if (value === '') {
        return true;
    }
    const reg =
        /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(?:\/([0-9]|[1-2][0-9]|3[0-2]))?$/;
    if (!reg.test(value) && value !== '') {
        return true;
    } else {
        return false;
    }
}

export function checkPort(value: string): boolean {
    if (Number(value) <= 0) {
        return true;
    }
    const reg = /^([1-9](\d{0,3}))$|^([1-5]\d{4})$|^(6[0-4]\d{3})$|^(65[0-4]\d{2})$|^(655[0-2]\d)$|^(6553[0-5])$/;
    if (!reg.test(value) && value !== '') {
        return true;
    } else {
        return false;
    }
}

export function getProvider(provider: string): string {
    switch (provider) {
        case 'dnsAccount':
            return i18n.global.t('website.dnsAccount');
        case 'dnsManual':
            return i18n.global.t('website.dnsManual');
        case 'http':
            return 'HTTP';
        case 'selfSigned':
            return i18n.global.t('ssl.selfSigned');
        default:
            return i18n.global.t('ssl.manualCreate');
    }
}

export function splitTime(item: string): any {
    if (item.indexOf('s') !== -1) {
        return { time: Number(item.replaceAll('s', '')), unit: 's' };
    }
    if (item.indexOf('m') !== -1) {
        return { time: Number(item.replaceAll('m', '')), unit: 'm' };
    }
    if (item.indexOf('h') !== -1) {
        return { time: Number(item.replaceAll('h', '')), unit: 'h' };
    }
    if (item.indexOf('d') !== -1) {
        return { time: Number(item.replaceAll('d', '')), unit: 'd' };
    }
    if (item.indexOf('y') !== -1) {
        return { time: Number(item.replaceAll('y', '')), unit: 'y' };
    }
    return { time: Number(item), unit: 's' };
}
export function transTimeUnit(val: string): any {
    if (val.indexOf('s') !== -1) {
        return val.replaceAll('s', i18n.global.t('commons.units.second'));
    }
    if (val.indexOf('m') !== -1) {
        return val.replaceAll('m', i18n.global.t('commons.units.minute'));
    }
    if (val.indexOf('h') !== -1) {
        return val.replaceAll('h', i18n.global.t('commons.units.hour'));
    }
    if (val.indexOf('d') !== -1) {
        return val.replaceAll('d', i18n.global.t('commons.units.day'));
    }
    if (val.indexOf('y') !== -1) {
        return val.replaceAll('y', i18n.global.t('commons.units.year'));
    }
    return val + i18n.global.t('commons.units.second');
}

export function splitHttp(url: string) {
    if (url.indexOf('https://') != -1) {
        return { proto: 'https', url: url.replaceAll('https://', '') };
    }
    if (url.indexOf('http://') != -1) {
        return { proto: 'http', url: url.replaceAll('http://', '') };
    }
    return { proto: '', url: url };
}
export function spliceHttp(proto: string, url: string) {
    return proto + '://' + url.replaceAll('https://', '').replaceAll('http://', '');
}

export function getAge(d1: string): string {
    const dateBegin = new Date(d1);
    const dateEnd = new Date();
    const dateDiff = dateEnd.getTime() - dateBegin.getTime();
    const dayDiff = Math.floor(dateDiff / (24 * 3600 * 1000));
    const leave1 = dateDiff % (24 * 3600 * 1000);
    const hours = Math.floor(leave1 / (3600 * 1000));
    const leave2 = leave1 % (3600 * 1000);
    const minutes = Math.floor(leave2 / (60 * 1000));

    let res = '';
    if (dayDiff > 0) {
        res += String(dayDiff) + i18n.global.t('commons.units.day');
        if (hours <= 0) {
            return res;
        }
    }
    if (hours > 0) {
        res += String(hours) + i18n.global.t('commons.units.hour');
        return res;
    }
    if (minutes > 0) {
        res += String(minutes) + i18n.global.t('commons.units.minute');
        return res;
    }
    return i18n.global.t('app.less1Minute');
}

export function isJson(str: string) {
    try {
        if (typeof JSON.parse(str) === 'object') {
            return true;
        }
    } catch {
        return false;
    }
}

export function toLowerCase(str: string) {
    return str.toLowerCase();
}

export function downloadFile(filePath: string) {
    let url = `${import.meta.env.VITE_API_URL as string}/files/download?`;
    let path = encodeURIComponent(filePath);
    window.open(url + 'path=' + path, '_blank');
}

export function downloadWithContent(content: string, fileName: string) {
    const downloadUrl = window.URL.createObjectURL(new Blob([content]));
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = downloadUrl;
    a.download = fileName;
    const event = new MouseEvent('click');
    a.dispatchEvent(event);
}
export function getDateStr() {
    let now: Date = new Date();

    let year: number = now.getFullYear();
    let month: number = now.getMonth() + 1;
    let date: number = now.getDate();
    let hours: number = now.getHours();
    let minutes: number = now.getMinutes();
    let seconds: number = now.getSeconds();

    let timestamp: string = `${year}-${month < 10 ? '0' + month : month}-${date < 10 ? '0' + date : date}-${
        hours < 10 ? '0' + hours : hours
    }-${minutes < 10 ? '0' + minutes : minutes}-${seconds < 10 ? '0' + seconds : seconds}`;

    return timestamp;
}

export function getAccountName(type: string) {
    for (const i of AcmeAccountTypes) {
        if (i.value === type) {
            return i.label;
        }
    }
    return '';
}

export function getKeyName(type: string) {
    for (const i of KeyTypes) {
        if (i.value === type) {
            return i.label;
        }
    }
    return '';
}

export function getDNSName(type: string) {
    for (const i of DNSTypes) {
        if (i.value === type) {
            return i.label;
        }
    }
    return '';
}

export async function copyText(content: string) {
    try {
        await toClipboard(content);
        MsgSuccess(i18n.global.t('commons.msg.copySuccess'));
    } catch (e) {
        MsgError(i18n.global.t('commons.msg.copyFailed'));
    }
}
