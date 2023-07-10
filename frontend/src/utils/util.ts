import i18n from '@/lang';

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

export function loadZero(i: number) {
    return i < 10 ? '0' + i : '' + i;
}

export function computeSize(size: number): string {
    const num = 1024.0;
    if (size < num) return size + ' B';
    if (size < Math.pow(num, 2)) return (size / num).toFixed(2) + ' KB';
    if (size < Math.pow(num, 3)) return (size / Math.pow(num, 2)).toFixed(2) + ' MB';
    if (size < Math.pow(num, 4)) return (size / Math.pow(num, 3)).toFixed(2) + ' GB';
    return (size / Math.pow(num, 4)).toFixed(2) + ' TB';
}

export function computeSizeFromMB(size: number): string {
    const num = 1024.0;
    if (size < num) return size + ' MB';
    if (size < Math.pow(num, 2)) return (size / num).toFixed(2) + ' GB';
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

export function getIcon(extention: string): string {
    if (icons.get(extention) != undefined) {
        const icon = icons.get(extention);
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
        default:
            return i18n.global.t('ssl.manualCreate');
    }
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
    window.open(url + 'path=' + filePath, '_blank');
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
