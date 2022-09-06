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

export function getRandomStr(e: number): string {
    const t = 'ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678';
    const a = t.length;
    let n = '';

    for (let i = 0; i < e; i++) {
        n += t.charAt(Math.floor(Math.random() * a));
    }
    return n;
}
