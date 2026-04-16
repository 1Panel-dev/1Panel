import { Base64 } from 'js-base64';

type StringLikeRecord = object;

export const encodeBase64 = (value: string) => Base64.encode(value);

export const decodeBase64 = (value: string) => Base64.decode(value);

export const encodeBase64Fields = <T extends StringLikeRecord>(target: T, fields: Array<Extract<keyof T, string>>) => {
    fields.forEach((field) => {
        const record = target as Record<string, unknown>;
        const value = record[field];
        if (typeof value === 'string' && value) {
            record[field] = encodeBase64(value);
        }
    });

    return target;
};
