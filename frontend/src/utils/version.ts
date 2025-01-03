export function compareVersion(version1: string, version2: string): boolean {
    const v1s = extractNumbers(version1);
    const v2s = extractNumbers(version2);

    const maxLen = Math.max(v1s.length, v2s.length);
    v1s.push(...new Array(maxLen - v1s.length).fill('0'));
    v2s.push(...new Array(maxLen - v2s.length).fill('0'));

    for (let i = 0; i < maxLen; i++) {
        const v1 = parseInt(v1s[i], 10);
        const v2 = parseInt(v2s[i], 10);
        if (v1 !== v2) {
            return v1 > v2;
        }
    }
    return true;
}

function extractNumbers(version: string): string[] {
    const numbers: string[] = [];
    let start = -1;
    for (let i = 0; i < version.length; i++) {
        if (isDigit(version[i])) {
            if (start === -1) {
                start = i;
            }
        } else {
            if (start !== -1) {
                numbers.push(version.slice(start, i));
                start = -1;
            }
        }
    }
    if (start !== -1) {
        numbers.push(version.slice(start));
    }
    return numbers;
}

function isDigit(char: string): boolean {
    return /^\d$/.test(char);
}
