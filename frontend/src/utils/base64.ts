import { Base64 } from 'js-base64';

type StringLikeRecord = Record<string, unknown>;

export const encodeBase64 = (value: string) => Base64.encode(value);

export const decodeBase64 = (value: string) => Base64.decode(value);

export const encodeBase64Fields = <T extends StringLikeRecord>(target: T, fields: Array<keyof T>) => {
    fields.forEach((field) => {
        const value = target[field];
        if (typeof value === 'string' && value) {
            target[field] = encodeBase64(value) as T[keyof T];
        }
    });

    return target;
};
