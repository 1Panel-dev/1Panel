import { v4 as uuidv4 } from 'uuid';

export function getRandomStr(length: number): string {
    const chars = 'ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678';
    let result = '';
    for (let i = 0; i < length; i++) {
        result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
}

export const newUUID = () => {
    return uuidv4();
};
