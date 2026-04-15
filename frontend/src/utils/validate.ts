import i18n from '@/lang';

export function checkIp(value: string): boolean {
    if (value === '') {
        return true;
    }
    const reg =
        /^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])$/;
    return !reg.test(value) && value !== '';
}

export function checkDomain(value: string): boolean {
    if (value === '') {
        return true;
    }
    const reg = /^(?=^.{3,255}$)[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+$/;
    return !reg.test(value) && value !== '';
}

export function isDomain(value: string): boolean {
    if (value === '') {
        return true;
    }
    const reg = /^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$/i;
    return value !== '' && reg.test(value);
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
    return !IPv4AddressRegExp.test(value) && !IPv6AddressRegExp.test(value) && value !== '';
}

export function checkIpV6(value: string): boolean {
    if (value === '' || typeof value === 'undefined' || value == null) {
        return true;
    }
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
    return !IPv6AddressRegExp.test(value) && value !== '';
}

export function checkCidr(value: string): boolean {
    if (value === '') {
        return true;
    }
    const reg =
        /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(?:\/([0-9]|[1-2][0-9]|3[0-2]))?$/;
    return !reg.test(value) && value !== '';
}

export function checkCidrV6(value: string): boolean {
    if (value === '') {
        return true;
    }
    if (checkIpV6(value.split('/')[0])) {
        return true;
    }
    const reg = /^(?:[0-9]|[1-9][0-9]|1[0-1][0-9]|12[0-8])$/;
    return !reg.test(value.split('/')[1]);
}

export function checkPort(value: string): boolean {
    if (Number(value) <= 0) {
        return true;
    }
    const reg = /^([1-9](\d{0,3}))$|^([1-5]\d{4})$|^(6[0-4]\d{3})$|^(65[0-4]\d{2})$|^(655[0-2]\d)$|^(6553[0-5])$/;
    return !reg.test(value) && value !== '';
}

export function transferTimeToSecond(item: string): any {
    if (item.indexOf('s') !== -1) {
        return Number(item.replaceAll('s', ''));
    }
    if (item.indexOf('m') !== -1) {
        return Number(item.replaceAll('m', '')) * 60;
    }
    if (item.indexOf('h') !== -1) {
        return Number(item.replaceAll('h', '')) * 60 * 60;
    }
    if (item.indexOf('d') !== -1) {
        return Number(item.replaceAll('d', '')) * 60 * 60 * 24;
    }
    return Number(item);
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

export function splitTimeFromSecond(item: number): any {
    if (item < 60) return { timeItem: item, timeUnit: 's' };
    if (item < 3600) return { timeItem: item / 60, timeUnit: 'm' };
    return { timeItem: item / 3600, timeUnit: 'h' };
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

export function emptyLineFilter(str: string, spilt: string) {
    let list = str.split(spilt);
    let results = [];
    for (let i = 0; i < list.length; i++) {
        if (list[i].trim() !== '') {
            results.push(list[i]);
        }
    }
    return results.join(spilt);
}
