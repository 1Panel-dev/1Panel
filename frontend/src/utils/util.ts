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
export function dateFromat(row: number, col: number, dataStr: any) {
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

// 20221013151302
export function dateFromatForName(dataStr: any) {
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
export function dateFromatWithoutYear(dataStr: any) {
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
export function dateFromatForSecond(dataStr: any) {
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

let icons = new Map([
    ['.zip', 'p-file-zip'],
    ['.gz', 'p-file-zip'],
    ['.tar.bz2', 'p-file-zip'],
    ['.tar', 'p-file-zip'],
    ['.tar.gz', 'p-file-zip'],
    ['.tar.xz', 'p-file-zip'],
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
    const reg =
        /^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])$/;
    if (!reg.test(value) && value !== '') {
        return true;
    } else {
        return false;
    }
}
